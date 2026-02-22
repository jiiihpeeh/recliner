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
			BorderStyle: c.BorderStyleSingle,
			BorderColor: "white",
			Style:       c.StyleProps{Width: 60, Height: 15, Background: "black", Display: c.DisplayFlex, FlexDirection: c.FlexDirectionColumn, Gap: 1},
		},
			// Row 1: Justify Content Space Between
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleNone,
				Style:       c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionRow, JustifyContent: c.JustifyContentSpaceBetween, Background: "blue"},
			},
				c.Text(c.TextProps{Content: "Left"}),
				c.Text(c.TextProps{Content: "Center"}),
				c.Text(c.TextProps{Content: "Right"}),
			),

			// Row 2: Flex Grow
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleSingle,
				Style:       c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionRow, Gap: 2},
			},
				c.Box(c.BoxProps{BorderStyle: c.BorderStyleSingle, Style: c.StyleProps{Background: "red"}}, c.Text(c.TextProps{Content: "Fixed"})),
				c.Box(c.BoxProps{BorderStyle: c.BorderStyleSingle, Style: c.StyleProps{FlexGrow: 1, Background: "green"}}, c.Text(c.TextProps{Content: "Grow 1"})),
				c.Box(c.BoxProps{BorderStyle: c.BorderStyleSingle, Style: c.StyleProps{FlexGrow: 2, Background: "yellow"}}, c.Text(c.TextProps{Content: "Grow 2"})),
			),

			// Row 3: Align Self
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleSingle,
				Style:       c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionRow, Height: 5, AlignItems: c.AlignItemsFlexStart, Gap: 3},
			},
				c.Text(c.TextProps{Content: "Start"}),
				c.Box(c.BoxProps{BorderStyle: c.BorderStyleSingle, Style: c.StyleProps{AlignSelf: c.AlignSelfCenter}}, c.Text(c.TextProps{Content: "Self Center"})),
				c.Box(c.BoxProps{BorderStyle: c.BorderStyleSingle, Style: c.StyleProps{AlignSelf: c.AlignSelfFlexEnd}}, c.Text(c.TextProps{Content: "Self End"})),
			),
		)
	}, app.AppOptions{Headless: true})

	output, err := appInstance.RunHeadless()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Flexbox Enhancements Demo Output:")
	fmt.Println(output)
}
