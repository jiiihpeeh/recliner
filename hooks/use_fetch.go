package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/j-p/recliner/utils"
)

type FetchResult[T any] struct {
	Data         T
	Loading      bool
	Error        error
	Progress     float64 // 0.0 .. 1.0
	ResponseCode int
}

func UseFetch[T any](hc *HooksContext, url string) FetchResult[T] {
	var zero T
	data, setData := UseState[T](hc, zero)
	loading, setLoading := UseState[bool](hc, true)
	errorVal, setError := UseState[error](hc, nil)
	progress, setProgress := UseState[float64](hc, 0.0)
	responseCode, setResponseCode := UseState[int](hc, 0)

	hc.UseEffect(func() func() {
		ctx, cancel := context.WithCancel(context.Background())

		setLoading(true)
		setError(nil)
		setProgress(0.0)

		go func() {
			defer cancel()
			// Use the asynchronous Fetch helper and wait for result
			ch := utils.Fetch(url, &utils.FetchOptions{Context: ctx})
			res := <-ch

			if ctx.Err() != nil {
				return
			}

			// Capture and expose HTTP response code (safe)
			if res == nil {
				setError(fmt.Errorf("no response"))
				setLoading(false)
				setProgress(0.0)
				setResponseCode(0)
				return
			}
			if res.Response != nil {
				setResponseCode(res.StatusCode)
			} else {
				setResponseCode(0)
			}

			if !res.Ok() {
				if res.Error() != nil {
					setError(res.Error())
				} else if res.Response != nil {
					setError(fmt.Errorf("http error: %d %s", res.StatusCode, res.Status))
				} else {
					setError(fmt.Errorf("http error"))
				}
				setLoading(false)
				setProgress(0.0)
				return
			}

			var result T
			// If Content-Length is provided, stream the body and report progress
			if res.Response != nil && res.Response.Body != nil && res.ContentLength > 0 {
				defer res.Response.Body.Close()
				cl := res.ContentLength
				var buf bytes.Buffer
				tmp := make([]byte, 4096)
				for {
					n, err := res.Response.Body.Read(tmp)
					if n > 0 {
						buf.Write(tmp[:n])
						setProgress(float64(buf.Len()) / float64(cl))
					}
					if err != nil {
						if err == io.EOF {
							break
						}
						if ctx.Err() != nil {
							return
						}
						setError(fmt.Errorf("read error: %w", err))
						setLoading(false)
						return
					}
				}
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					if ctx.Err() != nil {
						return
					}
					setError(fmt.Errorf("json decode error: %w", err))
					setLoading(false)
					return
				}
			} else {
				var resultBytes []byte
				var err error
				if res.Response != nil && res.Response.Body != nil {
					// No content length: read all and set progress to indeterminate (0->1 at end)
					resultBytes, err = io.ReadAll(res.Response.Body)
					res.Response.Body.Close()
				} else {
					// Fallback to FetchResponse helper
					resultBytes, err = res.Bytes()
				}
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					setError(fmt.Errorf("read error: %w", err))
					setLoading(false)
					return
				}
				if err := json.Unmarshal(resultBytes, &result); err != nil {
					if ctx.Err() != nil {
						return
					}
					setError(fmt.Errorf("json decode error: %w", err))
					setLoading(false)
					return
				}
				setProgress(1.0)
			}

			if ctx.Err() == nil {
				setData(result)
				setError(nil)
				setLoading(false)
				setProgress(1.0)
				setResponseCode(res.StatusCode)
			}
		}()

		return func() {
			cancel()
		}
	}, []any{url})

	return FetchResult[T]{
		Data:         data,
		Loading:      loading,
		Error:        errorVal,
		Progress:     progress,
		ResponseCode: responseCode,
	}
}
