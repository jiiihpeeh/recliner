package components

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/j-p/recliner/events"
	"github.com/j-p/recliner/hooks"
	"github.com/j-p/recliner/util"
	"github.com/j-p/recliner/utils"
	"github.com/j-p/recliner/vdom"
	"github.com/mattn/go-runewidth"
)

type inputInternalState struct {
	value      string
	cursorPos  int
	dragStart  int
	isDragging bool
	start      int
	width      int
	focused    bool
	onChange   func(string)
}

func Input(props any) vdom.Node {
	hc := hooks.GetContext()

	value, _ := util.GetProp[string](props, "value")
	onChange, _ := util.GetProp[func(string)](props, "onChange")
	placeholder, _ := util.GetProp[string](props, "placeholder")
	width, ok := util.GetProp[int](props, "width")
	if !ok {
		width = 20
	}
	id, _ := util.GetProp[string](props, "id")
	autoFocus, _ := util.GetProp[bool](props, "autoFocus")

	focusRes := hc.UseFocus(hooks.FocusOptions{ID: id, AutoFocus: autoFocus})
	focused := focusRes.IsFocused

	inputType, ok := util.GetProp[string](props, "type")
	if !ok {
		inputType = "text"
	}

	cursorPos, setCursorPos := hooks.UseState[int](hc, 0)
	dragStart, setDragStart := hooks.UseState[int](hc, -1)
	isDragging, setIsDragging := hooks.UseState[bool](hc, false)
	showCursor, setShowCursor := hooks.UseState[bool](hc, true)

	hc.UseEffect(func() func() {
		if !focused {
			setShowCursor(false)
			return func() {}
		}
		setShowCursor(true)
		done := make(chan struct{})
		ticker := time.NewTicker(800 * time.Millisecond)
		go func() {
			visible := true
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					visible = !visible
					setShowCursor(visible)
				}
			}
		}()
		return func() { ticker.Stop(); close(done) }
	}, []any{focused})

	runes := []rune(value)
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	displayContent := value
	if inputType == "password" {
		displayContent = strings.Repeat("*", utf8.RuneCountInString(value))
	}
	showPlaceholder := len(value) == 0 && placeholder != ""
	renderRunes := []rune(displayContent)
	if showPlaceholder {
		renderRunes = []rune(placeholder)
	}

	contentWidth := width - 2
	if contentWidth < 1 {
		contentWidth = 1
	}

	cursorVisualPos := 0
	for i := 0; i < cursorPos && i < len(renderRunes); i++ {
		w := runewidth.RuneWidth(renderRunes[i])
		if w == 0 {
			w = 1
		}
		cursorVisualPos += w
	}

	start := 0
	if !showPlaceholder && focused {
		if cursorVisualPos >= contentWidth {
			accumulatedWidth := 0
			targetWidth := contentWidth - 1
			for i := cursorPos - 1; i >= 0; i-- {
				w := runewidth.RuneWidth(renderRunes[i])
				if w == 0 {
					w = 1
				}
				accumulatedWidth += w
				if accumulatedWidth >= targetWidth {
					start = i + 1
					break
				}
			}
		}
	}

	// Stability Ref
	stateRef := hooks.UseRef(hc, &inputInternalState{})
	stateRef.Value.value = value
	stateRef.Value.cursorPos = cursorPos
	stateRef.Value.dragStart = dragStart
	stateRef.Value.isDragging = isDragging
	stateRef.Value.start = start
	stateRef.Value.width = width
	stateRef.Value.focused = focused
	stateRef.Value.onChange = onChange

	handleMouse := func(e events.MouseEvent) {
		s := stateRef.Value
		if onClick, ok := util.GetProp[func(events.MouseEvent)](props, "onClick"); ok {
			onClick(e)
		}
		if !s.focused && e.Action == events.MouseActionPress {
			focusRes.Focus()
		}
		if showPlaceholder {
			if e.Action == events.MouseActionPress {
				setCursorPos(0)
				setDragStart(-1)
				setIsDragging(false)
			}
			return
		}

		clickRelX := e.RelX - 1
		if clickRelX < 0 {
			clickRelX = 0
		}
		targetVisualX, clickIndex, currentVisualX, found := clickRelX, -1, 0, false
		for i := s.start; i < len(renderRunes); i++ {
			charWidth := runewidth.RuneWidth(renderRunes[i])
			if charWidth == 0 {
				charWidth = 1
			}
			if targetVisualX >= currentVisualX && targetVisualX < currentVisualX+charWidth {
				clickIndex = i
				found = true
				break
			}
			currentVisualX += charWidth
			if currentVisualX > contentWidth {
				break
			}
		}
		if !found {
			clickIndex = len(renderRunes)
		}
		if clickIndex > len(renderRunes) {
			clickIndex = len(renderRunes)
		}
		if clickIndex < 0 {
			clickIndex = 0
		}

		switch e.Action {
		case events.MouseActionPress:
			setCursorPos(clickIndex)
			setDragStart(clickIndex)
			setIsDragging(true)
		case events.MouseActionMotion:
			if s.isDragging {
				setCursorPos(clickIndex)
			}
		case events.MouseActionRelease:
			setIsDragging(false)
		}
	}

	hc.UseInput(func(event events.KeyPressEvent) {
		if !focused {
			return
		}
		setShowCursor(true)
		if onChange == nil {
			return
		}
		runes := []rune(value)

		getSelection := func() (int, int) {
			if dragStart != -1 && dragStart != cursorPos {
				s, e := dragStart, cursorPos
				if s > e {
					s, e = e, s
				}
				return s, e
			}
			return -1, -1
		}

		deleteRange := func(start, end int) {
			newVal := string(runes[:start]) + string(runes[end:])
			setCursorPos(start)
			onChange(newVal)
			setDragStart(-1)
		}

		switch event.Key {
		case "left", "ctrl+b":
			setDragStart(-1)
			setIsDragging(false)
			if cursorPos > 0 {
				setCursorPos(cursorPos - 1)
			}
			return
		case "right", "ctrl+f":
			setDragStart(-1)
			setIsDragging(false)
			if cursorPos < len(runes) {
				setCursorPos(cursorPos + 1)
			}
			return
		case "home":
			setDragStart(-1)
			setIsDragging(false)
			setCursorPos(0)
			return
		case "end", "ctrl+e":
			setDragStart(-1)
			setIsDragging(false)
			setCursorPos(len(runes))
			return
		case "ctrl+a":
			setDragStart(0)
			setCursorPos(len(runes))
			return
		case "ctrl+c":
			s, e := getSelection()
			if s != -1 {
				utils.WriteClipboard(string(runes[s:e]))
			}
			return
		case "ctrl+x":
			s, e := getSelection()
			if s != -1 {
				utils.WriteClipboard(string(runes[s:e]))
				deleteRange(s, e)
			}
			return
		case "ctrl+v":
			text, err := utils.ReadClipboard()
			if err == nil && len(text) > 0 {
				s, e := getSelection()
				if s != -1 {
					newVal := string(runes[:s]) + text + string(runes[e:])
					setCursorPos(s + utf8.RuneCountInString(text))
					onChange(newVal)
					setDragStart(-1)
				} else {
					newVal := string(runes[:cursorPos]) + text + string(runes[cursorPos:])
					setCursorPos(cursorPos + utf8.RuneCountInString(text))
					onChange(newVal)
				}
			}
			return
		case "\x7f", "\b", "ctrl+h":
			s, e := getSelection()
			if s != -1 {
				deleteRange(s, e)
				return
			}
			if cursorPos > 0 {
				newVal := string(runes[:cursorPos-1]) + string(runes[cursorPos:])
				setCursorPos(cursorPos - 1)
				onChange(newVal)
			}
			return
		case "delete":
			s, e := getSelection()
			if s != -1 {
				deleteRange(s, e)
				return
			}
			if cursorPos < len(runes) {
				newVal := string(runes[:cursorPos]) + string(runes[cursorPos+1:])
				onChange(newVal)
			}
			return
		}

		if event.Escape || event.Ctrl || event.Meta || event.Key == "tab" || event.Key == "enter" {
			return
		}

		s, e := getSelection()
		currRunes, currCursor := runes, cursorPos
		if s != -1 {
			currRunes = append(runes[:s], runes[e:]...)
			currCursor = s
		}

		if utf8.RuneCountInString(event.Key) == 1 {
			r, _ := utf8.DecodeRuneInString(event.Key)
			if r >= 32 {
				newVal := string(currRunes[:currCursor]) + event.Key + string(currRunes[currCursor:])
				setCursorPos(currCursor + 1)
				onChange(newVal)
				setDragStart(-1)
			}
		}
	}, []any{value, focused, cursorPos, dragStart, isDragging})

	children := []vdom.Node{}
	cvw := 0
	for i := start; i < len(renderRunes); i++ {
		if cvw >= contentWidth {
			break
		}
		char, rw := string(renderRunes[i]), runewidth.RuneWidth(renderRunes[i])
		if rw == 0 {
			rw = 1
		}
		if cvw+rw > contentWidth {
			if pad := contentWidth - cvw; pad > 0 {
				children = append(children, Text(struct {
					Children string
					Style    vdom.Style
				}{
					Children: strings.Repeat(" ", pad),
					Style:    vdom.Style{Dim: showPlaceholder},
				}))
			}
			cvw = contentWidth
			break
		}
		isSelected := false
		if !showPlaceholder && dragStart != -1 && dragStart != cursorPos {
			s, e := dragStart, cursorPos
			if s > e {
				s, e = e, s
			}
			if i >= s && i < e {
				isSelected = true
			}
		}
		isCursor := showCursor && focused && i == cursorPos
		charStyle := vdom.Style{}
		if showPlaceholder {
			charStyle.Dim = true
		}
		if isCursor || isSelected {
			charStyle.Reverse = true
		}
		children = append(children, Text(struct {
			Children string
			Style    vdom.Style
		}{
			Children: char,
			Style:    charStyle,
		}))
		cvw += rw
	}

	if showCursor && focused && !showPlaceholder && cursorPos == len(renderRunes) && cvw < contentWidth {
		children = append(children, Text(struct {
			Children string
			Style    vdom.Style
		}{
			Children: " ",
			Style:    vdom.Style{Reverse: true},
		}))
		cvw++
	}
	if cvw < contentWidth {
		children = append(children, Text(struct {
			Children string
			Style    vdom.Style
		}{
			Children: strings.Repeat(" ", contentWidth-cvw),
			Style:    vdom.Style{Dim: showPlaceholder},
		}))
	}

	borderColor := "gray"
	if focused {
		borderColor = "green"
	}

	return Box(struct {
		BorderStyle       string
		BorderColor       string
		Padding           int
		ClearFocusOnClick bool
		Style             vdom.Style
		Children          []vdom.Node
		OnClick           func(events.MouseEvent)
	}{
		BorderStyle: "single", BorderColor: borderColor, Padding: 0, ClearFocusOnClick: false,
		Style:    vdom.Style{Width: width, Display: "flex", FlexDirection: "row"},
		Children: children, OnClick: handleMouse,
	})
}
