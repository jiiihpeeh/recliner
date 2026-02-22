package main

import (
	"fmt"
	"github.com/j-p/recliner/app"
	"github.com/j-p/recliner/c"
	"github.com/j-p/recliner/components"
	"github.com/j-p/recliner/vdom"
)

func main() {
	components.Register()

	// Create a headless app
	a := app.NewWithOptions(func(props any) vdom.Node {
		return c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleRound,
			Style: c.StyleProps{
				Width: 40, Height: 5, Background: "blue",
				Display: c.DisplayFlex, JustifyContent: c.JustifyContentCenter, AlignItems: c.AlignItemsCenter,
			},
		},
			c.Text(c.TextProps{
				Content: "Stable Render Test",
				Style:   c.TextStyle().Bold(),
			}),
		)
	}, app.AppOptions{Headless: true})

	output, err := a.RunHeadless()
	if err != nil {
		panic(err)
	}

	fmt.Println("--- START RENDER ---")
	fmt.Print(output)
	fmt.Println("--- END RENDER ---")
}
