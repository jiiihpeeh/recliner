package main

import (
	"fmt"

	"github.com/jiiihpeeh/recliner/app"
	"github.com/jiiihpeeh/recliner/c"
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/vdom"
)

func main() {
	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		hc := hooks.GetContext()

		count, setCount := hooks.UseState(hc, 0)
		active, setActive := hooks.UseState(hc, true)

		bgColor := "blue"
		if !active {
			bgColor = "red"
		}

		return c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleRound,
			BorderColor: "cyan",
			Padding:     1,
			Style: c.StyleProps{
				Display:       c.DisplayFlex,
				FlexDirection: c.FlexDirectionColumn,
				AlignItems:    c.AlignItemsCenter,
				Gap:           1,
				Width:         50,
				Background:    "black",
			},
		},
			c.Text(c.TextProps{
				Content: " Interactive Update Test ",
				Style:   c.TextStyle().Bold().BG("white").Color("black"),
			}),
			c.Spacer(1),
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleSingle,
				Style: c.StyleProps{
					Width:          20,
					Height:         3,
					Background:     bgColor,
					Display:        c.DisplayFlex,
					AlignItems:     c.AlignItemsCenter,
					JustifyContent: c.JustifyContentCenter,
				},
			},
				c.Text(c.TextProps{
					Content: fmt.Sprintf("Count: %d", count),
					Style:   c.TextStyle().Bold(),
				}),
			),
			c.Spacer(1),
			c.Box(c.BoxProps{
				Style: c.StyleProps{
					Display:       c.DisplayFlex,
					FlexDirection: c.FlexDirectionRow,
					Gap:           2,
				},
			},
				c.Button(c.ButtonProps{
					Label: " +1 ",
					OnClick: func(e events.MouseEvent) {
						if e.Action == events.MouseActionPress {
							setCount(count + 1)
						}
					},
					Style: c.ButtonStyleSuccess,
				}),
				c.Button(c.ButtonProps{
					Label: " -1 ",
					OnClick: func(e events.MouseEvent) {
						if e.Action == events.MouseActionPress {
							setCount(count - 1)
						}
					},
					Style: c.ButtonStyleDanger,
				}),
			),
			c.Spacer(1),
			c.CheckBox(c.CheckBoxProps{
				Label:   c.Text(c.TextProps{Content: "Active State (Blue/Red)"}),
				Checked: active,
				OnChange: func(b bool) {
					setActive(b)
				},
			}),
			c.Spacer(1),
			c.Text(c.TextProps{
				Content: "Use Mouse or Tab/Enter to interact",
				Style:   c.TextStyle().Dim(),
			}),
		)
	}, app.AppOptions{})

	if err := appInstance.Run(); err != nil {
		panic(err)
	}
}
