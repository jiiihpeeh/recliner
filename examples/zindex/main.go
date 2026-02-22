package main

import (
	"fmt"
	"os"

	"github.com/j-p/recliner/app"
	"github.com/j-p/recliner/c"
	"github.com/j-p/recliner/vdom"
)

func main() {
	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		return c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleNone,
			Style:       c.StyleProps{Width: 40, Height: 20, Background: "black"},
		},
			// Background box (Z-Index 10, should be ON TOP)
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleSingle,
				BorderColor: "blue",
				Style:       c.StyleProps{Position: c.PositionAbsolute, Top: 2, Left: 5, Width: 20, Height: 10, Background: "blue", ZIndex: 10},
			},
				c.Text(c.TextProps{Content: "Background Layer (Z:10)"}),
			),
			// Foreground box (Z-Index 0, should be BEHIND)
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleDouble,
				BorderColor: "red",
				Style:       c.StyleProps{Position: c.PositionAbsolute, Top: 5, Left: 15, Width: 20, Height: 10, Background: "red", ZIndex: 0},
			},
				c.Text(c.TextProps{Content: "Foreground Layer (Z:0)"}),
			),
		)
	}, app.AppOptions{Headless: true})

	output, err := appInstance.RunHeadless()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Z-Index Demo Output:")
	fmt.Println(output)
}
