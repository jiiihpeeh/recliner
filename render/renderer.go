package render

import (
	"fmt"
	"image"
	"os"
	"strings"
	"sync"

	"github.com/j-p/recliner/events"
	"github.com/j-p/recliner/util"
	"github.com/j-p/recliner/vdom"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// --- Types ---

type Layout struct{ X, Y, Width, Height int }

type HitArea struct {
	XStart, XEnd, YStart, YEnd int
	Fixed                      bool
	Handler                    func(events.MouseEvent)
}

type measureKey struct {
	node     vdom.Node
	maxWidth int
}

type drawTask struct {
	node        vdom.Node
	layout      Layout
	clip        Layout
	inheritedBG string
	zIndex      int
	order       int
}

// --- Buffer ---

type Cell struct {
	Char  rune
	Style vdom.Style
}

type Buffer struct {
	Cells  [][]Cell
	Width  int
	Height int
}

func NewBuffer(width, height int) *Buffer {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	cells := make([][]Cell, height)
	for i := range cells {
		cells[i] = make([]Cell, width)
		for j := range cells[i] {
			cells[i][j] = Cell{Char: ' '}
		}
	}
	return &Buffer{Cells: cells, Width: width, Height: height}
}

func (b *Buffer) Set(x, y int, char rune, style vdom.Style) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}
	rw := runewidth.RuneWidth(char)
	if rw == 0 {
		return
	}
	if x+rw > b.Width {
		return
	}
	b.Cells[y][x] = Cell{Char: char, Style: style}
	for i := 1; i < rw; i++ {
		if x+i < b.Width {
			b.Cells[y][x+i] = Cell{Char: 0, Style: style}
		}
	}
}

func (b *Buffer) ToString() string {
	var sb strings.Builder
	var lastStyle vdom.Style
	styleActive := false

	for y := 0; y < b.Height; y++ {
		// Explicitly move to start of line
		sb.WriteString(fmt.Sprintf("\x1b[%d;1H", y+1))
		for x := 0; x < b.Width; x++ {
			cell := b.Cells[y][x]
			if cell.Char == 0 && x > 0 {
				continue
			}

			if !styleActive || !cell.Style.Equals(&lastStyle) {
				if styleActive {
					sb.WriteString("\x1b[0m")
				}
				sb.WriteString(getStyleSequence(cell.Style))
				lastStyle = cell.Style
				styleActive = true
			}

			char := cell.Char
			if char == 0 {
				char = ' '
			}
			sb.WriteRune(char)
		}
	}
	if styleActive {
		sb.WriteString("\x1b[0m")
	}
	return sb.String()
}

func getStyleSequence(style vdom.Style) string {
	var codes []string
	if style.Bold {
		codes = append(codes, "1")
	}
	if style.Dim {
		codes = append(codes, "2")
	}
	if style.Italic {
		codes = append(codes, "3")
	}
	if style.Underline {
		codes = append(codes, "4")
	}
	if style.Blink {
		codes = append(codes, "5")
	}
	if style.Reverse {
		codes = append(codes, "7")
	}
	if style.Hidden {
		codes = append(codes, "8")
	}
	if style.Strikethrough {
		codes = append(codes, "9")
	}

	setCol := func(c string, prefix string) {
		if c == "" {
			return
		}
		if strings.HasPrefix(c, "#") {
			rgb := util.HexToRGB(c)
			codes = append(codes, fmt.Sprintf("%s;2;%d;%d;%d", prefix, uint8(rgb.R), uint8(rgb.G), uint8(rgb.B)))
		} else if code, ok := colorCodes[c]; ok && prefix == "38" {
			codes = append(codes, code)
		} else if code, ok := bgColorCodes[c]; ok && prefix == "48" {
			codes = append(codes, code)
		} else {
			rgb := util.NamedToRGB(c)
			codes = append(codes, fmt.Sprintf("%s;2;%d;%d;%d", prefix, uint8(rgb.R), uint8(rgb.G), uint8(rgb.B)))
		}
	}
	setCol(style.Foreground, "38")
	setCol(style.Background, "48")

	if len(codes) == 0 {
		return ""
	}
	return fmt.Sprintf("\x1b[%sm", strings.Join(codes, ";"))
}

var colorCodes = map[string]string{
	"black": "30", "red": "31", "green": "32", "yellow": "33", "blue": "34", "magenta": "35", "cyan": "36", "white": "37",
	"gray": "90", "grey": "90", "redBright": "91", "greenBright": "92", "yellowBright": "93", "blueBright": "94", "magentaBright": "95", "cyanBright": "96", "whiteBright": "97",
}

var bgColorCodes = map[string]string{
	"black": "40", "red": "41", "green": "42", "yellow": "43", "blue": "44", "magenta": "45", "cyan": "46", "white": "47",
	"gray": "100", "grey": "100", "redBright": "101", "greenBright": "102", "yellowBright": "103", "blueBright": "104", "magentaBright": "105", "cyanBright": "106", "whiteBright": "107",
}

// --- Renderer ---

type Renderer struct {
	mu                       sync.Mutex
	width, height            int
	contentHeight            int
	measureCache             map[measureKey]Layout
	hitAreas                 []HitArea
	viewScrollX, viewScrollY int
	lastBuffer               *Buffer
	frameCount               int
}

func NewRenderer() *Renderer {
	r := &Renderer{
		width:        80,
		height:       24,
		measureCache: make(map[measureKey]Layout),
	}
	r.updateTerminalSize()
	return r
}

func (r *Renderer) SetSize(w, h int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.width, r.height = w, h
}

func (r *Renderer) GetSize() (int, int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.width, r.height
}

func (r *Renderer) GetContentHeight() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.contentHeight
}

func (r *Renderer) updateTerminalSize() {
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		r.width, r.height = w, h
	}
}

func (r *Renderer) HitTest(event events.MouseEvent) func(events.MouseEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := len(r.hitAreas) - 1; i >= 0; i-- {
		area := r.hitAreas[i]

		targetX, targetY := event.X, event.Y
		if !area.Fixed {
			targetX += r.viewScrollX
			targetY += r.viewScrollY
		}

		if targetX >= area.XStart && targetX < area.XEnd && targetY >= area.YStart && targetY < area.YEnd {
			return func(evt events.MouseEvent) {
				evt.RelX, evt.RelY = targetX-area.XStart, targetY-area.YStart
				area.Handler(evt)
			}
		}
	}
	return nil
}

func (r *Renderer) SetViewScroll(x, y int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.viewScrollX, r.viewScrollY = x, y
}

func (r *Renderer) Render(node vdom.Node) string {
	// 1. Measure & Layout
	ctx := &RenderContext{
		renderer: r,
		layouts:  make(map[vdom.Node]Layout),

		measureCache: make(map[measureKey]Layout), // Temporary cache for this frame
		width:        r.width,
		height:       r.height,
	}

	ctx.measureNode(node, ctx.width)
	ctx.layoutNode(node, 0, 0)

	// Determine total content size
	maxExtent := 0
	for _, l := range ctx.layouts {
		extent := l.Y + l.Height
		if extent > maxExtent {
			maxExtent = extent
		}
	}

	r.mu.Lock()
	r.contentHeight = maxExtent
	r.frameCount++

	bufW, bufH := r.width, r.height
	r.mu.Unlock()

	// 2. Buffer & Draw
	buffer := NewBuffer(bufW, bufH)
	ctx.collectDrawTasks(node, Layout{0, 0, bufW, bufH}, "", 0, false)

	util.StableSort(ctx.tasks, func(i, j int) bool {
		if ctx.tasks[i].zIndex != ctx.tasks[j].zIndex {
			return ctx.tasks[i].zIndex < ctx.tasks[j].zIndex
		}
		return ctx.tasks[i].order < ctx.tasks[j].order
	})

	for _, task := range ctx.tasks {
		ctx.executeTask(task, buffer)
	}

	r.mu.Lock()
	r.hitAreas = ctx.hitAreas

	var output string
	forceFull := r.frameCount%100 == 0

	if !forceFull && r.lastBuffer != nil && r.lastBuffer.Width == buffer.Width && r.lastBuffer.Height == buffer.Height {
		output = r.computeDiff(r.lastBuffer, buffer)
	} else {
		output = buffer.ToString()
	}

	r.lastBuffer = buffer
	r.mu.Unlock()

	return output
}

func (r *Renderer) computeDiff(oldBuf, newBuf *Buffer) string {
	var sb strings.Builder
	var lastStyle vdom.Style
	styleActive := false
	curX, curY := -1, -1

	for y := 0; y < newBuf.Height; y++ {
		for x := 0; x < newBuf.Width; x++ {
			oldCell := oldBuf.Cells[y][x]
			newCell := newBuf.Cells[y][x]

			if oldCell.Char == newCell.Char && oldCell.Style.Equals(&newCell.Style) {
				continue
			}

			if x != curX || y != curY {
				sb.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, x+1))
			}

			if !styleActive || !newCell.Style.Equals(&lastStyle) {
				if styleActive {
					sb.WriteString("\x1b[0m")
				}
				sb.WriteString(getStyleSequence(newCell.Style))
				lastStyle = newCell.Style
				styleActive = true
			}

			char := newCell.Char
			if char == 0 {
				char = ' '
			}
			sb.WriteRune(char)
			curX = x + runewidth.RuneWidth(char)
			curY = y
		}
	}
	if styleActive {
		sb.WriteString("\x1b[0m")
	}
	return sb.String()
}

// --- RenderContext ---

type RenderContext struct {
	renderer     *Renderer
	layouts      map[vdom.Node]Layout
	measureCache map[measureKey]Layout
	hitAreas     []HitArea
	tasks        []drawTask
	width        int
	height       int
}

func (ctx *RenderContext) measureNode(node vdom.Node, maxWidth int) Layout {
	if node == nil {
		return Layout{}
	}
	key := measureKey{node, maxWidth}
	if cached, ok := ctx.measureCache[key]; ok {
		return cached
	}

	var res Layout
	switch n := node.(type) {
	case *vdom.TextNode:
		res = ctx.measureText(n.Content)
	case *vdom.Element:
		switch n.Type {
		case "text":
			res = ctx.measureText(n.InnerText)
		case "image":
			w, _ := util.GetProp[int](n.Props, "width")
			h, _ := util.GetProp[int](n.Props, "height")
			res = Layout{Width: w, Height: h}
		default:
			res = ctx.measureBox(n, maxWidth)
		}
	case *vdom.Fragment:
		var fl Layout
		for _, child := range n.Children {
			childLayout := ctx.measureNode(child, maxWidth)
			fl.Width = util.Max(fl.Width, childLayout.Width)
			fl.Height += childLayout.Height
		}
		res = fl
	}

	ctx.measureCache[key] = res
	ctx.layouts[node] = res
	return res
}

func (ctx *RenderContext) measureText(content string) Layout {
	lines := strings.Split(content, "\n")
	w := 0
	for _, l := range lines {
		wl := runewidth.StringWidth(l)
		if wl > w {
			w = wl
		}
	}
	return Layout{Width: w, Height: len(lines)}
}

func (ctx *RenderContext) measureBox(n *vdom.Element, maxWidth int) Layout {
	p, _ := util.GetProp[int](n.Props, "padding")
	bs := util.Ternary(util.GetPropString(n.Props, "borderStyle") == "none", 0, 1)

	innerMaxWidth := maxWidth
	if n.Style.Width != 0 {
		innerMaxWidth = n.Style.Width
	}
	innerMaxWidth -= (p * 2) + (bs * 2)
	if innerMaxWidth < 0 {
		innerMaxWidth = 0
	}

	var cl Layout
	flexDir := util.Ternary(n.Style.FlexDirection == "", "row", n.Style.FlexDirection)
	gap := n.Style.Gap

	if n.Style.Display == "flex" {
		cl = ctx.measureFlex(n, innerMaxWidth, flexDir, gap)
	} else {
		for i, child := range n.Children {
			childLayout := ctx.measureNode(child, innerMaxWidth)
			if isAbsolute(child) {
				continue
			}
			if i > 0 {
				cl.Height += gap
			}
			cl.Width = util.Max(cl.Width, childLayout.Width)
			cl.Height += childLayout.Height
		}
	}

	if n.Type == "box" {
		cl.Width += (p * 2) + (bs * 2)
		cl.Height += (p * 2) + (bs * 2)
	}

	if n.Style.Width != 0 {
		cl.Width = n.Style.Width
	}
	if n.Style.Height != 0 {
		cl.Height = n.Style.Height
	}

	return cl
}

func (ctx *RenderContext) measureFlex(n *vdom.Element, maxWidth int, dir string, gap int) Layout {
	var cl Layout
	first := true
	totalGrow, totalShrink, totalBasis := 0, 0, 0
	type flexItem struct {
		node                vdom.Node
		basis, grow, shrink int
	}
	var items []flexItem

	for _, child := range n.Children {
		if isAbsolute(child) {
			ctx.measureNode(child, maxWidth)
			continue
		}
		childStyle := getStyle(child)
		childLayout := ctx.measureNode(child, maxWidth)
		basis := util.Ternary(childStyle.FlexBasis == 0, util.Ternary(dir == "row", childLayout.Width, childLayout.Height), childStyle.FlexBasis)
		grow := childStyle.FlexGrow
		shrink := util.Ternary(childStyle.FlexShrink == 0, 1, childStyle.FlexShrink)

		totalGrow += grow
		totalShrink += shrink
		if !first {
			totalBasis += gap
		}
		totalBasis += basis
		items = append(items, flexItem{child, basis, grow, shrink})
		first = false
	}

	mainSize := maxWidth
	if dir == "column" {
		p, _ := util.GetProp[int](n.Props, "padding")
		bs := util.Ternary(util.GetPropString(n.Props, "borderStyle") == "none", 0, 1)
		mainSize = util.Ternary(n.Style.Height != 0, n.Style.Height-(p*2)-(bs*2), totalBasis)
	}

	remaining := mainSize - totalBasis
	first = true
	for _, item := range items {
		finalSize := item.basis
		if remaining > 0 && totalGrow > 0 {
			finalSize += int(float64(remaining) * float64(item.grow) / float64(totalGrow))
		} else if remaining < 0 && totalShrink > 0 {
			finalSize += int(float64(remaining) * float64(item.shrink) / float64(totalShrink))
		}

		l := ctx.layouts[item.node]
		if dir == "row" {
			l.Width = finalSize
			cl.Width += finalSize
			if !first {
				cl.Width += gap
			}
			cl.Height = util.Max(cl.Height, l.Height)
		} else {
			l.Height = finalSize
			cl.Height += finalSize
			if !first {
				cl.Height += gap
			}
			cl.Width = util.Max(cl.Width, l.Width)
		}
		ctx.layouts[item.node] = l
		first = false
	}
	return cl
}

func (ctx *RenderContext) layoutNode(node vdom.Node, x, y int) {
	if node == nil {
		return
	}
	l := ctx.layouts[node]
	l.X, l.Y = x, y
	ctx.layouts[node] = l

	switch n := node.(type) {
	case *vdom.Element:
		if handler, ok := util.GetProp[func(events.MouseEvent)](n.Props, "onClick"); ok && handler != nil {
			isFixed := n.Style.Position == "fixed"
			ctx.hitAreas = append(ctx.hitAreas, HitArea{x, x + l.Width, y, y + l.Height, isFixed, handler})
		}
		p, _ := util.GetProp[int](n.Props, "padding")
		bs := util.Ternary(util.GetPropString(n.Props, "borderStyle") == "none", 0, 1)
		st, _ := util.GetProp[int](n.Props, "scrollTop")
		sl, _ := util.GetProp[int](n.Props, "scrollLeft")

		cx, cy := x+bs+p-sl, y+bs+p-st
		curX, curY := cx, cy

		dir := "column"
		if n.Style.Display == "flex" {
			dir = util.Ternary(n.Style.FlexDirection == "", "row", n.Style.FlexDirection)
		}
		gap := n.Style.Gap

		if n.Style.Display == "flex" {
			curX, curY = ctx.applyJustification(n, l, bs, p, cx, cy, dir, gap)
		}

		for _, child := range n.Children {
			childLayout, childStyle := ctx.layouts[child], getStyle(child)
			if childStyle.Position == "absolute" {
				ctx.layoutNode(child, x+childStyle.Left, y+childStyle.Top)
				continue
			}
			if childStyle.Position == "fixed" {
				ctx.layoutNode(child, ctx.renderer.viewScrollX+childStyle.Left, ctx.renderer.viewScrollY+childStyle.Top)
				continue
			}

			tx, ty := curX, curY
			if n.Style.Display == "flex" {
				tx, ty = ctx.applyAlignment(n, childLayout, childStyle, cx, cy, tx, ty, dir, bs, p, l)
			}
			ctx.layoutNode(child, tx, ty)
			if dir == "row" {
				curX += childLayout.Width + gap
			} else {
				curY += childLayout.Height + gap
			}
		}
	case *vdom.Fragment:
		curY := y
		for _, child := range n.Children {
			cl := ctx.layouts[child]
			ctx.layoutNode(child, x, curY)
			curY += cl.Height
		}
	}
}

func (ctx *RenderContext) applyJustification(n *vdom.Element, l Layout, bs, p int, cx, cy int, dir string, gap int) (int, int) {
	tcw, tch, num := 0, 0, 0
	for _, child := range n.Children {
		if isAbsolute(child) {
			continue
		}
		cl := ctx.layouts[child]
		if dir == "row" {
			if num > 0 {
				tcw += gap
			}
			tcw += cl.Width
		} else {
			if num > 0 {
				tch += gap
			}
			tch += cl.Height
		}
		num++
	}
	curX, curY := cx, cy
	contentW, contentH := l.Width-bs*2-p*2, l.Height-bs*2-p*2
	if dir == "row" {
		switch n.Style.JustifyContent {
		case "center":
			curX = cx + (contentW-tcw)/2
		case "flex-end":
			curX = cx + (contentW - tcw)
		case "space-around":
			if num > 0 {
				actualGap := (contentW - (tcw - gap*(num-1))) / num
				curX = cx + actualGap/2
			}
		case "space-evenly":
			if num > 0 {
				actualGap := (contentW - (tcw - gap*(num-1))) / (num + 1)
				curX = cx + actualGap
			}
		}
	} else {
		switch n.Style.JustifyContent {
		case "center":
			curY = cy + (contentH-tch)/2
		case "flex-end":
			curY = cy + (contentH - tch)
		case "space-around":
			if num > 0 {
				actualGap := (contentH - (tch - gap*(num-1))) / num
				curY = cy + actualGap/2
			}
		case "space-evenly":
			if num > 0 {
				actualGap := (contentH - (tch - gap*(num-1))) / (num + 1)
				curY = cy + actualGap
			}
		}
	}
	return curX, curY
}

func (ctx *RenderContext) applyAlignment(n *vdom.Element, cl Layout, cs vdom.Style, cx, cy int, tx, ty int, dir string, bs, p int, l Layout) (int, int) {
	align := util.Ternary(cs.AlignSelf != "" && cs.AlignSelf != "auto", cs.AlignSelf, n.Style.AlignItems)
	if dir == "row" {
		switch align {
		case "center":
			ty = cy + (l.Height-bs*2-p*2-cl.Height)/2
		case "flex-end":
			ty = cy + (l.Height - bs*2 - p*2 - cl.Height)
		}
	} else {
		switch align {
		case "center":
			tx = cx + (l.Width-bs*2-p*2-cl.Width)/2
		case "flex-end":
			tx = cx + (l.Width - bs*2 - p*2 - cl.Width)
		}
	}
	return tx, ty
}

func (ctx *RenderContext) collectDrawTasks(node vdom.Node, clip Layout, inheritedBG string, zIndex int, isFixed bool) {
	if node == nil {
		return
	}
	layout := ctx.layouts[node]
	order := len(ctx.tasks)

	switch n := node.(type) {
	case *vdom.Element:
		effBG := n.Style.Background
		if effBG == "" {
			effBG = inheritedBG
		}
		localZ := zIndex
		if n.Style.ZIndex != 0 {
			localZ = n.Style.ZIndex
		}

		fixed := isFixed || n.Style.Position == "fixed"

		// Apply global scroll if not fixed
		drawLayout := layout
		if !fixed {
			drawLayout.X -= ctx.renderer.viewScrollX
			drawLayout.Y -= ctx.renderer.viewScrollY
		}

		if n.Type == "text" || n.Type == "image" || n.Type == "box" {
			ctx.tasks = append(ctx.tasks, drawTask{n, drawLayout, clip, inheritedBG, localZ, order})
		}

		childClip := clip
		if n.Type == "box" {
			p, _ := util.GetProp[int](n.Props, "padding")
			bs := util.Ternary(util.GetPropString(n.Props, "borderStyle") == "none", 0, 1)
			// Intersection needs to be in screen space if clip is in screen space
			// layout is in canvas space. drawLayout is in screen space.
			cr := Layout{drawLayout.X + bs + p, drawLayout.Y + bs + p, layout.Width - (bs * 2) - (p * 2), layout.Height - (bs * 2) - (p * 2)}
			childClip, _ = intersect(cr, clip)
		}

		for _, child := range n.Children {
			cc := childClip
			cf := fixed
			if ce, ok := child.(*vdom.Element); ok && ce.Style.Position == "fixed" {
				cc = Layout{0, 0, ctx.renderer.width, ctx.renderer.height}
				cf = true
			}
			ctx.collectDrawTasks(child, cc, effBG, localZ, cf)
		}
	case *vdom.TextNode:
		drawLayout := layout
		if !isFixed {
			drawLayout.X -= ctx.renderer.viewScrollX
			drawLayout.Y -= ctx.renderer.viewScrollY
		}
		ctx.tasks = append(ctx.tasks, drawTask{n, drawLayout, clip, inheritedBG, zIndex, order})
	case *vdom.Fragment:
		for _, child := range n.Children {
			ctx.collectDrawTasks(child, clip, inheritedBG, zIndex, isFixed)
		}
	}
}

func (ctx *RenderContext) executeTask(task drawTask, buf *Buffer) {
	switch n := task.node.(type) {
	case *vdom.TextNode:
		ctx.drawTextNode(n, task.layout.X, task.layout.Y, buf, task.clip, task.inheritedBG)
	case *vdom.Element:
		switch n.Type {
		case "text":
			ctx.drawTextNode(&vdom.TextNode{Content: n.InnerText, Style: n.Style}, task.layout.X, task.layout.Y, buf, task.clip, task.inheritedBG)
		case "image":
			ctx.drawImage(n, task.layout, buf, task.clip)
		case "box":
			ctx.drawBox(n, task.layout, buf, task.clip, task.inheritedBG)
		}
	}
}

func (ctx *RenderContext) drawTextNode(text *vdom.TextNode, x, y int, buf *Buffer, clip Layout, inheritedBG string) {
	lines := strings.Split(text.Content, "\n")
	for i, line := range lines {
		yPos := y + i
		if yPos < clip.Y || yPos >= clip.Y+clip.Height {
			continue
		}
		runes := []rune(line)
		curX := x
		for charIdx, ch := range runes {
			rw := runewidth.RuneWidth(ch)
			if rw == 0 {
				continue
			}
			if curX+rw <= clip.X {
				curX += rw
				continue
			}
			if curX >= clip.X+clip.Width {
				break
			}
			s := text.Style
			if s.Background == "" {
				s.Background = inheritedBG
			}
			if g := s.ForegroundGradient; g != nil {
				s.Foreground = ctx.calcGradient(g, charIdx, len(runes), i, len(lines))
			}
			buf.Set(curX, yPos, ch, s)
			curX += rw
		}
	}
}

func (ctx *RenderContext) calcGradient(g *vdom.LinearGradient, charIdx, totalChars, lineIdx, totalLines int) string {
	from, to := util.HexToRGB(g.From), util.HexToRGB(g.To)
	t := 0.0
	if g.Direction == "vertical" {
		if totalLines > 1 {
			t = float64(lineIdx) / float64(totalLines-1)
		}
	} else {
		if totalChars > 1 {
			t = float64(charIdx) / float64(totalChars-1)
		}
	}
	return util.LerpRGB(from, to, t).ToHex()
}

func (ctx *RenderContext) drawImage(el *vdom.Element, layout Layout, buf *Buffer, clip Layout) {
	img, ok := util.GetProp[image.Image](el.Props, "image")
	if !ok {
		return
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h/2; y++ {
		for x := range w {
			wx, wy := layout.X+x, layout.Y+y
			if wx < clip.X || wx >= clip.X+clip.Width || wy < clip.Y || wy >= clip.Y+clip.Height {
				continue
			}
			c1, c2 := img.At(b.Min.X+x, b.Min.Y+y*2), img.At(b.Min.X+x, b.Min.Y+y*2+1)
			r1, g1, b1, _ := c1.RGBA()
			r2, g2, b2, _ := c2.RGBA()
			s := vdom.Style{
				Foreground: fmt.Sprintf("#%02x%02x%02x", r1>>8, g1>>8, b1>>8),
				Background: fmt.Sprintf("#%02x%02x%02x", r2>>8, g2>>8, b2>>8),
			}
			buf.Set(wx, wy, '▀', s)
		}
	}
}

func (ctx *RenderContext) drawBox(el *vdom.Element, layout Layout, buf *Buffer, clip Layout, inheritedBG string) {
	style, w, h, x, y := el.Style, layout.Width, layout.Height, layout.X, layout.Y
	bs := util.GetPropString(el.Props, "borderStyle")
	if bs == "" {
		bs = "single"
	}
	bc := util.GetPropString(el.Props, "borderColor")

	if bs != "none" {
		for i := 0; i < w; i++ {
			ctx.drawBorderCell(x+i, y, getBorderChar(bs, "top", i, w), el, layout, buf, clip, bc, inheritedBG)
			ctx.drawBorderCell(x+i, y+h-1, getBorderChar(bs, "bottom", i, w), el, layout, buf, clip, bc, inheritedBG)
		}
		for i := 1; i < h-1; i++ {
			ctx.drawBorderCell(x, y+i, getBorderChar(bs, "left", i, h), el, layout, buf, clip, bc, inheritedBG)
			ctx.drawBorderCell(x+w-1, y+i, getBorderChar(bs, "right", i, h), el, layout, buf, clip, bc, inheritedBG)
		}
	}

	if style.Background != "" {
		fillS := util.Ternary(bs == "none", 0, 1)
		for iy := fillS; iy < h-fillS; iy++ {
			yp := y + iy
			if yp < clip.Y || yp >= clip.Y+clip.Height {
				continue
			}
			for ix := fillS; ix < w-fillS; ix++ {
				xp := x + ix
				if xp < clip.X || xp >= clip.X+clip.Width {
					continue
				}
				buf.Set(xp, yp, ' ', style)
			}
		}
	}
}

func (ctx *RenderContext) drawBorderCell(x, y int, char rune, el *vdom.Element, layout Layout, buf *Buffer, clip Layout, color, bg string) {
	if x < clip.X || x >= clip.X+clip.Width || y < clip.Y || y >= clip.Y+clip.Height {
		return
	}
	s := el.Style
	s.Foreground, s.Background = color, bg
	if g := s.BorderGradient; g != nil {
		from, to := util.HexToRGB(g.From), util.HexToRGB(g.To)
		cx, cy := float64(layout.X)+float64(layout.Width)*g.CenterX, float64(layout.Y)+float64(layout.Height)*g.CenterY
		dist := util.Distance(float64(x), float64(y), cx, cy)
		maxDist := g.Radius * (float64(layout.Width+layout.Height) / 4.0)
		t := util.MinF(1.0, dist/maxDist)
		s.Foreground = util.LerpRGB(from, to, t).ToHex()
	}
	buf.Set(x, y, char, s)
}

func getBorderChar(style, side string, idx, total int) rune {
	switch style {
	case "round":
		if side == "top" {
			if idx == 0 {
				return '╭'
			}
			if idx == total-1 {
				return '╮'
			}
			return '─'
		}
		if side == "bottom" {
			if idx == 0 {
				return '╰'
			}
			if idx == total-1 {
				return '╯'
			}
			return '─'
		}
		return '│'
	case "double":
		if side == "top" {
			if idx == 0 {
				return '╔'
			}
			if idx == total-1 {
				return '╗'
			}
			return '═'
		}
		if side == "bottom" {
			if idx == 0 {
				return '╚'
			}
			if idx == total-1 {
				return '╝'
			}
			return '═'
		}
		return '║'
	case "classic":
		if side == "top" || side == "bottom" {
			if idx == 0 || idx == total-1 {
				return '+'
			}
			return '-'
		}
		return '|'
	default:
		if side == "top" {
			if idx == 0 {
				return '┌'
			}
			if idx == total-1 {
				return '┐'
			}
			return '─'
		}
		if side == "bottom" {
			if idx == 0 {
				return '└'
			}
			if idx == total-1 {
				return '┘'
			}
			return '─'
		}
		return '│'
	}
}

// --- Helpers ---

func isAbsolute(n vdom.Node) bool {
	s := getStyle(n)
	return s.Position == "absolute" || s.Position == "fixed"
}

func getStyle(n vdom.Node) vdom.Style {
	switch v := n.(type) {
	case *vdom.Element:
		return v.Style
	case *vdom.TextNode:
		return v.Style
	}
	return vdom.Style{}
}

func intersect(a, b Layout) (res Layout, ok bool) {
	x1, y1 := util.Max(a.X, b.X), util.Max(a.Y, b.Y)
	x2, y2 := util.Min(a.X+a.Width, b.X+b.Width), util.Min(a.Y+a.Height, b.Y+b.Height)
	if x2 <= x1 || y2 <= y1 {
		return Layout{}, false
	}
	return Layout{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}, true
}
