package main

import (
	"fmt"
	"time"

	"github.com/j-p/recliner/app"
	"github.com/j-p/recliner/c"
	"github.com/j-p/recliner/hooks"
	"github.com/j-p/recliner/vdom"
)

func main() {
	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		hc := hooks.GetContext()
		count, setCount := hooks.UseState(hc, 0)

		hooks.UseEffect(hc, func() func() {
			ticker := time.NewTicker(500 * time.Millisecond)
			go func() {
				c := 0
				for range ticker.C {
					c++
					setCount(c)
				}
			}()
			return func() { ticker.Stop() }
		}, []any{})

		return c.Box(c.BoxProps{
			Style: c.StyleProps{
				Width:          40,
				Height:         5,
				Background:     "blue",
				Display:        c.DisplayFlex,
				JustifyContent: c.JustifyContentCenter,
				AlignItems:     c.AlignItemsCenter,
			},
		},
			c.Text(c.TextProps{
				Content: fmt.Sprintf("Running Number: %d", count),
				Style:   c.TextStyle().Bold().Color("white"),
			}),
		)
	}, app.AppOptions{})

	if err := appInstance.Run(); err != nil {
		panic(err)
	}
}
