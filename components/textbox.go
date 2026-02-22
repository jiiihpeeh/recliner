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

type textBoxInternalState struct {
	value      string
	cursorLine int
	cursorCol  int
	dragStart  int
	dragEnd    int
	isDragging bool
	scrollTop  int
	focused    bool
	onChange   func(string)
}

func TextBox(props any) vdom.Node {
	hc := hooks.GetContext()

	value, _ := util.GetProp[string](props, "value")
	onChange, _ := util.GetProp[func(string)](props, "onChange")
	_, _ = util.GetProp[string](props, "placeholder")
	width, ok := util.GetProp[int](props, "width")
	if !ok {
		width = 40
	}
	height, ok := util.GetProp[int](props, "height")
	if !ok {
		height = 5
	}
	id, _ := util.GetProp[string](props, "id")
	autoFocus, _ := util.GetProp[bool](props, "autoFocus")
	readOnly, _ := util.GetProp[bool](props, "readOnly")
	scrollable, _ := util.GetProp[bool](props, "scrollable")

	focusRes := hc.UseFocus(hooks.FocusOptions{ID: id, AutoFocus: autoFocus})
	focused := focusRes.IsFocused

	cursorLine, setCursorLine := hooks.UseState[int](hc, 0)
	cursorCol, setCursorCol := hooks.UseState[int](hc, 0)
	scrollTop, setScrollTop := hooks.UseState[int](hc, 0)
	showCursor, setShowCursor := hooks.UseState[bool](hc, true)
	dragStartLine, setDragStartLine := hooks.UseState[int](hc, -1)
	dragStartCol, setDragStartCol := hooks.UseState[int](hc, -1)
	isDragging, setIsDragging := hooks.UseState[bool](hc, false)
	isDraggingScroll, setIsDraggingScroll := hooks.UseState[bool](hc, false)

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

	lines := strings.Split(value, "\n")
	if cursorLine >= len(lines) {
		setCursorLine(len(lines) - 1)
		cursorLine = len(lines) - 1
	}
	if cursorLine < 0 {
		cursorLine = 0
	}
	runes := []rune(lines[cursorLine])
	if cursorCol > len(runes) {
		setCursorCol(len(runes))
		cursorCol = len(runes)
	}

	stateRef := hooks.UseRef(hc, &textBoxInternalState{})
	stateRef.Value.value = value
	stateRef.Value.cursorLine = cursorLine
	stateRef.Value.cursorCol = cursorCol
	stateRef.Value.scrollTop = scrollTop
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

		if readOnly {
			return
		}

		// Calculate click line and col
		clickRelY := e.RelY - 1
		clickLineIdx := clickRelY + s.scrollTop
		if clickLineIdx < 0 {
			clickLineIdx = 0
		}
		if clickLineIdx >= len(lines) {
			clickLineIdx = len(lines) - 1
		}

		clickRelX := e.RelX - 1
		clickRunes := []rune(lines[clickLineIdx])
		currentVisualX, clickColIdx := 0, len(clickRunes)
		for i, r := range clickRunes {
			rw := runewidth.RuneWidth(r)
			if rw == 0 {
				rw = 1
			}
			if clickRelX >= currentVisualX && clickRelX < currentVisualX+rw {
				clickColIdx = i
				break
			}
			currentVisualX += rw
		}

		switch e.Action {
		case events.MouseActionPress:
			setCursorLine(clickLineIdx)
			setCursorCol(clickColIdx)
			setDragStartLine(clickLineIdx)
			setDragStartCol(clickColIdx)
			setIsDragging(true)
		case events.MouseActionMotion:
			if isDragging {
				setCursorLine(clickLineIdx)
				setCursorCol(clickColIdx)
			}
		case events.MouseActionRelease:
			setIsDragging(false)
		}
	}

	hc.UseInput(func(event events.KeyPressEvent) {
		if !focused || readOnly {
			return
		}
		setShowCursor(true)
		if onChange == nil {
			return
		}

		currentLines := strings.Split(value, "\n")
		cl, cc := cursorLine, cursorCol

		getSelection := func() (int, int, int, int) {
			if dragStartLine == -1 {
				return -1, -1, -1, -1
			}
			sl, sc, el, ec := dragStartLine, dragStartCol, cl, cc
			if sl > el || (sl == el && sc > ec) {
				sl, sc, el, ec = el, ec, sl, sc
			}
			if sl == el && sc == ec {
				return -1, -1, -1, -1
			}
			return sl, sc, el, ec
		}

		deleteSelection := func() {
			sl, sc, el, ec := getSelection()
			if sl == -1 {
				return
			}
			startLine := []rune(currentLines[sl])
			endLine := []rune(currentLines[el])
			newFirstPart := string(startLine[:sc])
			newLastPart := string(endLine[ec:])
			newLines := append([]string{}, currentLines[:sl]...)
			newLines = append(newLines, newFirstPart+newLastPart)
			newLines = append(newLines, currentLines[el+1:]...)
			onChange(strings.Join(newLines, "\n"))
			setCursorLine(sl)
			setCursorCol(sc)
			setDragStartLine(-1)
		}

		switch event.Key {
		case "up":
			setDragStartLine(-1)
			if cl > 0 {
				setCursorLine(cl - 1)
				prevRunes := []rune(currentLines[cl-1])
				if cc > len(prevRunes) {
					setCursorCol(len(prevRunes))
				}
				if cl-1 < scrollTop {
					setScrollTop(cl - 1)
				}
			}
		case "down":
			setDragStartLine(-1)
			if cl < len(currentLines)-1 {
				setCursorLine(cl + 1)
				nextRunes := []rune(currentLines[cl+1])
				if cc > len(nextRunes) {
					setCursorCol(len(nextRunes))
				}
				if cl+1 >= scrollTop+height {
					setScrollTop(cl + 1 - height + 1)
				}
			}
		case "left":
			setDragStartLine(-1)
			if cc > 0 {
				setCursorCol(cc - 1)
			} else if cl > 0 {
				setCursorLine(cl - 1)
				setCursorCol(len([]rune(currentLines[cl-1])))
			}
		case "right":
			setDragStartLine(-1)
			if cc < len([]rune(currentLines[cl])) {
				setCursorCol(cc + 1)
			} else if cl < len(currentLines)-1 {
				setCursorLine(cl + 1)
				setCursorCol(0)
			}
		case "enter":
			deleteSelection()
			lineRunes := []rune(currentLines[cl])
			newCurrent := string(lineRunes[:cc])
			newNext := string(lineRunes[cc:])
			newLines := append([]string{}, currentLines[:cl]...)
			newLines = append(newLines, newCurrent, newNext)
			newLines = append(newLines, currentLines[cl+1:]...)
			onChange(strings.Join(newLines, "\n"))
			setCursorLine(cl + 1)
			setCursorCol(0)
		case "backspace", "\x7f", "ctrl+h":
			sl, sc, el, ec := getSelection()
			if sl != -1 {
				deleteSelection()
				_ = sc
				_ = el
				_ = ec
				return
			}
			if cc > 0 {
				lineRunes := []rune(currentLines[cl])
				newVal := string(lineRunes[:cc-1]) + string(lineRunes[cc:])
				currentLines[cl] = newVal
				onChange(strings.Join(currentLines, "\n"))
				setCursorCol(cc - 1)
			} else if cl > 0 {
				prevLineRunes := []rune(currentLines[cl-1])
				currentLineRunes := []rune(currentLines[cl])
				newCursorCol := len(prevLineRunes)
				currentLines[cl-1] = string(prevLineRunes) + string(currentLineRunes)
				newLines := append(currentLines[:cl], currentLines[cl+1:]...)
				onChange(strings.Join(newLines, "\n"))
				setCursorLine(cl - 1)
				setCursorCol(newCursorCol)
			}
		case "delete":
			sl, sc, el, ec := getSelection()
			if sl != -1 {
				deleteSelection()
				_ = sc
				_ = el
				_ = ec
				return
			}
			lineRunes := []rune(currentLines[cl])
			if cc < len(lineRunes) {
				newVal := string(lineRunes[:cc]) + string(lineRunes[cc+1:])
				currentLines[cl] = newVal
				onChange(strings.Join(currentLines, "\n"))
			} else if cl < len(currentLines)-1 {
				nextLineRunes := []rune(currentLines[cl+1])
				currentLines[cl] = string(lineRunes) + string(nextLineRunes)
				newLines := append(currentLines[:cl+1], currentLines[cl+2:]...)
				onChange(strings.Join(newLines, "\n"))
			}
		case "ctrl+a":
			setDragStartLine(0)
			setDragStartCol(0)
			setCursorLine(len(currentLines) - 1)
			setCursorCol(len([]rune(currentLines[len(currentLines)-1])))
		case "ctrl+c":
			sl, sc, el, ec := getSelection()
			if sl != -1 {
				if sl == el {
					utils.WriteClipboard(string([]rune(currentLines[sl])[sc:ec]))
				} else {
					var selected []string
					selected = append(selected, string([]rune(currentLines[sl])[sc:]))
					selected = append(selected, currentLines[sl+1:el]...)
					selected = append(selected, string([]rune(currentLines[el])[:ec]))
					utils.WriteClipboard(strings.Join(selected, "\n"))
				}
			}
		case "ctrl+x":
			sl, sc, el, ec := getSelection()
			if sl != -1 {
				if sl == el {
					utils.WriteClipboard(string([]rune(currentLines[sl])[sc:ec]))
				} else {
					var selected []string
					selected = append(selected, string([]rune(currentLines[sl])[sc:]))
					selected = append(selected, currentLines[sl+1:el]...)
					selected = append(selected, string([]rune(currentLines[el])[:ec]))
					utils.WriteClipboard(strings.Join(selected, "\n"))
				}
				deleteSelection()
			}
		case "ctrl+v":
			text, err := utils.ReadClipboard()
			if err == nil && len(text) > 0 {
				deleteSelection()
				currentLines := strings.Split(stateRef.Value.value, "\n")
				cl, cc := stateRef.Value.cursorLine, stateRef.Value.cursorCol
				lineRunes := []rune(currentLines[cl])
				pastedLines := strings.Split(text, "\n")
				if len(pastedLines) == 1 {
					newVal := string(lineRunes[:cc]) + text + string(lineRunes[cc:])
					currentLines[cl] = newVal
					onChange(strings.Join(currentLines, "\n"))
					setCursorCol(cc + utf8.RuneCountInString(text))
				} else {
					firstPart := string(lineRunes[:cc]) + pastedLines[0]
					lastPart := pastedLines[len(pastedLines)-1] + string(lineRunes[cc:])
					newLines := append([]string{}, currentLines[:cl]...)
					newLines = append(newLines, firstPart)
					if len(pastedLines) > 2 {
						newLines = append(newLines, pastedLines[1:len(pastedLines)-1]...)
					}
					newLines = append(newLines, lastPart)
					newLines = append(newLines, currentLines[cl+1:]...)
					onChange(strings.Join(newLines, "\n"))
					setCursorLine(cl + len(pastedLines) - 1)
					setCursorCol(utf8.RuneCountInString(pastedLines[len(pastedLines)-1]))
				}
			}
		default:
			if utf8.RuneCountInString(event.Key) == 1 && !event.Ctrl && !event.Meta && event.Key != "escape" {
				r, _ := utf8.DecodeRuneInString(event.Key)
				if r >= 32 {
					deleteSelection()
					currentLines := strings.Split(stateRef.Value.value, "\n")
					cl, cc := stateRef.Value.cursorLine, stateRef.Value.cursorCol
					lineRunes := []rune(currentLines[cl])
					newVal := string(lineRunes[:cc]) + event.Key + string(lineRunes[cc:])
					currentLines[cl] = newVal
					onChange(strings.Join(currentLines, "\n"))
					setCursorCol(cc + 1)
				}
			}
		}
	}, []any{value, cursorLine, cursorCol, scrollTop, focused, dragStartLine, dragStartCol})

	contentChildren := []vdom.Node{}
	visibleLines := height
	if scrollable {
		visibleLines = height
	}

	sl, sc, el, ec := -1, -1, -1, -1
	if dragStartLine != -1 {
		sl, sc, el, ec = dragStartLine, dragStartCol, cursorLine, cursorCol
		if sl > el || (sl == el && sc > ec) {
			sl, sc, el, ec = el, ec, sl, sc
		}
	}

	currentTextWidth := width - 2
	if scrollable {
		currentTextWidth -= 1
	}

	for i := 0; i < visibleLines; i++ {
		actualLineIdx := i + scrollTop
		if actualLineIdx >= len(lines) {
			break
		}
		line := lines[actualLineIdx]
		runes := []rune(line)
		var lineChildren []vdom.Node
		cvw := 0
		for charIdx, r := range runes {
			rw := runewidth.RuneWidth(r)
			if rw == 0 {
				rw = 1
			}
			if cvw+rw > currentTextWidth {
				break
			}
			isSelected := false
			if sl != -1 {
				if actualLineIdx > sl && actualLineIdx < el {
					isSelected = true
				} else if actualLineIdx == sl && actualLineIdx == el {
					if charIdx >= sc && charIdx < ec {
						isSelected = true
					}
				} else if actualLineIdx == sl {
					if charIdx >= sc {
						isSelected = true
					}
				} else if actualLineIdx == el {
					if charIdx < ec {
						isSelected = true
					}
				}
			}
			charStyle := vdom.Style{Display: "inline"}
			if isSelected {
				charStyle.Reverse = true
			}
			if focused && showCursor && actualLineIdx == cursorLine && charIdx == cursorCol {
				charStyle.Reverse, charStyle.Underline = true, true
			}
			lineChildren = append(lineChildren, &vdom.Element{
				Type: "text",
				Props: struct {
					Style vdom.Style
				}{
					Style: charStyle,
				},
				InnerText: string(r),
				Style:     charStyle,
			})
			cvw += rw
		}
		if focused && showCursor && actualLineIdx == cursorLine && cursorCol == len(runes) && cvw < currentTextWidth {
			rsStyle := vdom.Style{Reverse: true, Display: "inline"}
			lineChildren = append(lineChildren, &vdom.Element{
				Type: "text",
				Props: struct {
					Style vdom.Style
				}{
					Style: rsStyle,
				},
				InnerText: " ",
				Style:     rsStyle,
			})
		}
		boxS := vdom.Style{Display: "flex", FlexDirection: "row", Height: 1}
		contentChildren = append(contentChildren, &vdom.Element{
			Type: "box",
			Props: struct {
				BorderStyle string
				Style       vdom.Style
			}{
				BorderStyle: "none",
				Style:       boxS,
			},
			Children: lineChildren,
			Style:    boxS,
		})
	}

	if scrollable {
		total := len(lines)
		barLength := height
		thumbSize := (barLength * barLength) / total
		if thumbSize < 1 {
			thumbSize = 1
		}
		maxPos := barLength - thumbSize
		maxScroll := total - height
		pos := 0
		if maxScroll > 0 && maxPos > 0 {
			pos = (scrollTop * maxPos) / maxScroll
		}

		sbProps := struct {
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
				if isDraggingScroll {
					return "yellow"
				}
				return "white"
			}(),
			TrackColor: "gray",
			OnClick: func(me events.MouseEvent) {
				if me.Button != events.MouseButtonLeft {
					return
				}
				desiredPos := me.RelY - (thumbSize / 2)
				if desiredPos < 0 {
					desiredPos = 0
				}
				if desiredPos > maxPos {
					desiredPos = maxPos
				}
				ns := 0
				if maxPos > 0 && maxScroll > 0 {
					ns = (desiredPos * maxScroll) / maxPos
				}
				switch me.Action {
				case events.MouseActionPress:
					setIsDraggingScroll(true)
					setScrollTop(ns)
				case events.MouseActionMotion:
					setScrollTop(ns)
				case events.MouseActionRelease:
					setScrollTop(ns)
					setIsDraggingScroll(false)
				}
			},
		}

		rowStyle := vdom.Style{Display: "flex", FlexDirection: "row"}
		borderSpace := 1
		innerWidth := width - (borderSpace * 2)
		cElemStyle := vdom.Style{Width: innerWidth - 1, Height: height + 2}
		contentElem := &vdom.Element{
			Type: "box",
			Props: struct {
				Padding     int
				BorderStyle string
				Style       vdom.Style
				OnClick     func(events.MouseEvent)
			}{
				Padding:     1,
				BorderStyle: "none",
				Style:       cElemStyle,
				OnClick:     handleMouse,
			},
			Children: contentChildren,
			Style:    cElemStyle,
		}
		rootStyle := vdom.Style{Width: width, Height: height + 4}
		return &vdom.Element{
			Type: "box",
			Props: struct {
				BorderStyle string
				BorderColor string
				Style       vdom.Style
			}{
				BorderStyle: "single",
				BorderColor: "blue",
				Style:       rootStyle,
			},
			Children: []vdom.Node{&vdom.Element{
				Type: "box",
				Props: struct {
					Style       vdom.Style
					BorderStyle string
				}{
					Style:       rowStyle,
					BorderStyle: "none",
				},
				Children: []vdom.Node{contentElem, ScrollBar(sbProps)},
				Style:    rowStyle,
			}},
			Style: rootStyle,
		}
	}

	rootStyle := vdom.Style{Width: width, Height: height + 4}
	return &vdom.Element{
		Type: "box",
		Props: struct {
			BorderStyle string
			BorderColor string
			Padding     int
			Style       vdom.Style
			OnClick     func(events.MouseEvent)
		}{
			BorderStyle: "single",
			BorderColor: "blue",
			Padding:     1,
			Style:       rootStyle,
			OnClick:     handleMouse,
		},
		Children: contentChildren,
		Style:    rootStyle,
	}
}
