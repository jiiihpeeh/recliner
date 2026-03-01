package components

import (
	"reflect"

	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
)

type AccordionItem struct {
	Title   string
	Content vdom.Node
	ID      string
}

func Accordion(props any) vdom.Node {
	hc := hooks.GetContext()

	// Parse items
	var items []AccordionItem
	if rawItems, ok := util.GetProp[any](props, "items"); ok {
		rv := reflect.ValueOf(rawItems)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				itemVal := rv.Index(i)
				if itemVal.Kind() == reflect.Interface {
					itemVal = itemVal.Elem()
				}

				if ai, ok := itemVal.Interface().(AccordionItem); ok {
					items = append(items, ai)
				} else {
					// Extract fields via reflection for compatibility with other identical structs (e.g. from package c)
					item := AccordionItem{}
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

	allowMultiple, _ := util.GetProp[bool](props, "allowMultiple")
	defaultExpanded, _ := util.GetProp[[]string](props, "defaultExpanded")
	variant, _ := util.GetProp[string](props, "variant")
	id, _ := util.GetProp[string](props, "id")

	// State for expanded items
	expandedItems, setExpandedItems := hooks.UseState[[]string](hc, defaultExpanded)

	// Build accordion items
	children := make([]vdom.Node, len(items))
	for i, item := range items {
		isExpanded := false
		for _, expandedID := range expandedItems {
			if expandedID == item.ID {
				isExpanded = true
				break
			}
		}

		// Create header button
		headerID := id + "-header-" + item.ID
		it := item // closure capture
		header := createAccordionHeader(item.Title, isExpanded, headerID, variant, func() {
			newExpanded := make([]string, 0, len(expandedItems))

			if allowMultiple {
				// Multiple mode: toggle this item
				found := false
				for _, expandedID := range expandedItems {
					if expandedID == it.ID {
						found = true
					} else {
						newExpanded = append(newExpanded, expandedID)
					}
				}
				if !found {
					newExpanded = append(newExpanded, it.ID)
				}
			} else {
				// Single mode: close others, toggle this one
				if !isExpanded {
					newExpanded = []string{it.ID}
				}
			}

			setExpandedItems(newExpanded)
		})

		// Create content area
		var content vdom.Node
		if isExpanded {
			content = &vdom.Element{
				Type: "box",
				Props: struct {
					BorderStyle string
					Padding     int
				}{
					BorderStyle: vdom.BorderStyleNone,
					Padding:     1,
				},
				Children: []vdom.Node{item.Content},
				Style:    vdom.Style{},
			}
		}

		// Combine header and content
		itemChildren := []vdom.Node{header}
		if content != nil {
			itemChildren = append(itemChildren, content)
		}

		children[i] = &vdom.Element{
			Type: "box",
			Props: struct {
				BorderStyle string
				Padding     int
			}{
				BorderStyle: vdom.BorderStyleNone,
				Padding:     0,
			},
			Children: itemChildren,
			Style:    vdom.Style{},
		}
	}

	boxS := vdom.Style{Display: vdom.DisplayFlex, FlexDirection: vdom.FlexDirectionColumn}
	return &vdom.Element{
		Type: "box",
		Props: struct {
			BorderStyle string
			Padding     int
			Style       vdom.Style
		}{
			BorderStyle: vdom.BorderStyleNone,
			Padding:     0,
			Style:       boxS,
		},
		Children: children,
		Style:    boxS,
	}
}

func createAccordionHeader(title string, isExpanded bool, id string, variant string, onClick func()) vdom.Node {
	hc := hooks.GetContext()

	focusRes := hc.UseFocus(hooks.FocusOptions{ID: id})
	focused := focusRes.IsFocused

	hc.UseInput(func(e events.KeyPressEvent) {
		if !focused {
			return
		}
		switch e.Key {
		case "enter", " ":
			onClick()
		}
	}, []any{focused})

	// Create expand/collapse indicator
	indicator := "[+]"
	if isExpanded {
		indicator = "[-]"
	}

	if variant == "unicode" {
		indicator = "▶"
		if isExpanded {
			indicator = "▼"
		}
	}

	// Style the header
	headerStyle := vdom.Style{
		Display:    vdom.DisplayFlex,
		AlignItems: vdom.AlignCenter,
		Foreground: "white",
	}

	if focused {
		headerStyle.Bold = true
		headerStyle.Background = "blue"
		headerStyle.Foreground = "white"
	} else if isExpanded {
		headerStyle.Foreground = "cyan"
	}

	headerProps := struct {
		BorderStyle string
		Style       vdom.Style
		OnClick     func(events.MouseEvent)
	}{
		BorderStyle: vdom.BorderStyleNone,
		Style:       headerStyle,
		OnClick: func(e events.MouseEvent) {
			if e.Action == events.MouseActionPress {
				onClick()
			}
		},
	}

	return &vdom.Element{
		Type:  "box",
		Props: headerProps,
		Children: []vdom.Node{
			&vdom.Element{
				Type:      "text",
				Props:     struct{ Style vdom.Style }{Style: headerStyle},
				InnerText: indicator + " " + title,
				Style:     headerStyle,
			},
		},
		Style: headerStyle,
	}
}
