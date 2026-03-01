package components

import (
	"fmt"
	"strings"

	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
)

type ProgressBarProps struct {
	Value         float64 // 0.0 to 1.0
	Width         int
	Height        int
	Orientation   string // "horizontal", "vertical"
	LabelPosition string // "inside-left", "inside-right", "above", "below", "none"
	FromColor     string
	ToColor       string
	FillChar      string
	EmptyChar     string
}

func ProgressBar(props any) vdom.Node {
	value, _ := util.GetProp[float64](props, "value")
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}

	width, _ := util.GetProp[int](props, "width")
	height, _ := util.GetProp[int](props, "height")
	orientation, ok := util.GetProp[string](props, "orientation")
	if !ok || orientation == "" {
		orientation = "horizontal"
	}

	labelPos, ok := util.GetProp[string](props, "labelPosition")
	if !ok || labelPos == "" {
		labelPos = "none"
	}

	fromColor, _ := util.GetProp[string](props, "fromColor")
	toColor, _ := util.GetProp[string](props, "toColor")

	fillChar, ok := util.GetProp[string](props, "fillChar")
	if !ok || fillChar == "" {
		fillChar = "█"
	}
	emptyChar, ok := util.GetProp[string](props, "emptyChar")
	if !ok || emptyChar == "" {
		emptyChar = "░"
	}

	percentText := fmt.Sprintf(" %d%% ", int(value*100))

	if orientation == "horizontal" {
		if width <= 0 {
			width = 20
		}

		filledLen := int(value * float64(width))
		emptyLen := width - filledLen

		var children []vdom.Node

		// Handle above label
		if labelPos == "above" {
			children = append(children, &vdom.Element{
				Type:      "text",
				InnerText: percentText,
				Style:     vdom.Style{Display: "block"},
			})
		}

		// Bar container
		barFilled := strings.Repeat(fillChar, filledLen)
		barEmpty := strings.Repeat(emptyChar, emptyLen)

		fullBar := barFilled + barEmpty

		// Inject label inside if requested
		if labelPos == "inside-left" {
			labelRunes := []rune(percentText)
			fullRunes := []rune(fullBar)
			if len(fullRunes) >= len(labelRunes) {
				copy(fullRunes[0:], labelRunes)
				fullBar = string(fullRunes)
			}
		} else if labelPos == "inside-right" {
			labelRunes := []rune(percentText)
			fullRunes := []rune(fullBar)
			if len(fullRunes) >= len(labelRunes) {
				copy(fullRunes[len(fullRunes)-len(labelRunes):], labelRunes)
				fullBar = string(fullRunes)
			}
		}

		// Split back into filled/empty parts for gradient
		// This is slightly tricky because the label might span across the boundary
		// Simplest: one text node for the whole bar with gradient
		barStyle := vdom.Style{Display: "block"}
		if fromColor != "" && toColor != "" {
			barStyle.ForegroundGradient = &vdom.LinearGradient{
				From:      fromColor,
				To:        toColor,
				Direction: "horizontal",
			}
		}

		children = append(children, &vdom.Element{
			Type:      "text",
			InnerText: fullBar,
			Props:     struct{ Style vdom.Style }{Style: barStyle},
			Style:     barStyle,
		})

		// Handle below label
		if labelPos == "below" {
			children = append(children, &vdom.Element{
				Type:      "text",
				InnerText: percentText,
				Style:     vdom.Style{Display: "block"},
			})
		}

		return &vdom.Element{
			Type: "box",
			Props: struct {
				BorderStyle string
				Padding     int
				Style       vdom.Style
			}{
				BorderStyle: "none",
				Padding:     0,
				Style:       vdom.Style{Display: "flex", FlexDirection: "column"},
			},
			Children: children,
			Style:    vdom.Style{Display: "flex", FlexDirection: "column"},
		}
	}

	// Vertical
	if height <= 0 {
		height = 10
	}
	filledLen := int(value * float64(height))
	emptyLen := height - filledLen

	var children []vdom.Node

	if labelPos == "above" {
		children = append(children, &vdom.Element{Type: "text", InnerText: percentText, Style: vdom.Style{Display: "block"}})
	}

	// For vertical, we build from top to bottom
	// Top is empty, bottom is filled (standard TUI progress)
	barChars := make([]string, height)
	for i := 0; i < emptyLen; i++ {
		barChars[i] = emptyChar
	}
	for i := emptyLen; i < height; i++ {
		barChars[i] = fillChar
	}

	barContent := strings.Join(barChars, "\n")

	barStyle := vdom.Style{Display: "block"}
	if fromColor != "" && toColor != "" {
		barStyle.ForegroundGradient = &vdom.LinearGradient{
			From:      toColor, // Reverse because we render top to bottom but progress is bottom up
			To:        fromColor,
			Direction: "vertical",
		}
	}

	children = append(children, &vdom.Element{
		Type:      "text",
		InnerText: barContent,
		Style:     barStyle,
	})

	if labelPos == "below" {
		children = append(children, &vdom.Element{Type: "text", InnerText: percentText, Style: vdom.Style{Display: "block"}})
	}

	return &vdom.Element{
		Type: "box",
		Props: struct {
			BorderStyle string
			Padding     int
			Style       vdom.Style
		}{
			BorderStyle: "none",
			Padding:     0,
			Style:       vdom.Style{Display: "flex", FlexDirection: "column", AlignItems: "center"},
		},
		Children: children,
		Style:    vdom.Style{Display: "flex", FlexDirection: "column", AlignItems: "center"},
	}
}
