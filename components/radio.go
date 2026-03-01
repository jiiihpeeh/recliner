package components

import (
	"reflect"

	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
)

type RadioOption struct {
	Label string
	Value string
}

func RadioGroup(props any) vdom.Node {
	hc := hooks.GetContext()

	// Parse options
	var options []RadioOption
	if rawOptions, ok := util.GetProp[any](props, "options"); ok {
		rv := reflect.ValueOf(rawOptions)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				itemVal := rv.Index(i)
				if itemVal.Kind() == reflect.Interface {
					itemVal = itemVal.Elem()
				}

				if ro, ok := itemVal.Interface().(RadioOption); ok {
					options = append(options, ro)
				} else {
					// Extract fields via reflection for compatibility
					opt := RadioOption{}
					if label, ok := util.GetProp[string](itemVal.Interface(), "Label"); ok {
						opt.Label = label
					}
					if val, ok := util.GetProp[string](itemVal.Interface(), "Value"); ok {
						opt.Value = val
					}
					options = append(options, opt)
				}
			}
		}
	}

	value, _ := util.GetProp[string](props, "value")
	onChange, _ := util.GetProp[func(string)](props, "onChange")
	disabled, _ := util.GetProp[bool](props, "disabled")
	id, _ := util.GetProp[string](props, "id")
	direction, _ := util.GetProp[string](props, "direction")
	if direction == "" {
		direction = vdom.FlexDirectionColumn
	}

	gap, ok := util.GetProp[int](props, "gap")
	if !ok {
		// Default to 1 for column, 3 for row
		if direction == vdom.FlexDirectionRow {
			gap = 3
		} else {
			gap = 1
		}
	}

	var children []vdom.Node

	for _, opt := range options {
		optVal := opt.Value
		isChecked := value == optVal
		optID := id + "-" + optVal

		focusRes := hc.UseFocus(hooks.FocusOptions{ID: optID})
		focused := focusRes.IsFocused

		hc.UseInput(func(e events.KeyPressEvent) {
			if !focused || disabled {
				return
			}
			switch e.Key {
			case "enter", " ":
				if onChange != nil {
					onChange(optVal)
				}
			}
		}, []any{value, disabled})

		// Render radio button
		symbol := "○" // Unicode unselected radio
		if isChecked {
			symbol = "◉" // Unicode selected radio (fisheye/nested circle)
		}

		if focused {
			symbol = "▸" + symbol // Use Unicode arrow for focus
		}

		vStyle := vdom.Style{}
		if disabled {
			vStyle.Dim = true
		}

		elementProps := struct {
			Style   vdom.Style
			OnClick func(events.MouseEvent)
		}{
			Style: vStyle,
		}

		if onChange != nil {
			elementProps.OnClick = func(e events.MouseEvent) {
				if !disabled && e.Action == events.MouseActionPress {
					onChange(optVal)
				}
			}
		}

		radioText := &vdom.Element{
			Type:      "text",
			Props:     elementProps,
			InnerText: symbol,
			Style:     vStyle,
		}

		labelText := &vdom.Element{
			Type: "text",
			Props: struct {
				Style vdom.Style
			}{Style: vStyle},
			InnerText: " " + opt.Label,
			Style:     vStyle,
		}

		itemStyle := vdom.Style{Display: vdom.DisplayFlex, Gap: 1}
		itemProps := struct {
			BorderStyle string
			Padding     int
			Style       vdom.Style
			OnClick     func(events.MouseEvent)
		}{
			BorderStyle: vdom.BorderStyleNone,
			Padding:     0,
			Style:       itemStyle,
		}

		if onChange != nil {
			itemProps.OnClick = func(e events.MouseEvent) {
				if !disabled && e.Action == events.MouseActionPress {
					onChange(optVal)
				}
			}
		}

		itemBox := &vdom.Element{
			Type:     "box",
			Props:    itemProps,
			Children: []vdom.Node{radioText, labelText},
			Style:    itemStyle,
		}

		children = append(children, itemBox)
	}

	groupStyle := vdom.Style{
		Display:       vdom.DisplayFlex,
		FlexDirection: direction,
		Gap:           gap,
	}

	return &vdom.Element{
		Type: "box",
		Props: struct {
			BorderStyle string
			Padding     int
			Style       vdom.Style
		}{
			BorderStyle: vdom.BorderStyleNone,
			Padding:     0,
			Style:       groupStyle,
		},
		Children: children,
		Style:    groupStyle,
	}
}
