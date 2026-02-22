package components

import (
	"github.com/j-p/recliner/events"
	"github.com/j-p/recliner/hooks"
	"github.com/j-p/recliner/util"
	"github.com/j-p/recliner/vdom"
)

func Box(props any) vdom.Node {
	hc := hooks.GetContext()

	var children []vdom.Node
	if c, ok := util.GetProp[[]vdom.Node](props, "children"); ok {
		children = c
	}

	borderStyle, ok := util.GetProp[string](props, "borderStyle")
	if !ok || borderStyle == "" {
		borderStyle = "single"
	}
	borderColor, _ := util.GetProp[string](props, "borderColor")
	padding, _ := util.GetProp[int](props, "padding")

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

	width, _ := util.GetProp[int](props, "width")
	if width == 0 {
		width = style.Width
	}
	height, _ := util.GetProp[int](props, "height")
	if height == 0 {
		height = style.Height
	}

	scrollable, _ := util.GetProp[bool](props, "scrollable")
	contentWidth, ok := util.GetProp[int](props, "contentWidth")
	if !ok {
		contentWidth = width
	}
	contentHeight, ok := util.GetProp[int](props, "contentHeight")
	if !ok {
		contentHeight = height
	}

	scrollTop, setScrollTop := hooks.UseState[int](hc, 0)
	scrollLeft, setScrollLeft := hooks.UseState[int](hc, 0)
	isDraggingV, setIsDraggingV := hooks.UseState[bool](hc, false)
	isDraggingH, setIsDraggingH := hooks.UseState[bool](hc, false)

	id, _ := util.GetProp[string](props, "id")
	autoFocus, _ := util.GetProp[bool](props, "autoFocus")
	focusRes := hc.UseFocus(hooks.FocusOptions{ID: id, AutoFocus: autoFocus})
	focused := focusRes.IsFocused

	hc.UseInput(func(event events.KeyPressEvent) {
		if !focused || !scrollable {
			return
		}
		switch event.Key {
		case "up":
			if scrollTop > 0 {
				setScrollTop(scrollTop - 1)
			}
		case "down":
			maxScroll := contentHeight - height
			if maxScroll > 0 && scrollTop < maxScroll {
				setScrollTop(scrollTop + 1)
			}
		case "left":
			if scrollLeft > 0 {
				setScrollLeft(scrollLeft - 1)
			}
		case "right":
			maxScroll := contentWidth - width
			if maxScroll > 0 && scrollLeft < maxScroll {
				setScrollLeft(scrollLeft + 1)
			}
		}
	}, []any{focused, scrollable, scrollTop, scrollLeft, contentHeight, contentWidth, height, width})

	finalProps := struct {
		BorderStyle string
		BorderColor string
		Padding     int
		Style       vdom.Style
		ScrollTop   int
		ScrollLeft  int
		OnClick     func(events.MouseEvent)
	}{
		BorderStyle: borderStyle,
		BorderColor: borderColor,
		Padding:     padding,
		Style:       style,
		ScrollTop:   scrollTop,
		ScrollLeft:  scrollLeft,
	}

	if v, ok := util.GetProp[string](props, "alignItems"); ok {
		style.AlignItems = v
		finalProps.Style.AlignItems = v
	}
	if v, ok := util.GetProp[string](props, "justifyContent"); ok {
		style.JustifyContent = v
		finalProps.Style.JustifyContent = v
	}

	clear, _ := util.GetProp[bool](props, "clearFocusOnClick")

	onClick := func(e events.MouseEvent) {
		if e.Action == events.MouseActionPress && e.Button == events.MouseButtonLeft {
			if clear {
				hc.UseFocusManager().Blur()
			}
			if scrollable && id != "" {
				hc.UseFocusManager().Focus(id)
			}
		}
		if oc, ok := util.GetProp[func(events.MouseEvent)](props, "onClick"); ok {
			oc(e)
		}
	}

	hasHandler := false
	if _, ok := util.GetProp[func(events.MouseEvent)](props, "onClick"); ok {
		hasHandler = true
	}
	if clear || scrollable {
		hasHandler = true
	}

	if hasHandler {
		finalProps.OnClick = onClick
	}

	if !scrollable {
		return &vdom.Element{
			Type:     "box",
			Props:    finalProps,
			Children: children,
			Style:    style,
		}
	}

	// Render scrollbars if content is larger than box
	showV := contentHeight > height && height > 0
	showH := contentWidth > width && width > 0

	innerChildren := children
	var rowChildren []vdom.Node

	borderSpace := 1
	if borderStyle == "none" {
		borderSpace = 0
	}

	innerWidth := width - (borderSpace * 2)
	innerHeight := height - (borderSpace * 2)

	contentBoxWidth := innerWidth
	if showV {
		contentBoxWidth--
	}

	contentBoxHeight := innerHeight
	if showH {
		contentBoxHeight--
	}

	contentBox := &vdom.Element{
		Type: "box",
		Props: struct {
			BorderStyle string
			Padding     int
			Style       vdom.Style
			ScrollTop   int
			ScrollLeft  int
		}{
			BorderStyle: "none",
			Padding:     padding,
			Style: vdom.Style{
				Width:  contentBoxWidth,
				Height: contentBoxHeight,
			},
			ScrollTop:  scrollTop,
			ScrollLeft: scrollLeft,
		},
		Children: innerChildren,
		Style:    vdom.Style{Width: contentBoxWidth, Height: contentBoxHeight},
	}

	rowChildren = append(rowChildren, contentBox)

	if showV {
		barLength := contentBoxHeight
		thumbSize := (barLength * barLength) / contentHeight
		if thumbSize < 1 {
			thumbSize = 1
		}
		maxPos := barLength - thumbSize
		maxScroll := contentHeight - height
		pos := 0
		if maxScroll > 0 && maxPos > 0 {
			pos = (scrollTop * maxPos) / maxScroll
		}

		vBar := ScrollBar(struct {
			Orientation string
			Length      int
			Pos         int
			Size        int
			ThumbColor  string
			TrackColor  string
			OnClick     func(events.MouseEvent)
		}{
			Orientation: "vertical",
			Length:      barLength,
			Pos:         pos,
			Size:        thumbSize,
			ThumbColor: func() string {
				if isDraggingV {
					return "yellow"
				}
				return "white"
			}(),
			TrackColor: "gray",
			OnClick: func(me events.MouseEvent) {
				if me.Button != events.MouseButtonLeft {
					return
				}
				desiredPos := max(me.RelY-(thumbSize/2), 0)
				if desiredPos > maxPos {
					desiredPos = maxPos
				}
				newScroll := 0
				if maxPos > 0 && maxScroll > 0 {
					newScroll = (desiredPos * maxScroll) / maxPos
				}
				switch me.Action {
				case events.MouseActionPress:
					setIsDraggingV(true)
					setScrollTop(newScroll)
				case events.MouseActionMotion:
					setScrollTop(newScroll)
				case events.MouseActionRelease:
					setScrollTop(newScroll)
					setIsDraggingV(false)
				}
			},
		})
		rowChildren = append(rowChildren, vBar)
	}

	mainContent := &vdom.Element{
		Type: "box",
		Props: struct {
			BorderStyle string
			Style       vdom.Style
		}{
			BorderStyle: "none",
			Style:       vdom.Style{Display: "flex", FlexDirection: "row"},
		},
		Children: rowChildren,
		Style:    vdom.Style{Display: "flex", FlexDirection: "row"},
	}

	finalChildren := []vdom.Node{mainContent}

	if showH {
		barLength := innerWidth
		thumbSize := max((barLength*barLength)/contentWidth, 1)
		maxPos := barLength - thumbSize
		maxScroll := contentWidth - width
		pos := 0
		if maxScroll > 0 && maxPos > 0 {
			pos = (scrollLeft * maxPos) / maxScroll
		}

		hBar := ScrollBar(struct {
			Orientation string
			Length      int
			Pos         int
			Size        int
			ThumbColor  string
			TrackColor  string
			Style       vdom.Style
			OnClick     func(events.MouseEvent)
		}{
			Orientation: "horizontal",
			Length:      barLength,
			Pos:         pos,
			Size:        thumbSize,
			ThumbColor: func() string {
				if isDraggingH {
					return "yellow"
				}
				return "white"
			}(),
			TrackColor: "gray",
			Style:      vdom.Style{Height: 1},
			OnClick: func(me events.MouseEvent) {
				if me.Button != events.MouseButtonLeft {
					return
				}
				desiredPos := me.RelX - (thumbSize / 2)
				if desiredPos < 0 {
					desiredPos = 0
				}
				if desiredPos > maxPos {
					desiredPos = maxPos
				}
				newScroll := 0
				if maxPos > 0 && maxScroll > 0 {
					newScroll = (desiredPos * maxScroll) / maxPos
				}
				switch me.Action {
				case events.MouseActionPress:
					setIsDraggingH(true)
					setScrollLeft(newScroll)
				case events.MouseActionMotion:
					setScrollLeft(newScroll)
				case events.MouseActionRelease:
					setScrollLeft(newScroll)
					setIsDraggingH(false)
				}
			},
		})
		finalChildren = append(finalChildren, hBar)
	}

	return &vdom.Element{
		Type:     "box",
		Props:    finalProps,
		Children: finalChildren,
		Style:    style,
	}
}

func Text(props any) vdom.Node {
	content := ""
	if c, ok := util.GetProp[string](props, "children"); ok {
		content = c
	} else if nodes, ok := util.GetProp[[]vdom.Node](props, "children"); ok {
		for _, n := range nodes {
			if tn, ok := n.(*vdom.TextNode); ok {
				content += tn.Content
			} else if el, ok := n.(*vdom.Element); ok && el.Type == "text" {
				content += el.InnerText
			}
		}
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

	elementProps := struct {
		Style   vdom.Style
		OnClick func(events.MouseEvent)
	}{
		Style: style,
	}
	if onClick, ok := util.GetProp[func(events.MouseEvent)](props, "onClick"); ok {
		elementProps.OnClick = onClick
	}

	return &vdom.Element{
		Type:      "text",
		Props:     elementProps,
		InnerText: content,
		Style:     style,
	}
}

func TextInput(props any) vdom.Node {
	value, _ := util.GetProp[string](props, "value")
	placeholder, _ := util.GetProp[string](props, "placeholder")
	width, _ := util.GetProp[int](props, "width")

	return &vdom.Element{
		Type: "textInput",
		Props: struct {
			Value       string
			Placeholder string
			Width       int
		}{
			Value:       value,
			Placeholder: placeholder,
			Width:       width,
		},
	}
}

func Spacer(props any) vdom.Node {
	height, ok := util.GetProp[int](props, "height")
	if !ok {
		height = 1
	}

	children := make([]vdom.Node, height)
	for i := 0; i < height; i++ {
		children[i] = &vdom.TextNode{Content: "\n"}
	}
	return &vdom.Fragment{Children: children}
}

func Newline(props any) vdom.Node {
	count, ok := util.GetProp[int](props, "count")
	if !ok {
		count = 1
	}

	children := make([]vdom.Node, count)
	for i := 0; i < count; i++ {
		children[i] = &vdom.TextNode{Content: "\n"}
	}
	return &vdom.Fragment{Children: children}
}

func Register() {
	vdom.RegisterComponent("box", Box)
	vdom.RegisterComponent("text", Text)
	vdom.RegisterComponent("textInput", TextInput)
	vdom.RegisterComponent("spacer", Spacer)
	vdom.RegisterComponent("newline", Newline)
	vdom.RegisterComponent("checkbox", CheckBox)
	vdom.RegisterComponent("textbox", TextBox)
	vdom.RegisterComponent("button", Button)
	vdom.RegisterComponent("scrollbar", ScrollBar)
	vdom.RegisterComponent("accordion", Accordion)
	vdom.RegisterComponent("tabs", Tabs)
	vdom.RegisterComponent("image", Image)
	vdom.RegisterComponent("progressbar", ProgressBar)
}
