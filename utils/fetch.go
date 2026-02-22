package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FetchOptions mimics the optional second argument of fetch API
type FetchOptions struct {
	Method  string
	Headers map[string]string
	Body    any
	Context context.Context
}

// FetchResponse wraps http.Response to provide convenience methods like .Json() and .Text()
type FetchResponse struct {
	*http.Response
	err error
}

type Blob struct {
	Data []byte
	Type string
}

// FetchSync provides a JS-like fetch API that is synchronous/blocking.
// Call `FetchSync` inside a goroutine to avoid freezing the UI.
func FetchSync(url string, opts *FetchOptions) *FetchResponse {
	client := &http.Client{}
	method := "GET"
	var bodyReader io.Reader
	var contentType string

	if opts != nil {
		if opts.Method != "" {
			method = opts.Method
		}
		if opts.Body != nil {
			// Automatically marshal body if it's not already reader/bytes
			switch v := opts.Body.(type) {
			case io.Reader:
				bodyReader = v
			case []byte:
				bodyReader = bytes.NewReader(v)
				if contentType == "" {
					contentType = "application/octet-stream"
				}
			case string:
				bodyReader = bytes.NewBufferString(v)
				if contentType == "" {
					contentType = "text/plain"
				}
			default:
				jsonBody, err := json.Marshal(opts.Body)
				if err != nil {
					return &FetchResponse{err: fmt.Errorf("failed to marshal body: %w", err)}
				}
				bodyReader = bytes.NewBuffer(jsonBody)
				if contentType == "" {
					contentType = "application/json"
				}
			}
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return &FetchResponse{err: fmt.Errorf("failed to create request: %w", err)}
	}

	if opts != nil {
		if opts.Context != nil {
			req = req.WithContext(opts.Context)
		}
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
		// Auto-set Content-Type if body is present and not set
		if opts.Body != nil && req.Header.Get("Content-Type") == "" && contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
	}

	resp, err := client.Do(req)
	return &FetchResponse{Response: resp, err: err}
}

// Fetch runs the request in a new goroutine and returns a channel which
// will receive the *FetchResponse once the request completes.
func Fetch(url string, opts *FetchOptions) <-chan *FetchResponse {
	ch := make(chan *FetchResponse, 1)
	go func() {
		ch <- FetchSync(url, opts)
		close(ch)
	}()
	return ch
}

// Ok returns true if the request succeeded (status 200-299) and no network error occurred
func (r *FetchResponse) Ok() bool {
	if r.err != nil || r.Response == nil {
		return false
	}
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// Json decodes the response body into the provided target v
func (r *FetchResponse) Json(v any) error {
	if r.err != nil {
		return r.err
	}
	if r.Response == nil {
		return fmt.Errorf("response is nil")
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// Text returns the response body as a string
func (r *FetchResponse) Text() (string, error) {
	b, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Bytes returns the response body as a byte slice
func (r *FetchResponse) Bytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.Response == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

// BytesWithProgress reads the response body in chunks and reports progress via the
// provided callback. `total` may be -1 if unknown. The callback receives the
// number of bytes read so far and the total content length.
func (r *FetchResponse) BytesWithProgress(progress func(read int64, total int64)) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.Response == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer r.Body.Close()

	if progress == nil {
		return io.ReadAll(r.Body)
	}

	total := r.ContentLength
	buf := bytes.NewBuffer(nil)
	tmp := make([]byte, 32*1024)
	var read int64
	for {
		n, err := r.Body.Read(tmp)
		if n > 0 {
			if _, werr := buf.Write(tmp[:n]); werr != nil {
				return nil, werr
			}
			read += int64(n)
			progress(read, total)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// JsonWithProgress decodes the response body as JSON while reporting progress
// through the provided callback. `total` may be -1 if unknown.
func (r *FetchResponse) JsonWithProgress(v any, progress func(read int64, total int64)) error {
	if r.err != nil {
		return r.err
	}
	if r.Response == nil {
		return fmt.Errorf("response is nil")
	}
	b, err := r.BytesWithProgress(progress)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// Blob returns the response body as a Blob (bytes + content type)
func (r *FetchResponse) Blob() (*Blob, error) {
	b, err := r.Bytes()
	if err != nil {
		return nil, err
	}
	return &Blob{
		Data: b,
		Type: r.Header.Get("Content-Type"),
	}, nil
}

// ArrayBuffer is an alias for Bytes to be more "JS-like"
func (r *FetchResponse) ArrayBuffer() ([]byte, error) {
	return r.Bytes()
}

// Error returns the underlying network error if any
func (r *FetchResponse) Error() error {
	return r.err
}
