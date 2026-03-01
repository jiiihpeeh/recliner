package vdom

import (
	"reflect"
	"sync"

	"github.com/jiiihpeeh/recliner/util"
)

type NodeType int

const (
	NodeTypeElement NodeType = iota
	NodeTypeText
	NodeTypeFragment
)

type Node interface {
	GetNodeType() NodeType
	Clone() Node
}

type Element struct {
	Type      string
	Props     any
	Children  []Node
	Key       string
	Style     Style
	InnerText string
}

func (e *Element) GetNodeType() NodeType { return NodeTypeElement }
func (e *Element) Clone() Node {
	children := make([]Node, len(e.Children))
	for i, c := range e.Children {
		children[i] = c.Clone()
	}

	return &Element{
		Type:      e.Type,
		Props:     e.Props, // Assuming immutability for non-map props
		Children:  children,
		Key:       e.Key,
		Style:     e.Style.Clone(),
		InnerText: e.InnerText,
	}
}

type TextNode struct {
	Content string
	Style   Style
}

func (t *TextNode) GetNodeType() NodeType { return NodeTypeText }
func (t *TextNode) Clone() Node           { return &TextNode{Content: t.Content, Style: t.Style.Clone()} }

type Fragment struct {
	Children []Node
}

func (f Fragment) GetNodeType() NodeType { return NodeTypeFragment }
func (f Fragment) Clone() Node {
	children := make([]Node, len(f.Children))
	for i, c := range f.Children {
		children[i] = c.Clone()
	}
	return Fragment{Children: children}
}

type LinearGradient struct {
	From      string
	To        string
	Direction string // "horizontal", "vertical"
}

type RadialGradient struct {
	From    string
	To      string
	CenterX float64 // 0.0 - 1.0
	CenterY float64 // 0.0 - 1.0
	Radius  float64
}

type Style struct {
	Bold              bool
	Italic            bool
	Underline         bool
	Strikethrough     bool
	Inverse           bool
	Blink             bool
	Reverse           bool
	Foreground        string
	Background        string
	Dim               bool
	Hidden            bool
	Display           string
	FlexDirection     string
	Gap               int
	ColumnGap         int
	RowGap            int
	MarginTop         int
	MarginBottom      int
	MarginLeft        int
	MarginRight       int
	BorderTopColor    string
	BorderBottomColor string
	BorderLeftColor   string
	BorderRightColor  string
	BorderTopDim      bool
	BorderBottomDim   bool
	BorderLeftDim     bool
	BorderRightDim    bool
	BorderTop         bool
	BorderBottom      bool
	BorderLeft        bool
	BorderRight       bool
	Position          string // "relative", "absolute", "fixed"
	Top               int
	Left              int
	Right             int
	Bottom            int
	ZIndex            int
	AlignItems        string // "flex-start", "center", "flex-end", "stretch"
	AlignSelf         string // "auto", "flex-start", "center", "flex-end", "stretch"
	JustifyContent    string // "flex-start", "center", "flex-end", "space-between", "space-around", "space-evenly"
	FlexGrow          int
	FlexShrink        int
	FlexBasis         int
	Width             int
	Height            int

	ForegroundGradient *LinearGradient
	BorderGradient     *RadialGradient
}

func (s *Style) Clone() Style {
	var fg *LinearGradient
	if s.ForegroundGradient != nil {
		fg = &LinearGradient{From: s.ForegroundGradient.From, To: s.ForegroundGradient.To, Direction: s.ForegroundGradient.Direction}
	}
	var bg *RadialGradient
	if s.BorderGradient != nil {
		bg = &RadialGradient{From: s.BorderGradient.From, To: s.BorderGradient.To, CenterX: s.BorderGradient.CenterX, CenterY: s.BorderGradient.CenterY, Radius: s.BorderGradient.Radius}
	}

	return Style{
		Bold:               s.Bold,
		Italic:             s.Italic,
		Underline:          s.Underline,
		Strikethrough:      s.Strikethrough,
		Inverse:            s.Inverse,
		Blink:              s.Blink,
		Reverse:            s.Reverse,
		Foreground:         s.Foreground,
		Background:         s.Background,
		Dim:                s.Dim,
		Hidden:             s.Hidden,
		Display:            s.Display,
		FlexDirection:      s.FlexDirection,
		Gap:                s.Gap,
		Position:           s.Position,
		Top:                s.Top,
		Left:               s.Left,
		Right:              s.Right,
		Bottom:             s.Bottom,
		ZIndex:             s.ZIndex,
		AlignItems:         s.AlignItems,
		AlignSelf:          s.AlignSelf,
		JustifyContent:     s.JustifyContent,
		FlexGrow:           s.FlexGrow,
		FlexShrink:         s.FlexShrink,
		FlexBasis:          s.FlexBasis,
		Width:              s.Width,
		Height:             s.Height,
		ColumnGap:          s.ColumnGap,
		RowGap:             s.RowGap,
		MarginTop:          s.MarginTop,
		MarginBottom:       s.MarginBottom,
		MarginLeft:         s.MarginLeft,
		MarginRight:        s.MarginRight,
		BorderTopColor:     s.BorderTopColor,
		BorderBottomColor:  s.BorderBottomColor,
		BorderLeftColor:    s.BorderLeftColor,
		BorderRightColor:   s.BorderRightColor,
		BorderTopDim:       s.BorderTopDim,
		BorderBottomDim:    s.BorderBottomDim,
		BorderLeftDim:      s.BorderLeftDim,
		BorderRightDim:     s.BorderRightDim,
		BorderTop:          s.BorderTop,
		BorderBottom:       s.BorderBottom,
		BorderLeft:         s.BorderLeft,
		BorderRight:        s.BorderRight,
		ForegroundGradient: fg,
		BorderGradient:     bg,
	}
}

type ComponentFunc func(props any) Node

var componentRegistry = make(map[string]ComponentFunc)

func RegisterComponent(name string, fn ComponentFunc) {
	componentRegistry[name] = fn
}

func GetComponent(name string) (ComponentFunc, bool) {
	fn, ok := componentRegistry[name]
	return fn, ok
}

var (
	elementPool = sync.Pool{
		New: func() any {
			return &Element{
				Children: make([]Node, 0, 4),
			}
		},
	}
	textNodePool = sync.Pool{
		New: func() any {
			return &TextNode{}
		},
	}
)

func (e *Element) Release() {
	for _, child := range e.Children {
		if releasable, ok := child.(interface{ Release() }); ok {
			releasable.Release()
		}
	}
	// Clear for reuse
	e.Props = nil
	e.Children = e.Children[:0]
	e.Type = ""
	e.Key = ""
	e.InnerText = ""
	e.Style = Style{}
	elementPool.Put(e)
}

func (t *TextNode) Release() {
	t.Content = ""
	t.Style = Style{}
	textNodePool.Put(t)
}

type ComponentProps struct {
	Props    any
	Children []Node
}

func CreateElement(typ string, props any, children ...any) Node {
	var childNodes []Node
	for _, c := range children {
		if c == nil {
			continue
		}
		switch v := c.(type) {
		case string:
			text := textNodePool.Get().(*TextNode)
			text.Content = v
			childNodes = append(childNodes, text)
		case Node:
			childNodes = append(childNodes, v)
		case []Node:
			childNodes = append(childNodes, v...)
		case []any:
			for _, item := range v {
				if item == nil {
					continue
				}
				if s, ok := item.(string); ok {
					text := textNodePool.Get().(*TextNode)
					text.Content = s
					childNodes = append(childNodes, text)
				} else if n, ok := item.(Node); ok {
					childNodes = append(childNodes, n)
				}
			}
		}
	}

	if fn, ok := GetComponent(typ); ok {
		return fn(ComponentProps{
			Props:    props,
			Children: flattenChildren(childNodes),
		})
	}

	if typ == "text" || typ == "" {
		var textNode *TextNode
		if len(childNodes) > 0 {
			if text, ok := childNodes[0].(*TextNode); ok {
				cloned := textNodePool.Get().(*TextNode)
				*cloned = *text
				textNode = cloned
			}
		}
		if textNode == nil {
			textNode = textNodePool.Get().(*TextNode)
			textNode.Content = getTextContent(childNodes)
		}

		if props != nil {
			if styleAny, ok := util.GetProp[any](props, "style"); ok {
				textNode.Style = ParseStyle(styleAny)
			}
		}
		return textNode
	}

	if typ == "fragment" {
		return Fragment{Children: flattenChildren(childNodes)}
	}

	elem := elementPool.Get().(*Element)
	elem.Type = typ
	elem.Props = props
	elem.Children = flattenChildren(childNodes)
	elem.Key = ""
	elem.Style = Style{}

	if props != nil {
		if styleAny, ok := util.GetProp[any](props, "style"); ok {
			elem.Style = ParseStyle(styleAny)
		}
		if key, ok := util.GetProp[string](props, "key"); ok {
			elem.Key = key
		}
		if text, ok := util.GetProp[string](props, "children"); ok {
			elem.InnerText = text
		}
	}

	return elem
}

func getTextContent(children []Node) string {
	if len(children) == 0 {
		return ""
	}
	if text, ok := children[0].(*TextNode); ok {
		return text.Content
	}
	return ""
}

func flattenChildren(children []Node) []Node {
	var result []Node
	for _, child := range children {
		if child == nil {
			continue
		}
		if frag, ok := child.(*Fragment); ok {
			result = append(result, flattenChildren(frag.Children)...)
		} else {
			result = append(result, child)
		}
	}
	return result
}

func ParseStyle(style any) Style {
	s := Style{}
	if style == nil {
		return s
	}

	if sm, ok := style.(Style); ok {
		return sm
	}
	if sm, ok := style.(*Style); ok {
		if sm != nil {
			return *sm
		}
		return s
	}

	// Helper to get from map or struct
	getBool := func(key string) bool { v, _ := util.GetProp[bool](style, key); return v }
	getString := func(key string) string { v, _ := util.GetProp[string](style, key); return v }
	getInt := func(key string) int { v, _ := util.GetProp[int](style, key); return v }

	s.Bold = getBool("bold")
	s.Italic = getBool("italic")
	s.Underline = getBool("underline")
	s.Strikethrough = getBool("strikethrough")
	s.Inverse = getBool("inverse")
	s.Blink = getBool("blink")
	s.Reverse = getBool("reverse")
	s.Foreground = getString("foreground")
	if s.Foreground == "" {
		s.Foreground = getString("color")
	}
	s.Background = getString("background")
	s.Dim = getBool("dim")
	s.Hidden = getBool("hidden")
	s.Display = getString("display")
	s.FlexDirection = getString("flexDirection")
	s.Gap = getInt("gap")
	s.ColumnGap = getInt("columnGap")
	s.RowGap = getInt("rowGap")
	s.MarginTop = getInt("marginTop")
	s.MarginBottom = getInt("marginBottom")
	s.MarginLeft = getInt("marginLeft")
	s.MarginRight = getInt("marginRight")

	s.BorderTopColor = getString("borderTopColor")
	s.BorderBottomColor = getString("borderBottomColor")
	s.BorderLeftColor = getString("borderLeftColor")
	s.BorderRightColor = getString("borderRightColor")

	s.BorderTopDim = getBool("borderTopDimColor")
	s.BorderBottomDim = getBool("borderBottomDimColor")
	s.BorderLeftDim = getBool("borderLeftDimColor")
	s.BorderRightDim = getBool("borderRightDimColor")

	s.BorderTop = true
	s.BorderBottom = true
	s.BorderLeft = true
	s.BorderRight = true

	if v, ok := util.GetProp[bool](style, "borderTop"); ok {
		s.BorderTop = v
	}
	if v, ok := util.GetProp[bool](style, "borderBottom"); ok {
		s.BorderBottom = v
	}
	if v, ok := util.GetProp[bool](style, "borderLeft"); ok {
		s.BorderLeft = v
	}
	if v, ok := util.GetProp[bool](style, "borderRight"); ok {
		s.BorderRight = v
	}

	if getBool("borderDimColor") {
		s.BorderTopDim = true
		s.BorderBottomDim = true
		s.BorderLeftDim = true
		s.BorderRightDim = true
	}

	// Shorthands
	if mx := getInt("marginX"); mx != 0 {
		s.MarginLeft = mx
		s.MarginRight = mx
	}
	if my := getInt("marginY"); my != 0 {
		s.MarginTop = my
		s.MarginBottom = my
	}
	if m := getInt("margin"); m != 0 {
		s.MarginTop = m
		s.MarginBottom = m
		s.MarginLeft = m
		s.MarginRight = m
	}
	if s.ColumnGap == 0 {
		s.ColumnGap = s.Gap
	}
	if s.RowGap == 0 {
		s.RowGap = s.Gap
	}

	s.Position = getString("position")
	s.Top = getInt("top")
	s.Left = getInt("left")
	s.Right = getInt("right")
	s.Bottom = getInt("bottom")
	s.ZIndex = getInt("zIndex")
	s.AlignItems = getString("alignItems")
	s.AlignSelf = getString("alignSelf")
	s.JustifyContent = getString("justifyContent")
	s.FlexGrow = getInt("flexGrow")
	s.FlexShrink = getInt("flexShrink")
	s.FlexBasis = getInt("flexBasis")
	s.Width = getInt("width")
	s.Height = getInt("height")

	if v, ok := util.GetProp[any](style, "foregroundGradient"); ok {
		g := &LinearGradient{Direction: "horizontal"}
		if from, ok := util.GetProp[string](v, "from"); ok {
			g.From = from
		}
		if to, ok := util.GetProp[string](v, "to"); ok {
			g.To = to
		}
		if dir, ok := util.GetProp[string](v, "direction"); ok {
			g.Direction = dir
		}
		s.ForegroundGradient = g
	}

	if v, ok := util.GetProp[any](style, "borderGradient"); ok {
		g := &RadialGradient{CenterX: 0.5, CenterY: 0.5, Radius: 1.0}
		if from, ok := util.GetProp[string](v, "from"); ok {
			g.From = from
		}
		if to, ok := util.GetProp[string](v, "to"); ok {
			g.To = to
		}
		if cx, ok := util.GetProp[float64](v, "centerX"); ok {
			g.CenterX = cx
		}
		if cy, ok := util.GetProp[float64](v, "centerY"); ok {
			g.CenterY = cy
		}
		if r, ok := util.GetProp[float64](v, "radius"); ok {
			g.Radius = r
		}
		s.BorderGradient = g
	}

	return s
}

type DiffOp int

const (
	DiffOpSame DiffOp = iota
	DiffOpUpdate
	DiffOpReplace
	DiffOpInsert
	DiffOpRemove
)

type Diff struct {
	Op      DiffOp
	Path    []int
	OldNode Node
	NewNode Node
	Index   int
}

func DiffTrees(oldTree, newTree Node) []Diff {
	var diffs []Diff
	diffRecursive([]int{}, oldTree, newTree, &diffs)
	return diffs
}

func diffRecursive(path []int, oldNode, newNode Node, diffs *[]Diff) {
	if oldNode == nil && newNode == nil {
		return
	}

	if oldNode == nil {
		*diffs = append(*diffs, Diff{Op: DiffOpInsert, Path: path, NewNode: newNode})
		return
	}

	if newNode == nil {
		*diffs = append(*diffs, Diff{Op: DiffOpRemove, Path: path, OldNode: oldNode})
		return
	}

	if oldNode.GetNodeType() != newNode.GetNodeType() {
		*diffs = append(*diffs, Diff{Op: DiffOpReplace, Path: path, OldNode: oldNode, NewNode: newNode})
		return
	}

	switch old := oldNode.(type) {
	case *TextNode:
		new := newNode.(*TextNode)
		if old.Content != new.Content {
			*diffs = append(*diffs, Diff{Op: DiffOpUpdate, Path: path, OldNode: oldNode, NewNode: newNode})
		}

	case *Element:
		new := newNode.(*Element)
		if old.Type != new.Type {
			*diffs = append(*diffs, Diff{Op: DiffOpReplace, Path: path, OldNode: oldNode, NewNode: newNode})
			return
		}

		if !propsEqual(old.Props, new.Props) || !old.Style.Equals(&new.Style) {
			*diffs = append(*diffs, Diff{Op: DiffOpUpdate, Path: path, OldNode: oldNode, NewNode: newNode})
		}

		oldChildren := old.Children
		newChildren := new.Children

		maxLen := len(oldChildren)
		if len(newChildren) > maxLen {
			maxLen = len(newChildren)
		}

		for i := 0; i < maxLen; i++ {
			newPath := append([]int{}, path...)
			newPath = append(newPath, i)
			if i >= len(oldChildren) {
				*diffs = append(*diffs, Diff{Op: DiffOpInsert, Path: newPath, NewNode: newChildren[i], Index: i})
			} else if i >= len(newChildren) {
				*diffs = append(*diffs, Diff{Op: DiffOpRemove, Path: newPath, OldNode: oldChildren[i], Index: i})
			} else {
				diffRecursive(newPath, oldChildren[i], newChildren[i], diffs)
			}
		}

	case *Fragment:
		new := newNode.(*Fragment)
		oldChildren := old.Children
		newChildren := new.Children

		maxLen := len(oldChildren)
		if len(newChildren) > maxLen {
			maxLen = len(newChildren)
		}

		for i := 0; i < maxLen; i++ {
			newPath := append([]int{}, path...)
			newPath = append(newPath, i)
			if i >= len(oldChildren) {
				*diffs = append(*diffs, Diff{Op: DiffOpInsert, Path: newPath, NewNode: newChildren[i], Index: i})
			} else if i >= len(newChildren) {
				*diffs = append(*diffs, Diff{Op: DiffOpRemove, Path: newPath, OldNode: oldChildren[i], Index: i})
			} else {
				diffRecursive(newPath, oldChildren[i], newChildren[i], diffs)
			}
		}
	}
}

func propsEqual(a, b any) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return reflect.DeepEqual(a, b)
}

func (s *Style) Equals(other *Style) bool {
	if s == other {
		return true
	}
	if s == nil || other == nil {
		return false
	}
	return s.Bold == other.Bold &&
		s.Italic == other.Italic &&
		s.Underline == other.Underline &&
		s.Strikethrough == other.Strikethrough &&
		s.Inverse == other.Inverse &&
		s.Blink == other.Blink &&
		s.Reverse == other.Reverse &&
		s.Foreground == other.Foreground &&
		s.Background == other.Background &&
		s.Dim == other.Dim &&
		s.Hidden == other.Hidden &&
		s.AlignItems == other.AlignItems &&
		s.JustifyContent == other.JustifyContent &&
		s.ZIndex == other.ZIndex &&
		s.FlexGrow == other.FlexGrow &&
		s.FlexShrink == other.FlexShrink &&
		s.FlexBasis == other.FlexBasis &&
		s.ColumnGap == other.ColumnGap &&
		s.RowGap == other.RowGap &&
		s.MarginTop == other.MarginTop &&
		s.MarginBottom == other.MarginBottom &&
		s.MarginLeft == other.MarginLeft &&
		s.MarginRight == other.MarginRight &&
		s.BorderTopColor == other.BorderTopColor &&
		s.BorderBottomColor == other.BorderBottomColor &&
		s.BorderLeftColor == other.BorderLeftColor &&
		s.BorderRightColor == other.BorderRightColor &&
		s.BorderTopDim == other.BorderTopDim &&
		s.BorderBottomDim == other.BorderBottomDim &&
		s.BorderLeftDim == other.BorderLeftDim &&
		s.BorderRightDim == other.BorderRightDim &&
		s.BorderTop == other.BorderTop &&
		s.BorderBottom == other.BorderBottom &&
		s.BorderLeft == other.BorderLeft &&
		s.BorderRight == other.BorderRight
}

type TreeState struct {
	mu       sync.RWMutex
	Root     Node
	Version  int
	PrevTree Node
}

func NewTreeState(root Node) *TreeState {
	return &TreeState{
		Root:    root,
		Version: 1,
	}
}

func (ts *TreeState) Update(newRoot Node) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.PrevTree != nil {
		if releasable, ok := ts.PrevTree.(interface{ Release() }); ok {
			releasable.Release()
		}
	}
	ts.PrevTree = ts.Root
	ts.Root = newRoot
	ts.Version++
}

func (ts *TreeState) GetDiff() []Diff {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	if ts.PrevTree == nil {
		return []Diff{{Op: DiffOpInsert, Path: []int{}, NewNode: ts.Root}}
	}
	return DiffTrees(ts.PrevTree, ts.Root)
}
