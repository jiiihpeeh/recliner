package components

import (
	"strings"

	"github.com/j-p/recliner/events"
	"github.com/j-p/recliner/util"
	"github.com/j-p/recliner/vdom"
)

// ScrollBar renders a simple unicode scrollbar. Props:
// - orientation: "vertical" or "horizontal" (default "vertical")
// - length: int (height for vertical, width for horizontal)
// - pos: int (start index of thumb)
// - size: int (thumb length)
// - style: vdom.Style (optional style)
// - onClick: func(events.MouseEvent) (optional)
func ScrollBar(props any) vdom.Node {
	orientation, ok := util.GetProp[string](props, "orientation")
	if !ok {
		orientation = "vertical"
	}

	length, _ := util.GetProp[int](props, "length")
	pos, _ := util.GetProp[int](props, "pos")
	size, ok := util.GetProp[int](props, "size")
	if !ok {
		size = 1
	}

	var style vdom.Style
	if s, ok := util.GetProp[vdom.Style](props, "style"); ok {
		style = s
	} else if sm, ok := util.GetProp[any](props, "style"); ok {
		switch v := sm.(type) {
		case map[string]any:
			style = vdom.ParseStyle(v)
		case vdom.Style:
			style = v
		case *vdom.Style:
			if v != nil {
				style = *v
			}
		}
	}

	// Colors
	thumbColor, ok := util.GetProp[string](props, "thumbColor")
	if !ok {
		thumbColor = "white"
	}
	trackColor, ok := util.GetProp[string](props, "trackColor")
	if !ok {
		trackColor = "gray"
	}

	// Build children nodes so we can style thumb and track separately
	children := []vdom.Node{}

	if orientation == "horizontal" {
		if length <= 0 {
			length = 10
		}
		trackChar := "░"
		thumbChar := "█"

		// Clamp pos and size
		if pos < 0 {
			pos = 0
		}
		if size < 1 {
			size = 1
		}
		if pos+size > length {
			pos = length - size
		}
		if pos < 0 {
			pos = 0
			size = length
		}

		// before thumb
		if pos > 0 {
			seg := strings.Repeat(trackChar, pos)
			segStyle := vdom.Style{Foreground: trackColor}
			children = append(children, &vdom.Element{
				Type: "text",
				Props: struct {
					Style vdom.Style
				}{
					Style: segStyle,
				},
				InnerText: seg,
				Style:     segStyle,
			})
		}

		// thumb
		if size > 0 {
			thumbSeg := strings.Repeat(thumbChar, size)
			thumbStyle := vdom.Style{Foreground: thumbColor}
			children = append(children, &vdom.Element{
				Type: "text",
				Props: struct {
					Style vdom.Style
				}{
					Style: thumbStyle,
				},
				InnerText: thumbSeg,
				Style:     thumbStyle,
			})
		}

		// after thumb
		after := length - (pos + size)
		if after > 0 {
			seg := strings.Repeat(trackChar, after)
			segStyle := vdom.Style{Foreground: trackColor}
			children = append(children, &vdom.Element{
				Type: "text",
				Props: struct {
					Style vdom.Style
				}{
					Style: segStyle,
				},
				InnerText: seg,
				Style:     segStyle,
			})
		}

		// Wrap in a flex row so children render inline
		boxStyle := style
		boxStyle.Display = "flex"
		boxStyle.FlexDirection = "row"

		finalProps := struct {
			Style       vdom.Style
			BorderStyle string
			OnClick     func(events.MouseEvent)
		}{
			Style:       boxStyle,
			BorderStyle: "none",
		}
		if onClick, ok := util.GetProp[func(events.MouseEvent)](props, "onClick"); ok {
			finalProps.OnClick = onClick
		}

		return &vdom.Element{Type: "box", Props: finalProps, Children: children, Style: boxStyle}
	}

	// vertical
	if length <= 0 {
		length = 5
	}
	trackChar := "░"
	thumbChar := "█"

	// Clamp pos and size
	if pos < 0 {
		pos = 0
	}
	if size < 1 {
		size = 1
	}
	if pos+size > length {
		pos = length - size
	}
	if pos < 0 {
		pos = 0
		size = length
	}

	for i := 0; i < length; i++ {
		var ch string
		var s vdom.Style
		if i >= pos && i < pos+size {
			ch = thumbChar
			s = vdom.Style{Foreground: thumbColor}
		} else {
			ch = trackChar
			s = vdom.Style{Foreground: trackColor}
		}
		children = append(children, &vdom.Element{
			Type: "text",
			Props: struct {
				Style vdom.Style
			}{
				Style: s,
			},
			InnerText: ch,
			Style:     s,
		})
	}

	finalProps := struct {
		Style       vdom.Style
		BorderStyle string
		OnClick     func(events.MouseEvent)
	}{
		Style:       style,
		BorderStyle: "none",
	}
	if onClick, ok := util.GetProp[func(events.MouseEvent)](props, "onClick"); ok {
		finalProps.OnClick = onClick
	}

	// Vertical stack (default block flow stacks children vertically)
	return &vdom.Element{Type: "box", Props: finalProps, Children: children, Style: style}
}
