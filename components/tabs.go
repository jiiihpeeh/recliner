package components

import (
	"reflect"

	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
)

type TabItem struct {
	Title   string
	Content vdom.Node
	ID      string
}

func Tabs(props any) vdom.Node {
	hc := hooks.GetContext()

	// Parse items
	var items []TabItem
	if rawItems, ok := util.GetProp[any](props, "items"); ok {
		rv := reflect.ValueOf(rawItems)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				itemVal := rv.Index(i)
				if itemVal.Kind() == reflect.Interface {
					itemVal = itemVal.Elem()
				}

				if ti, ok := itemVal.Interface().(TabItem); ok {
					items = append(items, ti)
				} else {
					// Extract fields via reflection for compatibility with other identical structs (e.g. from package c)
					item := TabItem{}
					if id, ok := util.GetProp[string](itemVal.Interface(), "ID"); ok {
						item.ID = id
					}
					if title, ok := util.GetProp[string](itemVal.Interface(), "Title"); ok {
						item.Title = title
					}
					if content, ok := util.GetProp[vdom.Node](itemVal.Interface(), "Content"); ok {
						item.Content = content
					}
					items = append(items, item)
				}
			}
		}
	}

	id, _ := util.GetProp[string](props, "id")
	variant, _ := util.GetProp[string](props, "variant")

	defaultActive, ok := util.GetProp[string](props, "defaultActive")
	if !ok && len(items) > 0 {
		defaultActive = items[0].ID
	}

	// State for active tab
	activeID, setActiveID := hooks.UseState[string](hc, defaultActive)

	// Controlled mode
	if v, ok := util.GetProp[string](props, "activeTab"); ok && v != "" {
		activeID = v
	}

	onChange, _ := util.GetProp[func(string)](props, "onChange")

	// Build headers
	headerNodes := []vdom.Node{}
	for _, item := range items {
		itemID := item.ID
		isActive := itemID == activeID

		headerStyle := vdom.Style{
			Foreground: "white",
		}
		if isActive {
			headerStyle.Foreground = "black"
			headerStyle.Background = "cyan"
			headerStyle.Bold = true
		}

		// Focus management for headers
		tabFocusID := id + "-tab-" + itemID
		focusRes := hc.UseFocus(hooks.FocusOptions{ID: tabFocusID})
		if focusRes.IsFocused {
			headerStyle.Background = "blue"
			headerStyle.Foreground = "white"

			hc.UseInput(func(e events.KeyPressEvent) {
				if e.Key == "enter" || e.Key == " " {
					if onChange != nil {
						onChange(itemID)
					}
					setActiveID(itemID)
				}
			}, []any{itemID})
		}

		// Style the headers based on variant
		label := "[" + item.Title + "]"
		borderStyle := "none"
		padding := 0

		if variant == "unicode" {
			label = " " + item.Title + " "
			if isActive {
				label = "◉ " + item.Title + " "
			} else {
				label = "○ " + item.Title + " "
			}
		} else if variant == "box" {
			label = item.Title
			borderStyle = vdom.BorderStyleSingle
			if isActive {
				borderStyle = "bold"
			}
			padding = 0
		}

		headerNodes = append(headerNodes, &vdom.Element{
			Type: "box",
			Props: struct {
				BorderStyle string
				Padding     int
				Style       vdom.Style
				OnClick     func(events.MouseEvent)
			}{
				BorderStyle: borderStyle,
				Padding:     padding,
				Style:       headerStyle,
				OnClick: func(e events.MouseEvent) {
					if e.Action == events.MouseActionPress {
						if onChange != nil {
							onChange(itemID)
						}
						setActiveID(itemID)
					}
				},
			},
			Children: []vdom.Node{
				&vdom.Element{
					Type:      "text",
					Props:     struct{ Style vdom.Style }{Style: headerStyle},
					InnerText: label,
					Style:     headerStyle,
				},
			},
			Style: headerStyle,
		})
	}

	rowStyle := vdom.Style{
		Display:       vdom.DisplayFlex,
		FlexDirection: vdom.FlexDirectionRow,
		Gap:           1,
	}
	headerRow := &vdom.Element{
		Type: "box",
		Props: struct {
			BorderStyle string
			Style       vdom.Style
		}{
			BorderStyle: vdom.BorderStyleNone,
			Style:       rowStyle,
		},
		Children: headerNodes,
		Style:    rowStyle,
	}

	// Content area
	var activeContent vdom.Node
	for _, item := range items {
		if item.ID == activeID {
			activeContent = item.Content
			break
		}
	}

	contentBox := &vdom.Element{
		Type: "box",
		Props: struct {
			BorderStyle string
			Padding     int
		}{
			BorderStyle: vdom.BorderStyleSingle,
			Padding:     1,
		},
		Children: []vdom.Node{activeContent},
		Style:    vdom.Style{},
	}

	rootStyle := vdom.Style{Display: vdom.DisplayFlex, FlexDirection: vdom.FlexDirectionColumn}
	return &vdom.Element{
		Type: "box",
		Props: struct {
			BorderStyle string
			Style       vdom.Style
		}{
			BorderStyle: vdom.BorderStyleNone,
			Style:       rootStyle,
		},
		Children: []vdom.Node{headerRow, contentBox},
		Style:    rootStyle,
	}
}
