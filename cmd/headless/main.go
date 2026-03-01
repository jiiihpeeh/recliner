package main

import (
	"fmt"

	"github.com/jiiihpeeh/recliner/app"
	"github.com/jiiihpeeh/recliner/c"
	"github.com/jiiihpeeh/recliner/components"
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/vdom"
)

func main() {
	components.Register()

	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		// replicate createApp from main.go
		hooksCtx := hooks.GetContext()

		count, setCount := hooks.UseState[int](hooksCtx, 0)
		text, setText := hooks.UseState[string](hooksCtx, "Hello, World!")

		hooksCtx.UseInput(func(key events.KeyPressEvent) {
			switch key.Key {
			case "up":
				setCount(count + 1)
			case "down":
				setCount(count - 1)
			case "r":
				setText("Random: " + fmt.Sprint(count*17))
			}
		}, []any{count, text})

		return c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleRound,
			BorderColor: "cyan",
			Padding:     1,
		},
			c.Text(c.TextProps{
				Content: "Title",
				Style:   c.TextStyle().Bold(),
			}),
			c.Text(c.TextProps{
				Content: "Subtitle",
				Style:   nil,
			}),
			c.Box(c.BoxProps{
				BorderColor: "yellow",
				Padding:     0,
			},
				c.Text(c.TextProps{
					Content: "Nested Box",
					Style:   nil,
				}),
			),
		)
	}, app.AppOptions{Headless: true})

	out, err := appInstance.RunHeadless()
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
