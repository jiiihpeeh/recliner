package main

import (
	"github.com/j-p/recliner/app"
	"github.com/j-p/recliner/c"
	"github.com/j-p/recliner/vdom"
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
				JustifyContent: c.JustifyContentCenter,
				AlignItems:     c.AlignItemsCenter,
				Width:          60,
				Height:         15,
				Background:     "#1a1a1a",
			},
		},
			c.Text(c.TextProps{
				Content: " Layout Test: Center Alignment ",
				Style:   c.TextStyle().Bold().BG("blue").Color("white"),
			}),
			c.Spacer(1),
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleDouble,
				BorderColor: "green",
				Padding:     1,
				Style: c.StyleProps{
					Display:       c.DisplayFlex,
					FlexDirection: c.FlexDirectionRow,
					Gap:           2,
					Width:         40,
					Height:        5,
				},
			},
				c.Box(c.BoxProps{
					Style: c.StyleProps{Width: 10, Height: 3, Background: "red"},
				}, c.Text(c.TextProps{Content: "Left"})),
				c.Box(c.BoxProps{
					Style: c.StyleProps{Width: 10, Height: 3, Background: "yellow"},
				}, c.Text(c.TextProps{Content: "Mid"})),
				c.Box(c.BoxProps{
					Style: c.StyleProps{Width: 10, Height: 3, Background: "cyan"},
				}, c.Text(c.TextProps{Content: "Right"})),
			),
			c.Spacer(1),
			c.Text(c.TextProps{
				Content: "Testing Flex Column + Row Nesting",
				Style:   c.TextStyle().Dim(),
			}),
		)
	}, app.AppOptions{})

	if err := appInstance.Run(); err != nil {
		panic(err)
	}
}
