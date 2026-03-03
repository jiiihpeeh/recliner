package components

import (
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
)

func Modal(props any) vdom.Node {
	hc := hooks.GetContext()

	title, _ := util.GetProp[string](props, "title")
	children, _ := util.GetProp[[]vdom.Node](props, "children")
	onClose, _ := util.GetProp[func()](props, "onClose")
	onConfirm, _ := util.GetProp[func()](props, "onConfirm")
	onCancel, _ := util.GetProp[func()](props, "onCancel")
	confirmText, _ := util.GetProp[string](props, "confirmText")
	if confirmText == "" {
		confirmText = "OK"
	}
	cancelText, _ := util.GetProp[string](props, "cancelText")
	if cancelText == "" {
		cancelText = "Cancel"
	}
	showButtons, _ := util.GetProp[bool](props, "showButtons")
	if !showButtons {
		showButtons = true
	}
	showCloseButton, _ := util.GetProp[bool](props, "showCloseButton")
	if !showCloseButton {
		showCloseButton = true
	}
	showBackdrop, _ := util.GetProp[bool](props, "showBackdrop")
	backgroundColor, _ := util.GetProp[string](props, "backgroundColor")
	if backgroundColor == "" {
		backgroundColor = "transparent"
	}

	width, _ := util.GetProp[int](props, "width")
	height, _ := util.GetProp[int](props, "height")

	termWidth, termHeight := hooks.UseWindowSize(hc)

	if width == 0 {
		width = termWidth / 2
	}
	if height == 0 {
		height = termHeight / 2
	}

	// Backdrop - only render if showBackdrop is true
	var backdrop vdom.Node
	if showBackdrop {
		backdrop = Box(struct {
			Style vdom.Style
		}{
			Style: vdom.Style{
				Position:   "fixed",
				Top:        0,
				Left:       0,
				Width:      termWidth,
				Height:     termHeight,
				ZIndex:     997,
				Background: backgroundColor,
			},
		})
	}

	var titleContent vdom.Node
	if title != "" || showCloseButton {
		headerContent := []vdom.Node{}

		if title != "" {
			headerContent = append(headerContent, Text(struct {
				Children string
				Style    vdom.Style
			}{
				Children: title,
				Style: vdom.Style{
					Bold:       true,
					Foreground: "yellow",
				},
			}))
		}

		if showCloseButton {
			closeX := Text(struct {
				Children string
				Style    vdom.Style
				OnClick  func(events.MouseEvent)
			}{
				Children: " ✕",
				Style: vdom.Style{
					Foreground: "red",
				},
				OnClick: func(e events.MouseEvent) {
					if e.Action == events.MouseActionPress {
						if onClose != nil {
							onClose()
						}
					}
				},
			})
			headerContent = append(headerContent, closeX)
		}

		if len(headerContent) == 1 {
			titleContent = headerContent[0]
		} else {
			titleContent = Box(struct {
				Style    vdom.Style
				Children []vdom.Node
			}{
				Style:    vdom.Style{Display: vdom.DisplayFlex, FlexDirection: vdom.FlexDirectionRow, JustifyContent: "spaceBetween"},
				Children: headerContent,
			})
		}
	}

	var buttons vdom.Node
	if showButtons {
		buttons = Box(struct {
			Style    vdom.Style
			Children []vdom.Node
		}{
			Style: vdom.Style{Display: vdom.DisplayFlex, FlexDirection: vdom.FlexDirectionRow, Gap: 1, JustifyContent: "flexEnd"},
			Children: []vdom.Node{
				Button(struct {
					Label   string
					OnClick func()
				}{
					Label: cancelText,
					OnClick: func() {
						if onCancel != nil {
							onCancel()
						}
						if onClose != nil {
							onClose()
						}
					},
				}),
				Button(struct {
					Label   string
					OnClick func()
				}{
					Label: confirmText,
					OnClick: func() {
						if onConfirm != nil {
							onConfirm()
						}
						if onClose != nil {
							onClose()
						}
					},
				}),
			},
		})
	}

	modalContent := []vdom.Node{}
	if title != "" {
		modalContent = append(modalContent, titleContent)
		modalContent = append(modalContent, Newline(1))
	}
	for _, child := range children {
		modalContent = append(modalContent, child)
	}
	if showButtons {
		modalContent = append(modalContent, Newline(1))
		modalContent = append(modalContent, buttons)
	}

	if height == 0 {
		height = termHeight / 2
	}

	modal := Box(struct {
		Style       vdom.Style
		BorderStyle string
		BorderColor string
		Children    []vdom.Node
	}{
		Style: vdom.Style{
			Position:   "fixed",
			Top:        (termHeight - height) / 2,
			Left:       (termWidth - width - 2) / 2,
			Width:      width,
			Height:     height,
			ZIndex:     999,
			Background: backgroundColor,
			BorderGradient: &vdom.RadialGradient{
				From:    "magenta",
				To:      "darkmagenta",
				CenterX: 0.5,
				CenterY: 0.5,
				Radius:  0.5,
			},
		},
		BorderStyle: vdom.BorderStyleRound,
		BorderColor: "magenta",
		Children:    modalContent,
	})

	if showBackdrop {
		return &vdom.Fragment{Children: []vdom.Node{backdrop, modal}}
	}
	return &vdom.Fragment{Children: []vdom.Node{modal}}
}

func RegisterModal() {
	vdom.RegisterComponent("modal", Modal)
}
