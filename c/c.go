package c

import (
	"github.com/j-p/recliner/components"
	"github.com/j-p/recliner/events"
	"github.com/j-p/recliner/vdom"
)

type BorderStyle string

const (
	BorderStyleNone         BorderStyle = "none"
	BorderStyleSingle       BorderStyle = "single"
	BorderStyleDouble       BorderStyle = "double"
	BorderStyleRound        BorderStyle = "round"
	BorderStyleBold         BorderStyle = "bold"
	BorderStyleSingleDouble BorderStyle = "singleDouble"
	BorderStyleDoubleSingle BorderStyle = "doubleSingle"
	BorderStyleClassic      BorderStyle = "classic"
)

type DisplayStyle string

const (
	DisplayBlock  DisplayStyle = "block"
	DisplayInline DisplayStyle = "inline"
	DisplayFlex   DisplayStyle = "flex"
	DisplayNone   DisplayStyle = "none"
)

type FlexDirection string

const (
	FlexDirectionRow    FlexDirection = "row"
	FlexDirectionColumn FlexDirection = "column"
)

type JustifyContent string

const (
	JustifyContentFlexStart    JustifyContent = "flex-start"
	JustifyContentFlexEnd      JustifyContent = "flex-end"
	JustifyContentCenter       JustifyContent = "center"
	JustifyContentSpaceBetween JustifyContent = "space-between"
	JustifyContentSpaceAround  JustifyContent = "space-around"
	JustifyContentSpaceEvenly  JustifyContent = "space-evenly"
	JustifyContentStretch      JustifyContent = "stretch"
)

type AlignItems string

const (
	AlignItemsFlexStart AlignItems = "flex-start"
	AlignItemsFlexEnd   AlignItems = "flex-end"
	AlignItemsCenter    AlignItems = "center"
	AlignItemsStretch   AlignItems = "stretch"
	AlignItemsBaseline  AlignItems = "baseline"
)

type AlignSelf string

const (
	AlignSelfAuto      AlignSelf = "auto"
	AlignSelfFlexStart AlignSelf = "flex-start"
	AlignSelfFlexEnd   AlignSelf = "flex-end"
	AlignSelfCenter    AlignSelf = "center"
	AlignSelfStretch   AlignSelf = "stretch"
)

type Position string

const (
	PositionRelative Position = "relative"
	PositionAbsolute Position = "absolute"
	PositionFixed    Position = "fixed"
)

type ButtonVariant string

const (
	ButtonVariantFilled   ButtonVariant = "filled"
	ButtonVariantOutlined ButtonVariant = "outlined"
	ButtonVariantText     ButtonVariant = "text"
)

type ButtonStyle string

const (
	ButtonStylePrimary   ButtonStyle = "primary"
	ButtonStyleSecondary ButtonStyle = "secondary"
	ButtonStyleSuccess   ButtonStyle = "success"
	ButtonStyleDanger    ButtonStyle = "danger"
	ButtonStyleWarning   ButtonStyle = "warning"
	ButtonStyleInfo      ButtonStyle = "info"
)

type ButtonSize string

const (
	ButtonSizeSmall  ButtonSize = "small"
	ButtonSizeMedium ButtonSize = "medium"
	ButtonSizeLarge  ButtonSize = "large"
)

type LinearGradient struct{ From, To, Direction string }
type RadialGradient struct {
	From, To                 string
	CenterX, CenterY, Radius float64
}

type StyleProps struct {
	Display            DisplayStyle
	FlexDirection      FlexDirection
	Gap                interface{}
	Width, Height      interface{}
	Background         string
	AlignItems         AlignItems
	AlignSelf          AlignSelf
	FlexGrow           int
	Position           Position
	Top, Left          interface{}
	JustifyContent     JustifyContent
	ZIndex             int
	ForegroundGradient *LinearGradient
	BorderGradient     *RadialGradient
}

func (p StyleProps) ToStyle() vdom.Style {
	s := vdom.Style{}
	if p.Display != "" {
		s.Display = string(p.Display)
	}
	if p.FlexDirection != "" {
		s.FlexDirection = string(p.FlexDirection)
	}
	if v, ok := p.Gap.(int); ok {
		s.Gap = v
	}
	if v, ok := p.Width.(int); ok {
		s.Width = v
	}
	if v, ok := p.Height.(int); ok {
		s.Height = v
	}
	if p.Background != "" {
		s.Background = p.Background
	}
	if p.Position != "" {
		s.Position = string(p.Position)
	}
	if v, ok := p.Top.(int); ok {
		s.Top = v
	}
	if v, ok := p.Left.(int); ok {
		s.Left = v
	}
	if p.AlignItems != "" {
		s.AlignItems = string(p.AlignItems)
	}
	if p.AlignSelf != "" {
		s.AlignSelf = string(p.AlignSelf)
	}
	if p.FlexGrow != 0 {
		s.FlexGrow = p.FlexGrow
	}
	if p.JustifyContent != "" {
		s.JustifyContent = string(p.JustifyContent)
	}
	if p.ZIndex != 0 {
		s.ZIndex = p.ZIndex
	}
	if p.ForegroundGradient != nil {
		s.ForegroundGradient = &vdom.LinearGradient{
			From:      p.ForegroundGradient.From,
			To:        p.ForegroundGradient.To,
			Direction: p.ForegroundGradient.Direction,
		}
	}
	if p.BorderGradient != nil {
		s.BorderGradient = &vdom.RadialGradient{
			From:    p.BorderGradient.From,
			To:      p.BorderGradient.To,
			CenterX: p.BorderGradient.CenterX,
			CenterY: p.BorderGradient.CenterY,
			Radius:  p.BorderGradient.Radius,
		}
	}
	return s
}

type BoxProps struct {
	BorderStyle       BorderStyle
	BorderColor       string
	Padding           int
	Style             StyleProps
	OnClick           func(events.MouseEvent)
	ClearFocusOnClick *bool
}

type TextStyler interface {
	ToStyle() vdom.Style
	Color(string) TextStyler
	BG(string) TextStyler
	Bold() TextStyler
	Italic() TextStyler
	Underline() TextStyler
	Dim() TextStyler
	Inverse() TextStyler
	Hidden() TextStyler
	Strikethrough() TextStyler
	Blink() TextStyler
	Gradient(string, string) TextStyler
	GradientVertical(string, string) TextStyler
}

type TextStyleProps struct {
	foreground, background                                              string
	bold, italic, underline, dim, inverse, hidden, strikethrough, blink bool
	foregroundGradient                                                  *LinearGradient
}

func (t *TextStyleProps) ToStyle() vdom.Style {
	s := vdom.Style{
		Foreground:    t.foreground,
		Background:    t.background,
		Bold:          t.bold,
		Italic:        t.italic,
		Underline:     t.underline,
		Dim:           t.dim,
		Inverse:       t.inverse,
		Hidden:        t.hidden,
		Strikethrough: t.strikethrough,
		Blink:         t.blink,
	}
	if t.foregroundGradient != nil {
		s.ForegroundGradient = &vdom.LinearGradient{
			From:      t.foregroundGradient.From,
			To:        t.foregroundGradient.To,
			Direction: t.foregroundGradient.Direction,
		}
	}
	return s
}

func (t *TextStyleProps) Color(c string) TextStyler { t.foreground = c; return t }
func (t *TextStyleProps) BG(c string) TextStyler    { t.background = c; return t }
func (t *TextStyleProps) Bold() TextStyler          { t.bold = true; return t }
func (t *TextStyleProps) Italic() TextStyler        { t.italic = true; return t }
func (t *TextStyleProps) Underline() TextStyler     { t.underline = true; return t }
func (t *TextStyleProps) Dim() TextStyler           { t.dim = true; return t }
func (t *TextStyleProps) Inverse() TextStyler       { t.inverse = true; return t }
func (t *TextStyleProps) Hidden() TextStyler        { t.hidden = true; return t }
func (t *TextStyleProps) Strikethrough() TextStyler { t.strikethrough = true; return t }
func (t *TextStyleProps) Blink() TextStyler         { t.blink = true; return t }
func (t *TextStyleProps) Gradient(f, to string) TextStyler {
	t.foregroundGradient = &LinearGradient{f, to, "horizontal"}
	return t
}
func (t *TextStyleProps) GradientVertical(f, to string) TextStyler {
	t.foregroundGradient = &LinearGradient{f, to, "vertical"}
	return t
}

func TextStyle() TextStyler { return &TextStyleProps{} }

type ButtonProps struct {
	Label    string
	OnClick  func(events.MouseEvent)
	Variant  ButtonVariant
	Style    ButtonStyle
	Size     ButtonSize
	Disabled bool
	ID       string
}

type CheckBoxProps struct {
	Label    vdom.Node
	Checked  bool
	OnChange func(bool)
	ID       string
}

type TextBoxProps struct {
	Value, Placeholder   string
	Width, Height        int
	ReadOnly, Scrollable bool
	OnChange             func(string)
	OnClick              func(events.MouseEvent)
	Label                vdom.Node
	ID                   string
}

type InputProps struct {
	Value, Placeholder string
	Width              int
	OnChange           func(string)
	ID                 string
	AutoFocus          bool
}

type MenuProps struct {
	Title    string
	Items    []string
	OnSelect func(string)
	OnClose  func()
	X, Y     int
}

type ScrollBarProps struct {
	Orientation            string
	Length, Pos, Size      int
	ThumbColor, TrackColor string
	OnClick                func(events.MouseEvent)
}

type AccordionItem struct {
	Title   string
	Content vdom.Node
	ID      string
}
type AccordionProps struct {
	Items           []AccordionItem
	AllowMultiple   bool
	DefaultExpanded []string
	ID              string
}

type TabItem struct {
	Title   string
	Content vdom.Node
	ID      string
}
type TabsProps struct {
	Items                    []TabItem
	DefaultActive, ActiveTab string
	OnChange                 func(string)
	ID                       string
}

type ImageProps struct {
	Src           string
	Width, Height int
	Mode, ID      string
}

type ProgressBarProps struct {
	Value                                                               float64
	Width, Height                                                       int
	Orientation, LabelPosition, FromColor, ToColor, FillChar, EmptyChar string
}

type BoxElementProps struct {
	BorderStyle string
	BorderColor string
	Padding     int
	Style       vdom.Style
	OnClick     func(events.MouseEvent)
	ScrollTop   int
	ScrollLeft  int
}

func Box(p BoxProps, c ...vdom.Node) vdom.Node {
	res := &BoxElementProps{
		BorderStyle: string(p.BorderStyle),
		BorderColor: p.BorderColor,
		Padding:     p.Padding,
		Style:       p.Style.ToStyle(),
		OnClick:     p.OnClick,
	}
	children := make([]any, len(c))
	for i, v := range c {
		children[i] = v
	}
	return vdom.CreateElement("box", res, children...)
}

type TextProps struct {
	Content string
	Style   TextStyler
}

type TextElementProps struct {
	Style   vdom.Style
	OnClick func(events.MouseEvent)
}

func Text(p TextProps) vdom.Node {
	var style vdom.Style
	if p.Style != nil {
		style = p.Style.ToStyle()
	}

	res := &TextElementProps{
		Style: style,
	}

	return vdom.CreateElement("text", res, p.Content)
}

func Button(p ButtonProps) vdom.Node           { return components.Button(p) }
func Input(p InputProps) vdom.Node             { return components.Input(p) }
func CheckBox(p CheckBoxProps) vdom.Node       { return components.CheckBox(p) }
func TextBox(p TextBoxProps) vdom.Node         { return components.TextBox(p) }
func Menu(p MenuProps) vdom.Node               { return components.Menu(p) }
func Accordion(p AccordionProps) vdom.Node     { return components.Accordion(p) }
func ScrollBar(p ScrollBarProps) vdom.Node     { return components.ScrollBar(p) }
func Tabs(p TabsProps) vdom.Node               { return components.Tabs(p) }
func Image(p ImageProps) vdom.Node             { return components.Image(p) }
func ProgressBar(p ProgressBarProps) vdom.Node { return components.ProgressBar(p) }

func Spacer(h int) vdom.Node {
	c := make([]vdom.Node, h)
	for i := range h {
		c[i] = &vdom.TextNode{Content: " "}
	}
	return &vdom.Fragment{Children: c}
}
func Newline(n int) vdom.Node {
	c := make([]vdom.Node, n)
	for i := range n {
		c[i] = &vdom.TextNode{Content: " "}
	}
	return &vdom.Fragment{Children: c}
}
