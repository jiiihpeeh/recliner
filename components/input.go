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
	value             string
	cursorPos         int
	lastSeenPropValue string
	dragStart         int
	isDragging        bool
	start             int
	width             int
	focused           bool
	onChange          func(string)
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

	initialized, setInitialized := hooks.UseState[string](hc, "")
	// We use UseState for cursorPos to trigger re-renders,
	// but we don't treat it as the source of truth for the logic.
	_, setCursorPos := hooks.UseState[int](hc, 0)
	stateRef := hooks.UseRef(hc, &inputInternalState{})

	s := stateRef.Value

	// Initialization logic: snap cursor to end once per unique component ID
	if initialized != id {
		valLen := utf8.RuneCountInString(value)
		s.value = value
		s.lastSeenPropValue = value
		s.cursorPos = valLen
		setCursorPos(valLen)
		setInitialized(id)
	}

	// Synchronization: If the prop 'value' changed externally (not via our own onChange),
	// we must update our local state and Ref to match.
	if value != s.lastSeenPropValue {
		s.value = value
		s.lastSeenPropValue = value
		valLen := utf8.RuneCountInString(value)
		// Only snap to end if we were at the end of the previous string
		// or if we are currently out of bounds.
		if s.cursorPos > valLen {
			s.cursorPos = valLen
			setCursorPos(valLen)
		}
	}

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

	displayContent := s.value
	if inputType == "password" {
		displayContent = strings.Repeat("*", utf8.RuneCountInString(s.value))
	}
	showPlaceholder := len(s.value) == 0 && placeholder != ""
	renderRunes := []rune(displayContent)
	if showPlaceholder {
		renderRunes = []rune(placeholder)
	}

	contentWidth := width - 2
	if contentWidth < 1 {
		contentWidth = 1
	}

	// ALWAYS use the Ref's cursorPos for measurement and rendering.
	// This makes the UI instant and bypasses render-loop lag.
	renderCursorPos := s.cursorPos
	if renderCursorPos > len(renderRunes) {
		renderCursorPos = len(renderRunes)
	}

	cursorVisualPos := 0
	for i := 0; i < renderCursorPos && i < len(renderRunes); i++ {
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
			for i := renderCursorPos - 1; i >= 0; i-- {
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

	// Update reference values for the event handler
	s.dragStart = dragStart
	s.isDragging = isDragging
	s.start = start
	s.width = width
	s.focused = focused
	s.onChange = onChange

	handleMouse := func(e events.MouseEvent) {
		if e.Action == events.MouseActionPress {
			hc.UseFocusManager().Focus(id)
		}

		if onClick, ok := util.GetProp[func(events.MouseEvent)](props, "onClick"); ok && onClick != nil {
			onClick(e)
		}

		if showPlaceholder {
			if e.Action == events.MouseActionPress {
				s.cursorPos = 0
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
			s.cursorPos = clickIndex
			setCursorPos(clickIndex)
			setDragStart(clickIndex)
			setIsDragging(true)
		case events.MouseActionMotion:
			if s.isDragging {
				s.cursorPos = clickIndex
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

		// Absolute latest state from Ref
		currentValue := s.value
		currentPos := s.cursorPos
		currentDragStart := s.dragStart
		runes := []rune(currentValue)

		getSelection := func() (int, int) {
			if currentDragStart != -1 && currentDragStart != currentPos {
				start, end := currentDragStart, currentPos
				if start > end {
					start, end = end, start
				}
				return start, end
			}
			return -1, -1
		}

		deleteRange := func(start, end int) {
			newVal := string(runes[:start]) + string(runes[end:])
			s.value = newVal
			s.lastSeenPropValue = newVal
			s.cursorPos = start
			s.dragStart = -1
			setCursorPos(start)
			onChange(newVal)
			setDragStart(-1)
		}

		switch event.Key {
		case "left", "ctrl+b":
			setDragStart(-1)
			setIsDragging(false)
			if currentPos > 0 {
				s.cursorPos = currentPos - 1
				setCursorPos(currentPos - 1)
			}
			return
		case "right", "ctrl+f":
			setDragStart(-1)
			setIsDragging(false)
			if currentPos < len(runes) {
				s.cursorPos = currentPos + 1
				setCursorPos(currentPos + 1)
			}
			return
		case "home":
			setDragStart(-1)
			setIsDragging(false)
			s.cursorPos = 0
			setCursorPos(0)
			return
		case "end", "ctrl+e":
			setDragStart(-1)
			setIsDragging(false)
			s.cursorPos = len(runes)
			setCursorPos(len(runes))
			return
		case "ctrl+a":
			setDragStart(0)
			s.cursorPos = len(runes)
			setDragStart(0)
			setCursorPos(len(runes))
			return
		case "ctrl+c":
			selStart, selEnd := getSelection()
			if selStart != -1 {
				utils.WriteClipboard(string(runes[selStart:selEnd]))
			}
			return
		case "ctrl+x":
			selStart, selEnd := getSelection()
			if selStart != -1 {
				utils.WriteClipboard(string(runes[selStart:selEnd]))
				deleteRange(selStart, selEnd)
			}
			return
		case "ctrl+v":
			text, err := utils.ReadClipboard()
			if err == nil && len(text) > 0 {
				selStart, selEnd := getSelection()
				if selStart != -1 {
					newVal := string(runes[:selStart]) + text + string(runes[selEnd:])
					newPos := selStart + utf8.RuneCountInString(text)
					s.value = newVal
					s.lastSeenPropValue = newVal
					s.cursorPos = newPos
					setCursorPos(newPos)
					onChange(newVal)
					setDragStart(-1)
				} else {
					newVal := string(runes[:currentPos]) + text + string(runes[currentPos:])
					newPos := currentPos + utf8.RuneCountInString(text)
					s.value = newVal
					s.lastSeenPropValue = newVal
					s.cursorPos = newPos
					setCursorPos(newPos)
					onChange(newVal)
				}
			}
			return
		case "backspace", "\x7f", "\b", "ctrl+h":
			selStart, selEnd := getSelection()
			if selStart != -1 {
				deleteRange(selStart, selEnd)
				return
			}
			if currentPos > 0 {
				newVal := string(runes[:currentPos-1]) + string(runes[currentPos:])
				newPos := currentPos - 1
				s.value = newVal
				s.lastSeenPropValue = newVal
				s.cursorPos = newPos
				setCursorPos(newPos)
				onChange(newVal)
			}
			return
		case "delete":
			selStart, selEnd := getSelection()
			if selStart != -1 {
				deleteRange(selStart, selEnd)
				return
			}
			if currentPos < len(runes) {
				newVal := string(runes[:currentPos]) + string(runes[currentPos+1:])
				s.value = newVal
				s.lastSeenPropValue = newVal
				onChange(newVal)
				setCursorPos(currentPos) // Just trigger render
			}
			return
		}

		if event.Escape || event.Ctrl || event.Meta || event.Key == "tab" || event.Key == "enter" {
			return
		}

		if utf8.RuneCountInString(event.Key) == 1 {
			r, _ := utf8.DecodeRuneInString(event.Key)
			if r >= 32 {
				selStart, selEnd := getSelection()
				activeValue := s.value
				activePos := s.cursorPos

				var activeRunes []rune
				if selStart != -1 {
					activeRunes = append([]rune(activeValue)[:selStart], []rune(activeValue)[selEnd:]...)
					activePos = selStart
				} else {
					activeRunes = []rune(activeValue)
				}

				newVal := string(activeRunes[:activePos]) + event.Key + string(activeRunes[activePos:])
				newPos := activePos + 1

				s.value = newVal
				s.lastSeenPropValue = newVal
				s.cursorPos = newPos

				setCursorPos(newPos)
				onChange(newVal)
				setDragStart(-1)
			}
		}
	}, []any{focused})

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
		if !showPlaceholder && dragStart != -1 && dragStart != renderCursorPos {
			startSel, endSel := dragStart, renderCursorPos
			if startSel > endSel {
				startSel, endSel = endSel, startSel
			}
			if i >= startSel && i < endSel {
				isSelected = true
			}
		}
		isCursor := showCursor && focused && i == renderCursorPos
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

	if showCursor && focused && !showPlaceholder && renderCursorPos == len(renderRunes) && cvw < contentWidth {
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
