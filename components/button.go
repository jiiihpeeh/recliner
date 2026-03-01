package components

import (
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
)

func Button(props any) vdom.Node {
	hc := hooks.GetContext()

	var children []vdom.Node
	if c, ok := util.GetProp[[]vdom.Node](props, "children"); ok {
		children = c
	}

	// Support both map props and struct props
	label, _ := util.GetProp[string](props, "label")
	variant, ok := util.GetProp[string](props, "variant")
	if !ok || variant == "" {
		variant = "filled"
	}
	buttonStyleStr, ok := util.GetProp[string](props, "style")
	if !ok || buttonStyleStr == "" {
		buttonStyleStr = "primary"
	}
	size, ok := util.GetProp[string](props, "size")
	if !ok || size == "" {
		size = "medium"
	}
	disabled, _ := util.GetProp[bool](props, "disabled")
	id, _ := util.GetProp[string](props, "id")
	onClick, _ := util.GetProp[func(events.MouseEvent)](props, "onClick")

	focusRes := hc.UseFocus(hooks.FocusOptions{ID: id})
	focused := focusRes.IsFocused

	hc.UseInput(func(e events.KeyPressEvent) {
		if !focused || disabled {
			return
		}
		switch e.Key {
		case "enter", " ":
			if onClick != nil {
				onClick(events.MouseEvent{X: 0, Y: 0, Button: events.MouseButtonLeft, Action: events.MouseActionPress})
			}
		}
	}, []any{disabled})

	// Determine button styling based on variant, style, and state
	finalStyle := getButtonStyle(variant, buttonStyleStr, size, focused, disabled)

	// Create the button content
	var buttonContent vdom.Node
	if len(children) > 0 {
		buttonContent = &vdom.Fragment{Children: children}
	} else {
		buttonContent = &vdom.TextNode{Content: label}
	}

	// Create the button element
	borderStyle, ok := util.GetProp[string](props, "borderStyle")
	if !ok || borderStyle == "" {
		borderStyle = vdom.BorderStyleSingle
		if variant == "text" {
			borderStyle = "none"
		}
	}
	padding, ok := util.GetProp[int](props, "padding")
	if !ok {
		padding = 2
		if size == "small" {
			padding = 1
		} else if size == "large" {
			padding = 3
		} else if variant == "text" {
			padding = 0
		}
	}
	borderColor, _ := util.GetProp[string](props, "borderColor")

	elementProps := struct {
		Style       vdom.Style
		OnClick     func(events.MouseEvent)
		BorderStyle string
		Padding     int
		BorderColor string
	}{
		Style:       finalStyle,
		BorderStyle: borderStyle,
		Padding:     padding,
		BorderColor: borderColor,
	}
	if onClick != nil && !disabled {
		elementProps.OnClick = onClick
	}

	return &vdom.Element{
		Type:     "box",
		Props:    elementProps,
		Children: []vdom.Node{buttonContent},
		Style:    finalStyle,
	}
}

func getButtonStyle(variant, style, size string, focused, disabled bool) vdom.Style {
	baseStyle := vdom.Style{
		Display:    vdom.DisplayFlex,
		AlignItems: vdom.AlignCenter,
	}

	// Size/Padding simulation (vdom.Style doesn't have Padding, but renderer handles Props)
	// We'll leave padding to the renderer via Props if needed, but for now focus on what's in vdom.Style

	// Style colors
	var bgColor, fgColor, borderColor string
	switch style {
	case "primary":
		fgColor = "white"
		borderColor = "blue"
		if variant == "filled" {
			bgColor = "blue"
		}
	case "secondary":
		fgColor = "white"
		borderColor = "gray"
		if variant == "filled" {
			bgColor = "gray"
		}
	case "success":
		fgColor = "white"
		borderColor = "green"
		if variant == "filled" {
			bgColor = "green"
		}
	case "danger":
		fgColor = "white"
		borderColor = "red"
		if variant == "filled" {
			bgColor = "red"
		}
	case "warning":
		fgColor = "black"
		borderColor = "yellow"
		if variant == "filled" {
			bgColor = "yellow"
		}
	case "info":
		fgColor = "white"
		borderColor = "cyan"
		if variant == "filled" {
			bgColor = "cyan"
		}
	default:
		fgColor = "white"
		borderColor = "blue"
		if variant == "filled" {
			bgColor = "blue"
		}
	}

	// Variant styles
	switch variant {
	case "outlined":
		baseStyle.Background = ""
		baseStyle.Foreground = borderColor
	case "text":
		baseStyle.Background = ""
		baseStyle.Foreground = borderColor
	default: // filled
		baseStyle.Background = bgColor
		baseStyle.Foreground = fgColor
	}

	// State styles
	if disabled {
		baseStyle.Dim = true
		if variant == "filled" {
			baseStyle.Background = "gray"
		}
	}

	if focused && !disabled {
		baseStyle.Bold = true
		if variant == "outlined" || variant == "text" {
			baseStyle.Background = borderColor
			baseStyle.Foreground = "black"
		}
	}

	return baseStyle
}
