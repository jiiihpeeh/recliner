package components

import (
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
)

func Menu(props any) vdom.Node {
	hooksCtx := hooks.GetContext()

	title, _ := util.GetProp[string](props, "title")
	items, _ := util.GetProp[[]string](props, "items")
	onSelect, _ := util.GetProp[func(string)](props, "onSelect")
	onClose, _ := util.GetProp[func()](props, "onClose")
	x, _ := util.GetProp[int](props, "x")
	y, _ := util.GetProp[int](props, "y")

	selectedIndex, setSelectedIndex := hooks.UseState[int](hooksCtx, 0)

	// Reset selection when items change
	hooksCtx.UseEffect(func() func() {
		setSelectedIndex(0)
		return nil
	}, []any{items})

	hooksCtx.UseInput(func(e events.KeyPressEvent) {
		switch e.Key {
		case "up":
			if selectedIndex > 0 {
				setSelectedIndex(selectedIndex - 1)
			}
		case "down":
			if selectedIndex < len(items)-1 {
				setSelectedIndex(selectedIndex + 1)
			}
		case "enter":
			if len(items) > 0 {
				if onSelect != nil {
					onSelect(items[selectedIndex])
				}
				if onClose != nil {
					onClose()
				}
			}
		case "escape":
			if onClose != nil {
				onClose()
			}
		}
	}, []any{selectedIndex, items})

	children := []vdom.Node{}
	if title != "" {
		children = append(children, Text(struct {
			Children string
			Style    vdom.Style
		}{
			Children: title,
			Style: vdom.Style{
				Bold:       true,
				Underline:  true,
				Foreground: "yellow",
			},
		}))
		children = append(children, Text(struct {
			Children string
			Style    vdom.Style
		}{
			Children: " ",
			Style:    vdom.Style{},
		}))
	}

	for i, item := range items {
		style := vdom.Style{
			Foreground: "white",
		}
		prefix := "  "
		if i == selectedIndex {
			style.Foreground = "black"
			style.Background = "cyan"
			prefix = "> "
		}

		idx := i
		it := item
		// Make each menu item clickable with mouse
		children = append(children, Text(struct {
			Children string
			Style    vdom.Style
			OnClick  func(events.MouseEvent)
		}{
			Children: prefix + it,
			Style:    style,
			OnClick: func(e events.MouseEvent) {
				switch e.Action {
				case events.MouseActionMotion:
					// Hover: update selection but don't activate
					setSelectedIndex(idx)
				case events.MouseActionPress:
					setSelectedIndex(idx)
					if onSelect != nil {
						onSelect(it)
					}
					if onClose != nil {
						onClose()
					}
				}
			},
		}))
	}

	return Box(struct {
		Style       vdom.Style
		BorderStyle string
		BorderColor string
		Children    []vdom.Node
	}{
		Style: vdom.Style{
			Position:   "fixed",
			Top:        y,
			Left:       x,
			ZIndex:     999, // High zIndex as requested
			Background: "black",
		},
		BorderStyle: "round",
		BorderColor: "cyan",
		Children:    children,
	})
}
