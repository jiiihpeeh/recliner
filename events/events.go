package events

// Events package holds shared input event types (keyboard/mouse) used across the app.

type KeyPressEvent struct {
	Key    string
	Escape bool
	Ctrl   bool
	Shift  bool
	Meta   bool
}

type MouseAction string

const (
	MouseActionPress      MouseAction = "Press"
	MouseActionRelease    MouseAction = "Release"
	MouseActionMotion     MouseAction = "Motion"
	MouseActionHover      MouseAction = "Hover"
	MouseActionScrollUp   MouseAction = "ScrollUp"
	MouseActionScrollDown MouseAction = "ScrollDown"
)

// MouseButton represents common mouse button codes (SGR encoding)
type MouseButton int

const (
	MouseButtonLeft       MouseButton = 0
	MouseButtonMiddle     MouseButton = 1
	MouseButtonRight      MouseButton = 2
	MouseButtonScrollUp   MouseButton = 64
	MouseButtonScrollDown MouseButton = 65
)

type MouseEvent struct {
	X       int
	Y       int
	ScreenX int
	ScreenY int
	RelX    int
	RelY    int
	Button  MouseButton
	Action  MouseAction
	Ctrl    bool
}
