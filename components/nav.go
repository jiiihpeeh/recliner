package components

import (
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
)

type NavItem struct {
	ID      string
	Label   string
	Icon    string
	OnClick func()
}

func Nav(props any) vdom.Node {
	hc := hooks.GetContext()

	items, _ := util.GetProp[[]NavItem](props, "items")
	activeID, _ := util.GetProp[string](props, "activeID")
	onSelect, _ := util.GetProp[func(string)](props, "onSelect")

	width, _ := util.GetProp[int](props, "width")
	if width == 0 {
		width = 20
	}

	termWidth, _ := hooks.UseWindowSize(hc)

	navItems := []vdom.Node{}
	for _, item := range items {
		it := item
		isActive := it.ID == activeID

		style := vdom.Style{
			Foreground: "white",
		}
		if isActive {
			style.Foreground = "black"
			style.Background = "cyan"
		}

		navItems = append(navItems, Text(struct {
			Children string
			Style    vdom.Style
			OnClick  func(events.MouseEvent)
		}{
			Children: " " + it.Label + " ",
			Style:    style,
			OnClick: func(e events.MouseEvent) {
				if e.Action == events.MouseActionPress {
					if onSelect != nil {
						onSelect(it.ID)
					}
					if it.OnClick != nil {
						it.OnClick()
					}
				}
			},
		}))
	}

	return Box(struct {
		Style       vdom.Style
		BorderStyle string
		Children    []vdom.Node
	}{
		Style: vdom.Style{
			Position:       "fixed",
			Top:            0,
			Left:           0,
			Width:          termWidth,
			Display:        vdom.DisplayFlex,
			FlexDirection:  vdom.FlexDirectionRow,
			JustifyContent: "flexStart",
			Background:     "blue",
			ZIndex:         100,
		},
		BorderStyle: vdom.BorderStyleNone,
		Children:    navItems,
	})
}

func Navbar(props any) vdom.Node {
	hc := hooks.GetContext()

	items, _ := util.GetProp[[]NavItem](props, "items")
	activeID, _ := util.GetProp[string](props, "activeID")
	onSelect, _ := util.GetProp[func(string)](props, "onSelect")
	title, _ := util.GetProp[string](props, "title")

	termWidth, _ := hooks.UseWindowSize(hc)

	var navItems []vdom.Node
	if title != "" {
		navItems = append(navItems, Text(struct {
			Children string
			Style    vdom.Style
		}{
			Children: " " + title + " ",
			Style: vdom.Style{
				Bold:       true,
				Foreground: "white",
			},
		}))
	}

	for _, item := range items {
		it := item
		isActive := it.ID == activeID

		style := vdom.Style{
			Foreground: "white",
		}
		if isActive {
			style.Foreground = "black"
			style.Background = "cyan"
		}

		navItems = append(navItems, Text(struct {
			Children string
			Style    vdom.Style
			OnClick  func(events.MouseEvent)
		}{
			Children: " " + it.Label + " ",
			Style:    style,
			OnClick: func(e events.MouseEvent) {
				if e.Action == events.MouseActionPress {
					if onSelect != nil {
						onSelect(it.ID)
					}
					if it.OnClick != nil {
						it.OnClick()
					}
				}
			},
		}))
	}

	return Box(struct {
		Style       vdom.Style
		BorderStyle string
		Children    []vdom.Node
	}{
		Style: vdom.Style{
			Position:       "fixed",
			Top:            0,
			Left:           0,
			Width:          termWidth,
			Display:        vdom.DisplayFlex,
			FlexDirection:  vdom.FlexDirectionRow,
			JustifyContent: "flexStart",
			Background:     "blue",
			ZIndex:         100,
		},
		BorderStyle: vdom.BorderStyleNone,
		Children:    navItems,
	})
}

func RegisterNav() {
	vdom.RegisterComponent("nav", Nav)
	vdom.RegisterComponent("navbar", Navbar)
}
