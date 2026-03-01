package main

import (
	"github.com/jiiihpeeh/recliner/app"
	"github.com/jiiihpeeh/recliner/c"
	"github.com/jiiihpeeh/recliner/vdom"
)

func main() {
	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		return c.Box(c.BoxProps{
			Style: c.StyleProps{
				Display:       c.DisplayFlex,
				FlexDirection: c.FlexDirectionColumn,
				Width:         100,
				Height:        40,
				Background:    "#0a0a0a",
			},
		},
			// Header
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleBold,
				BorderColor: "blue",
				Style: c.StyleProps{
					Width:          100,
					Height:         3,
					Display:        c.DisplayFlex,
					JustifyContent: c.JustifyContentCenter,
					AlignItems:     c.AlignItemsCenter,
					Background:     "#1e1e1e",
				},
			},
				c.Text(c.TextProps{
					Content: " Recliner Layout Test Suite ",
					Style:   c.TextStyle().Bold().Color("cyan"),
				}),
			),

			// Sidebar + Main Content Row
			c.Box(c.BoxProps{
				Style: c.StyleProps{
					Display:       c.DisplayFlex,
					FlexDirection: c.FlexDirectionRow,
					FlexGrow:      1,
				},
			},
				// Sidebar
				c.Box(c.BoxProps{
					BorderStyle: c.BorderStyleRound,
					BorderColor: "gray",
					Padding:     1,
					Style: c.StyleProps{
						Width:      20,
						Background: "#111111",
					},
				},
					c.Text(c.TextProps{Content: "Layout Options", Style: c.TextStyle().Underline()}),
					c.Spacer(1),
					c.Text(c.TextProps{Content: "- Flex Row"}),
					c.Text(c.TextProps{Content: "- Flex Column"}),
					c.Text(c.TextProps{Content: "- Margins"}),
					c.Text(c.TextProps{Content: "- Padding"}),
					c.Text(c.TextProps{Content: "- Nesting"}),
				),

				// Main Area
				c.Box(c.BoxProps{
					Style: c.StyleProps{
						FlexGrow:      1,
						Display:       c.DisplayFlex,
						FlexDirection: c.FlexDirectionColumn,
					},
				},
					// Row 1: Margins & Padding
					c.Box(c.BoxProps{
						Style: c.StyleProps{
							Display:       c.DisplayFlex,
							FlexDirection: c.FlexDirectionRow,
							Gap:           2,
						},
					},
						c.Box(c.BoxProps{
							BorderStyle: c.BorderStyleSingle,
							BorderColor: "green",
							Style: c.StyleProps{
								Width:      25,
								Height:     7,
								MarginLeft: 2,
								MarginTop:  1,
								Background: "#220000",
							},
						},
							c.Text(c.TextProps{Content: "Margin: L=2, T=1", Style: c.TextStyle().Color("red")}),
						),
						c.Box(c.BoxProps{
							BorderStyle: c.BorderStyleDouble,
							BorderColor: "yellow",
							Padding:     1,
							Style: c.StyleProps{
								Width:      25,
								Height:     7,
								Background: "#002200",
							},
						},
							c.Text(c.TextProps{Content: "Padding: 1", Style: c.TextStyle().Color("green")}),
						),
					),

					c.Spacer(1),

					// Row 2: Justify Content
					c.Box(c.BoxProps{
						BorderStyle: c.BorderStyleRound,
						BorderColor: "magenta",
						Padding:     1,
						Style: c.StyleProps{
							Display:        c.DisplayFlex,
							FlexDirection:  c.FlexDirectionRow,
							JustifyContent: c.JustifyContentSpaceBetween,
							Height:         5,
						},
					},
						c.Box(c.BoxProps{Style: c.StyleProps{Width: 10, Background: "red"}}),
						c.Box(c.BoxProps{Style: c.StyleProps{Width: 10, Background: "blue"}}),
						c.Box(c.BoxProps{Style: c.StyleProps{Width: 10, Background: "green"}}),
					),
					c.Text(c.TextProps{Content: " JustifyContent: SpaceBetween", Style: c.TextStyle().Dim()}),

					c.Spacer(1),

					// Row 3: Align Items
					c.Box(c.BoxProps{
						BorderStyle: c.BorderStyleRound,
						BorderColor: "cyan",
						Padding:     1,
						Style: c.StyleProps{
							Display:       c.DisplayFlex,
							FlexDirection: c.FlexDirectionRow,
							AlignItems:    c.AlignItemsCenter,
							Height:        7,
							Gap:           5,
						},
					},
						c.Box(c.BoxProps{Style: c.StyleProps{Width: 10, Height: 3, Background: "yellow"}}),
						c.Box(c.BoxProps{Style: c.StyleProps{Width: 10, Height: 5, Background: "magenta"}}),
						c.Box(c.BoxProps{Style: c.StyleProps{Width: 10, Height: 2, Background: "white"}}),
					),
					c.Spacer(1),

					// Row 4: Links
					c.Box(c.BoxProps{
						Style: c.StyleProps{
							Display:       c.DisplayFlex,
							FlexDirection: c.FlexDirectionRow,
							Gap:           2,
						},
					},
						c.Link(c.LinkProps{
							Href:        "https://github.com/jiiihpeeh/recliner",
							BorderStyle: c.BorderStyleSingle,
							BorderColor: "blue",
							Padding:     1,
							Style: c.StyleProps{
								Background: "#000033",
							},
						},
							c.Text(c.TextProps{Content: "Open GitHub", Style: c.TextStyle().Color("blue").Bold()}),
						),
						c.Link(c.LinkProps{
							Command:     "ls -la",
							BorderStyle: c.BorderStyleSingle,
							BorderColor: "green",
							Padding:     1,
							Style: c.StyleProps{
								Background: "#003300",
							},
						},
							c.Text(c.TextProps{Content: "Run ls -la", Style: c.TextStyle().Color("green").Bold()}),
						),
					),
					c.Spacer(1),

					// Row 5: Accordion & Tabs Variants
					c.Box(c.BoxProps{
						Style: c.StyleProps{
							Display:       c.DisplayFlex,
							FlexDirection: c.FlexDirectionRow,
							Gap:           4,
						},
					},
						// Accordion Unicode
						c.Box(c.BoxProps{
							Style: c.StyleProps{Width: 30},
						},
							c.Text(c.TextProps{Content: "Accordion (Unicode)", Style: c.TextStyle().Bold()}),
							c.Accordion(c.AccordionProps{
								Variant: c.AccordionVariantUnicode,
								Items: []c.AccordionItem{
									{ID: "u1", Title: "Section 1", Content: c.Text(c.TextProps{Content: "Content 1"})},
									{ID: "u2", Title: "Section 2", Content: c.Text(c.TextProps{Content: "Content 2"})},
								},
							}),
							c.Spacer(1),
							c.Text(c.TextProps{Content: "Radio Group", Style: c.TextStyle().Bold()}),
							c.RadioGroup(c.RadioGroupProps{
								ID:        "test-radio",
								Value:     "opt1",
								Direction: c.FlexDirectionColumn,
								Options: []c.RadioOption{
									{Label: "Option 1", Value: "opt1"},
									{Label: "Option 2", Value: "opt2"},
								},
							}),
						),

						// Tabs Unicode
						c.Box(c.BoxProps{
							Style: c.StyleProps{FlexGrow: 1},
						},
							c.Text(c.TextProps{Content: "Tabs (Unicode)", Style: c.TextStyle().Bold()}),
							c.Tabs(c.TabsProps{
								Variant: c.TabsVariantUnicode,
								Items: []c.TabItem{
									{ID: "t1", Title: "Tab 1", Content: c.Text(c.TextProps{Content: "Content for Tab 1"})},
									{ID: "t2", Title: "Tab 2", Content: c.Text(c.TextProps{Content: "Content for Tab 2"})},
								},
							}),
							c.Spacer(1),
							c.Text(c.TextProps{Content: "Tabs (Box)", Style: c.TextStyle().Bold()}),
							c.Tabs(c.TabsProps{
								Variant: c.TabsVariantBox,
								Items: []c.TabItem{
									{ID: "b1", Title: "Box 1", Content: c.Text(c.TextProps{Content: "Content for Box 1"})},
									{ID: "b2", Title: "Box 2", Content: c.Text(c.TextProps{Content: "Content for Box 2"})},
								},
							}),
						),
					),
				),
			),

			// Footer
			c.Box(c.BoxProps{
				Style: c.StyleProps{
					Width:          100,
					Height:         1,
					Background:     "blue",
					Display:        c.DisplayFlex,
					JustifyContent: c.JustifyContentSpaceBetween,
				},
			},
				c.Text(c.TextProps{Content: " Press q to quit ", Style: c.TextStyle().Inverse()}),
				c.Text(c.TextProps{Content: " Recliner v0.1.0 ", Style: c.TextStyle().Inverse()}),
			),
		)
	}, app.AppOptions{})

	if err := appInstance.Run(); err != nil {
		panic(err)
	}
}
