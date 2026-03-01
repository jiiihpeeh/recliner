package main

import (
	"github.com/jiiihpeeh/recliner/app"
	"github.com/jiiihpeeh/recliner/c"
	"github.com/jiiihpeeh/recliner/vdom"
)

func main() {
	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		return c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleRound,
			BorderColor: "blue",
			Padding:     1,
			Style: c.StyleProps{
				Display:        c.DisplayFlex,
				FlexDirection:  c.FlexDirectionColumn,
				JustifyContent: c.JustifyContentSpaceEvenly,
				AlignItems:     c.AlignItemsCenter,
				Width:          60,
				Height:         15,
				Background:     "#1a1a1a",
			},
		},
			c.Text(c.TextProps{
				Content: " Z-Index & Overlap Test ",
				Style:   c.TextStyle().Bold().BG("magenta").Color("white"),
			}),
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleSingle,
				Style: c.StyleProps{
					Width:      20,
					Height:     5,
					Background: "red",
					ZIndex:     1,
				},
			}, c.Text(c.TextProps{Content: "Base (Z:1)"})),
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleDouble,
				Style: c.StyleProps{
					Position:   c.PositionAbsolute,
					Top:        4,
					Left:       25,
					Width:      20,
					Height:     5,
					Background: "green",
					ZIndex:     10,
				},
			}, c.Text(c.TextProps{Content: "Overlay (Z:10)"})),
			c.Box(c.BoxProps{
				Style: c.StyleProps{
					Position:   c.PositionAbsolute,
					Top:        6,
					Left:       15,
					Width:      20,
					Height:     5,
					Background: "blue",
					ZIndex:     5,
				},
			}, c.Text(c.TextProps{Content: "Middle (Z:5)"})),
		)
	}, app.AppOptions{})

	if err := appInstance.Run(); err != nil {
		panic(err)
	}
}
