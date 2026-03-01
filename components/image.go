package components

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
	"github.com/nfnt/resize"
)

var (
	imageCache   = make(map[string]image.Image)
	imageCacheMu sync.RWMutex
	imageClient  = &http.Client{Timeout: 10 * time.Second}
)

// Image renders an image from a local path or HTTP URL. Loading is cancellable
// via the effect cleanup (context cancellation). A simple in-memory cache is used.
func Image(props any) vdom.Node {
	hc := hooks.GetContext()

	src, _ := util.GetProp[string](props, "src")
	width, _ := util.GetProp[int](props, "width")
	height, _ := util.GetProp[int](props, "height")
	mode, _ := util.GetProp[string](props, "mode")

	if mode == "" || mode == "auto" {
		caps := util.GetTermCapabilities()
		if caps.TrueColor {
			mode = "halfblock"
		} else {
			mode = "halfblock"
		}
	}

	img, setImg := hooks.UseState[image.Image](hc, nil)
	loading, setLoading := hooks.UseState[bool](hc, false)
	loadErr, setErr := hooks.UseState[error](hc, nil)

	hc.UseEffect(func() func() {
		if src == "" {
			return func() {}
		}

		// Return cached image if present
		imageCacheMu.RLock()
		if cached, ok := imageCache[src]; ok {
			setImg(cached)
			imageCacheMu.RUnlock()
			return func() {}
		}
		imageCacheMu.RUnlock()

		ctx, cancel := context.WithCancel(context.Background())
		setLoading(true)
		setErr(nil)

		go func() {
			defer setLoading(false)

			isCancelled := func() bool {
				select {
				case <-ctx.Done():
					return true
				default:
					return false
				}
			}

			var reader image.Image
			var decodeErr error

			var buf bytes.Buffer

			if strings.HasPrefix(src, "http") {
				req, err := http.NewRequestWithContext(ctx, "GET", src, nil)
				if err != nil {
					if !isCancelled() {
						setErr(err)
					}
					return
				}
				resp, err := imageClient.Do(req)
				if err != nil {
					if !isCancelled() {
						setErr(err)
					}
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					if !isCancelled() {
						setErr(fmt.Errorf("http status: %s", resp.Status))
					}
					return
				}

				tmp := make([]byte, 32*1024)
				for {
					if isCancelled() {
						return
					}
					n, rerr := resp.Body.Read(tmp)
					if n > 0 {
						if _, werr := buf.Write(tmp[:n]); werr != nil {
							if !isCancelled() {
								setErr(werr)
							}
							return
						}
					}
					if rerr != nil {
						if rerr == io.EOF {
							break
						}
						if !isCancelled() {
							setErr(rerr)
						}
						return
					}
				}

				reader, _, decodeErr = image.Decode(bytes.NewReader(buf.Bytes()))
			} else {
				f, err := os.Open(src)
				if err != nil {
					if !isCancelled() {
						setErr(err)
					}
					return
				}
				defer f.Close()

				br := bufio.NewReader(f)
				tmp := make([]byte, 32*1024)
				for {
					if isCancelled() {
						return
					}
					n, rerr := br.Read(tmp)
					if n > 0 {
						if _, werr := buf.Write(tmp[:n]); werr != nil {
							if !isCancelled() {
								setErr(werr)
							}
							return
						}
					}
					if rerr != nil {
						if rerr == io.EOF {
							break
						}
						if !isCancelled() {
							setErr(rerr)
						}
						return
					}
				}

				reader, _, decodeErr = image.Decode(bytes.NewReader(buf.Bytes()))
			}

			if isCancelled() {
				return
			}

			if decodeErr != nil {
				setErr(decodeErr)
				return
			}

			// store in cache and set state
			imageCacheMu.Lock()
			imageCache[src] = reader
			imageCacheMu.Unlock()
			setImg(reader)
		}()

		return func() {
			cancel()
		}
	}, []any{src})

	if loading {
		return &vdom.Element{Type: "text", InnerText: " ⌛ Loading image... ", Style: vdom.Style{Foreground: "gray"}}
	}
	if loadErr != nil {
		return &vdom.Element{Type: "text", InnerText: " ❌ Error: " + loadErr.Error(), Style: vdom.Style{Foreground: "red"}}
	}
	if img == nil {
		return &vdom.Element{Type: "text", InnerText: ""}
	}

	targetW := uint(width)
	targetH := uint(height)
	if targetW == 0 {
		targetW = 40
	}
	if targetH == 0 {
		bounds := img.Bounds()
		ratio := float64(bounds.Dy()) / float64(bounds.Dx())
		targetH = uint(float64(targetW) * ratio * 0.5)
	}

	resized := hooks.UseMemo(hc, func() image.Image {
		// We use half-block, so we need 2x vertical pixels
		return resize.Resize(targetW, targetH*2, img, resize.Lanczos3)
	}, []any{img, targetW, targetH})

	return &vdom.Element{
		Type: "image",
		Props: struct {
			Image  image.Image
			Mode   string
			Width  int
			Height int
		}{
			Image:  resized,
			Mode:   mode,
			Width:  int(targetW),
			Height: int(targetH),
		},
		Style: vdom.Style{
			Width:  int(targetW),
			Height: int(targetH),
		},
	}
}
