package components

import (
	"github.com/j-p/recliner/events"
	"github.com/j-p/recliner/hooks"
	"github.com/j-p/recliner/util"
	"github.com/j-p/recliner/vdom"
)

func CheckBox(props any) vdom.Node {
	hc := hooks.GetContext()

	checked, _ := util.GetProp[bool](props, "checked")
	onChange, _ := util.GetProp[func(bool)](props, "onChange")
	labelAny, _ := util.GetProp[any](props, "label")
	var label vdom.Node
	if ln, ok := labelAny.(vdom.Node); ok {
		label = ln
	}
	disabled, _ := util.GetProp[bool](props, "disabled")
	id, _ := util.GetProp[string](props, "id")

	focusRes := hc.UseFocus(hooks.FocusOptions{ID: id})
	focused := focusRes.IsFocused

	hc.UseInput(func(e events.KeyPressEvent) {
		if !focused || disabled {
			return
		}
		switch e.Key {
		case "enter", " ":
			if onChange != nil {
				onChange(!checked)
			}
		}
	}, []any{checked, disabled})

	// Render checkbox
	checkSymbol := "☐" // Unicode unchecked box
	if checked {
		checkSymbol = "☑" // Unicode checked box
	}

	if focused {
		checkSymbol = "▸" + checkSymbol // Use Unicode arrow for focus
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
				onChange(!checked)
			}
		}
	}

	checkText := &vdom.Element{
		Type:      "text",
		Props:     elementProps,
		InnerText: checkSymbol,
		Style:     vStyle,
	}

	if label != nil {
		// Return Box with flex row containing check symbol and label
		boxStyle := vdom.Style{Display: "flex", Gap: 1}
		boxProps := struct {
			BorderStyle string
			Padding     int
			Style       vdom.Style
			OnClick     func(events.MouseEvent)
		}{
			BorderStyle: "none",
			Padding:     0,
			Style:       boxStyle,
		}
		if onChange != nil {
			boxProps.OnClick = func(e events.MouseEvent) {
				if !disabled && e.Action == events.MouseActionPress {
					onChange(!checked)
				}
			}
		}
		return &vdom.Element{
			Type:     "box",
			Props:    boxProps,
			Children: []vdom.Node{checkText, label},
			Style:    boxStyle,
		}
	}
	// Return just the check symbol
	return checkText
}
