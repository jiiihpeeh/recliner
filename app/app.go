package app

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jiiihpeeh/recliner/debug"
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/render"
	"github.com/jiiihpeeh/recliner/vdom"
	"golang.org/x/term"
)

var logFile *os.File
var debugEnabled bool

func init() {
	var err error
	logFile, err = os.OpenFile("/tmp/recliner_debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		logDebug("APP", "=== ReCLIner Started ===")
	}
}

func SetDebugMode(enabled bool) {
	debugEnabled = enabled
}

func DebugLog(category, msg string) {
	if !debugEnabled {
		return
	}
	logDebug(category, msg)
}

func startDebugServer() error {
	return debug.StartDebugServer()
}

func logDebug(category, msg string) {
	if logFile != nil {
		fmt.Fprintf(logFile, "[%s] [%s] %s\n", time.Now().Format("15:04:05"), category, msg)
		logFile.Sync()
	}

	// Send to debug server if available
	debug.SendDebugMessage(category, msg)
}

type App struct {
	root         vdom.ComponentFunc
	hooksCtx     *hooks.HooksContext
	renderer     *render.Renderer
	treeState    *vdom.TreeState
	mu           sync.Mutex
	running      int32
	exiting      int32
	needsRender  int32
	stdin        *os.File
	oldState     *term.State
	isTTY        bool
	altScreen    bool  // State: are we currently in alt screen
	useAltScreen bool  // Config: should we use alt screen
	scrollY      int64 // Vertical scroll position (atomic)
	// runtime options
	headless     bool
	keyMapping   map[string]string
	mouseCapture func(events.MouseEvent) // Handler for dragged mouse events
	debug        bool
}

func New(root vdom.ComponentFunc) *App {
	return NewWithOptions(root, AppOptions{})
}

// AppOptions controls runtime options for the App
type AppOptions struct {
	Headless   bool
	KeyMapping map[string]string
	NoAlt      bool
	Debug      bool
}

// NewWithOptions constructs an App with options
func NewWithOptions(root vdom.ComponentFunc, opts AppOptions) *App {
	SetDebugMode(opts.Debug) // Set global debug mode

	// Set debug callback for hooks if debug mode is enabled
	if opts.Debug {
		hooks.SetDebugCallback(func(category, msg string) {
			logDebug(category, msg)
		})
	}

	ctx := hooks.NewHooksContext()
	app := &App{
		root:       root,
		hooksCtx:   ctx,
		renderer:   render.NewRenderer(),
		treeState:  vdom.NewTreeState(nil),
		running:    0,
		stdin:      os.Stdin,
		headless:   opts.Headless,
		keyMapping: opts.KeyMapping,
		// respect NoAlt option
		useAltScreen: !opts.NoAlt,
		altScreen:    false,
		debug:        opts.Debug,
	}

	hooks.SetContext(ctx)
	// Initialize HooksContext handles
	ctx.App = app
	ctx.Stdin = os.Stdin
	ctx.Stdout = os.Stdout
	ctx.Stderr = os.Stderr

	// Set update callback immediately
	ctx.SetUpdateCallback(app.onUpdate)

	// Start debug server if debug mode is enabled
	if opts.Debug {
		if err := startDebugServer(); err != nil {
			fmt.Printf("Warning: Failed to start debug server: %v\n", err)
		}
	}

	return app
}

func (a *App) Run() error {
	logDebug("APP", "App.Run() called")

	// Ensure the hooks context used by this app is global
	hooks.SetContext(a.hooksCtx)
	a.hooksCtx.SetUpdateCallback(a.onUpdate)

	defer func() {
		if r := recover(); r != nil {
			logDebug("APP", fmt.Sprintf("PANIC: %v", r))
			a.teardownTerminal()
			fmt.Printf("PANIC: %v\n", r)
			os.Exit(1)
		}
	}()

	if !a.headless {
		if err := a.setupTerminal(); err != nil {
			logDebug("APP", fmt.Sprintf("setupTerminal failed: %v", err))
			return err
		}
		logDebug("APP", "Terminal setup complete")
	}

	atomic.StoreInt32(&a.running, 1)
	logDebug("APP", "App running")

	// Get initial window size
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		a.hooksCtx.SetWindowSize(w, h)
		a.renderer.SetSize(w, h)
	}

	go a.inputLoop()
	logDebug("INPUT", "Input loop goroutine spawned")

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logDebug("SIGNAL", "Signal received, exiting")
		a.Exit()
	}()
	logDebug("SIGNAL", "Signal handler registered")

	// Handle terminal resize (SIGWINCH) and re-render
	go func() {
		sigWin := make(chan os.Signal, 1)
		signal.Notify(sigWin, syscall.SIGWINCH)
		for range sigWin {
			// debounce a bit
			time.Sleep(50 * time.Millisecond)
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				logDebug("SIGNAL", fmt.Sprintf("SIGWINCH: GetSize error: %v", err))
				continue
			}
			logDebug("SIGNAL", fmt.Sprintf("SIGWINCH: new size %dx%d", w, h))
			a.hooksCtx.SetWindowSize(w, h)
			a.renderer.SetSize(w, h)
			a.render()
		}
	}()

	if a.isTTY && a.useAltScreen && !a.headless {
		a.enterAltScreen()
		// fmt.Print("\x1b[H\x1b[2J") // Initial clear handled by first render
	}
	a.render()
	logDebug("RENDER", "Initial render complete")

	// Keep the app running
	if !a.isTTY && !a.headless {
		// When not in TTY mode, wait for stdin to close
		buf := make([]byte, 1)
		os.Stdin.Read(buf)
		a.Exit()
		return nil
	}

	logDebug("INPUT", "Entering select...")
	select {}
}

func (a *App) setupTerminal() error {
	a.stdin = os.Stdin
	logDebug("APP", "setupTerminal: stdin obtained")

	// Check if stdin is a TTY
	a.isTTY = term.IsTerminal(int(a.stdin.Fd()))
	logDebug("APP", fmt.Sprintf("setupTerminal: isTTY=%v", a.isTTY))

	// If stdin is not a TTY, try to open /dev/tty as a fallback
	if !a.isTTY {
		logDebug("APP", "setupTerminal: Trying /dev/tty as fallback")
		f, err := os.Open("/dev/tty")
		if err == nil {
			a.stdin = f
			a.isTTY = term.IsTerminal(int(a.stdin.Fd()))
			logDebug("APP", fmt.Sprintf("setupTerminal: /dev/tty opened, isTTY=%v", a.isTTY))
		} else {
			logDebug("APP", fmt.Sprintf("setupTerminal: /dev/tty open failed: %v", err))
		}
	}

	if a.isTTY {
		state, err := term.MakeRaw(int(a.stdin.Fd()))
		if err != nil {
			logDebug("APP", fmt.Sprintf("setupTerminal: MakeRaw failed %v, continuing without raw mode", err))
			a.isTTY = false
		} else {
			a.oldState = state
			logDebug("APP", "setupTerminal: Raw mode enabled")
		}
	}

	// Update hooks context with the finalized stdin handle
	a.hooksCtx.Stdin = a.stdin

	return nil
}

func (a *App) teardownTerminal() {
	logDebug("APP", "teardownTerminal() called")
	if a == nil || a.stdin == nil {
		logDebug("APP", "teardownTerminal: early return - nil check failed")
		return
	}

	// Leave alternate screen if active
	if a.altScreen {
		logDebug("APP", "teardownTerminal: leaving alt screen")
		a.leaveAltScreen()
	} else {
		logDebug("APP", "teardownTerminal: not in alt screen")
	}

	if a.oldState != nil {
		logDebug("APP", "teardownTerminal: restoring terminal")
		term.Restore(int(a.stdin.Fd()), a.oldState)
	}
	fmt.Print("\x1b[0m")
	logDebug("APP", "teardownTerminal: complete")
}

func (a *App) inputLoop() {
	logDebug("INPUT", "inputLoop thread started")
	// Re-set context for input loop goroutine
	hooks.SetContext(a.hooksCtx)

	defer func() {
		if r := recover(); r != nil {
			logDebug("INPUT", fmt.Sprintf("Input loop panic: %v", r))
		}
		logDebug("INPUT", "Input loop ended")
	}()

	logDebug("INPUT", "Input loop started")

	if a.stdin == nil {
		logDebug("INPUT", "Input loop: stdin is nil, returning")
		return
	}

	buf := make([]byte, 1024)
	for atomic.LoadInt32(&a.running) == 1 {
		n, err := a.stdin.Read(buf)
		if err != nil {
			logDebug("INPUT", fmt.Sprintf("Input loop read error: %v (isTTY=%v)", err, a.isTTY))
			if atomic.LoadInt32(&a.running) == 0 {
				return
			}
			// In non-TTY mode, just wait and retry
			if !a.isTTY {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			// Otherwise exit
			logDebug("INPUT", "Exiting input loop due to fatal error")
			atomic.StoreInt32(&a.running, 0)
			return
		}

		// Check for Ctrl+C (ASCII 3) anywhere in the buffer
		for i := 0; i < n; i++ {
			if buf[i] == 3 {
				fmt.Fprintln(os.Stderr, "\nCtrl+C detected, exiting...")
				logDebug("INPUT", "Ctrl+C detected in raw buffer, exiting")
				a.Exit()
				return
			}
		}

		event := a.parseInput(buf[:n])
		if event != nil {
			switch e := event.(type) {
			case *events.KeyPressEvent:
				hooks.LogKeyPress(*e)

				// Global scroll keys if nothing is focused or specific keys are used
				if e.Key == "pageup" && a.hooksCtx.FocusID == "" {
					a.Scroll(-10)
				} else if e.Key == "pagedown" && a.hooksCtx.FocusID == "" {
					a.Scroll(10)
				}

				if e.Key == "ctrl+c" {
					fmt.Fprintln(os.Stderr, "Ctrl+C detected, exiting...")
					logDebug("INPUT", "Ctrl+C detected in loop, exiting")
					a.Exit()
					return
				}

				handlers := a.hooksCtx.GetInputHandlers()
				for _, h := range handlers {
					func() {
						defer func() {
							if r := recover(); r != nil {
								logDebug("INPUT", fmt.Sprintf("KeyHandler panic: %v", r))
							}
						}()
						h(*e)
					}()
				}

			case *events.MouseEvent:
				logDebug("INPUT", fmt.Sprintf("Mouse: %v", e))

				// Adjust mouse coordinates
				adjustedEvent := *e
				adjustedEvent.ScreenX = e.X
				adjustedEvent.ScreenY = e.Y

				func() {
					defer func() {
						if r := recover(); r != nil {
							logDebug("INPUT", fmt.Sprintf("MouseEvent panic: %v", r))
							a.mouseCapture = nil
						}
					}()

					// Handle Capture
					if a.mouseCapture != nil {
						a.mouseCapture(adjustedEvent)
						if e.Action == events.MouseActionRelease {
							a.mouseCapture = nil
						}
						return
					}

					// Standard Hit Test
					switch e.Action {
					case events.MouseActionScrollUp:
						handler := a.renderer.HitTest(adjustedEvent)
						if handler == nil {
							a.Scroll(-3)
						} else {
							func() {
								defer func() { recover() }()
								handler(adjustedEvent)
							}()
						}
					case events.MouseActionScrollDown:
						handler := a.renderer.HitTest(adjustedEvent)
						if handler == nil {
							a.Scroll(3)
						} else {
							func() {
								defer func() { recover() }()
								handler(adjustedEvent)
							}()
						}
					case events.MouseActionPress: // Only new presses trigger hit tests if not capturing
						handler := a.renderer.HitTest(adjustedEvent)

						if e.Button == events.MouseButtonLeft {
							a.hooksCtx.UseFocusManager().Blur()
						}

						if handler != nil {
							a.mouseCapture = handler // Start capture
							handler(adjustedEvent)
						}
					case events.MouseActionMotion:
						// Hover handling
					}
				}()
			}
		}
	}
}

func (a *App) parseInput(buf []byte) interface{} {
	if len(buf) == 0 {
		return nil
	}

	// Mouse Event Parsing (SGR 1006)
	// Sequence: \x1b[<b;x;yM or m
	if len(buf) > 3 && buf[0] == '\x1b' && buf[1] == '[' && buf[2] == '<' {
		var btn, x, y int
		var typeCode rune
		// Find end of sequence
		endIdx := -1
		for i := 3; i < len(buf); i++ {
			if buf[i] == 'M' || buf[i] == 'm' {
				endIdx = i
				break
			}
		}
		if endIdx != -1 {
			// Extract b;x;y then typeCode is the char at endIdx
			part := string(buf[3:endIdx])
			typeCode = rune(buf[endIdx])

			// Parse strictly numbers
			n, err := fmt.Sscanf(part, "%d;%d;%d", &btn, &x, &y)
			if err == nil && n == 3 {
				// Handle Scroll Wheel
				if btn == 64 {
					return &events.MouseEvent{
						X:      x - 1,
						Y:      y - 1,
						Button: events.MouseButtonScrollUp,
						Action: events.MouseActionScrollUp,
					}
				}
				if btn == 65 {
					return &events.MouseEvent{
						X:      x - 1,
						Y:      y - 1,
						Button: events.MouseButtonScrollDown,
						Action: events.MouseActionScrollDown,
					}
				}

				var action events.MouseAction = events.MouseActionPress

				// Check for drag (bit 5 / 32)
				// Note: SGR mode separates release with 'm', but drag is encoded in the button value
				isDrag := (btn & 32) != 0

				if typeCode == 'm' {
					action = events.MouseActionRelease
					// Release events in SGR might still have drag bit set if released during drag?
					// Usually SGR release just tells which button was released.
					// We'll rely on 'm' for release.
				} else if isDrag {
					action = events.MouseActionMotion
					btn = btn &^ 32 // Strip drag bit to get meaningful button
				}

				return &events.MouseEvent{
					X:      x - 1, // Convert 1-based to 0-based
					Y:      y - 1,
					Button: events.MouseButton(btn),
					Action: action,
				}
			}
		}
	}

	if buf[0] == 3 {
		return &events.KeyPressEvent{
			Key:    "ctrl+c",
			Ctrl:   true,
			Escape: false,
		}
	}

	if buf[0] == 127 || buf[0] == 8 {
		return &events.KeyPressEvent{
			Key:    "backspace",
			Escape: false,
		}
	}

	if len(buf) >= 3 && buf[0] == '\x1b' && buf[1] == '[' && buf[2] == '3' && len(buf) >= 4 && buf[3] == '~' {
		return &events.KeyPressEvent{
			Key:    "delete",
			Escape: true,
		}
	}

	if len(buf) >= 1 && buf[0] == '\x1b' {
		if len(buf) >= 2 && buf[1] == '[' {
			if len(buf) >= 3 {
				// Handle Shift+Tab (ESC [ Z)
				if buf[2] == 'Z' {
					return &events.KeyPressEvent{
						Key:   "tab",
						Shift: true,
					}
				}
				keyMap := map[byte]string{
					'A': "up",
					'B': "down",
					'C': "right",
					'D': "left",
					'H': "home",
					'F': "end",
				}
				if key, ok := keyMap[buf[2]]; ok {
					return &events.KeyPressEvent{
						Key:    key,
						Escape: true,
						Ctrl:   false,
						Shift:  false,
						Meta:   false,
					}
				}
			}

			// Check for tilde sequences (PageUp, PageDown, Delete, Home\x1b[1~, End\x1b[4~)
			// Format: ESC [ <num> ~
			if len(buf) >= 4 && buf[len(buf)-1] == '~' {
				var code int
				// extract content between [ (index 1) and ~ (last index)
				// buf[2 : len(buf)-1]
				if _, err := fmt.Sscanf(string(buf[2:len(buf)-1]), "%d", &code); err == nil {
					key := ""
					switch code {
					case 1:
						key = "home"
					case 2:
						key = "insert" // Optional
					case 3:
						key = "delete"
					case 4:
						key = "end"
					case 5:
						key = "pageup"
					case 6:
						key = "pagedown"
					}
					if key != "" {
						return &events.KeyPressEvent{
							Key:    key,
							Escape: true,
						}
					}
				}
			}

			// Incomplete escape sequence - ignore for now
			return nil
		}
		// If it's a bare escape followed by another character, ignore it
		// to avoid false positives from terminal startup sequences
		if len(buf) > 1 {
			return nil
		}
		// Only treat a bare \x1b as escape if it's the only character
		// and only after a small delay to ensure we got the whole sequence
		return &events.KeyPressEvent{
			Key:    "escape",
			Escape: true,
		}
	}

	if len(buf) >= 1 {
		// Single character handling
		if buf[0] == '\t' {
			return &events.KeyPressEvent{
				Key:    "tab",
				Escape: false,
			}
		}
		if buf[0] == '\r' || buf[0] == '\n' {
			return &events.KeyPressEvent{
				Key:    "enter",
				Escape: false,
			}
		}

		// Map Control Characters (0x01-0x1A -> ctrl+a-z)
		// Exclude 3 (ctrl+c), 13 (cr), 10 (lf), 9 (tab), 27 (esc) if any slip through
		// We already handle \t(9), \r(13), \n(10). \x1b(27) is handled above.
		// We also want to map backspace (ctrl+h -> 8).
		if buf[0] >= 1 && buf[0] <= 26 {
			// ctrl+c is 3, ctrl+a is 1
			if buf[0] == 3 {
				return &events.KeyPressEvent{Key: "ctrl+c", Ctrl: true}
			}
			// letter = value + 96 (1 -> 'a')
			letter := rune(buf[0] + 96)
			return &events.KeyPressEvent{
				Key:  fmt.Sprintf("ctrl+%c", letter),
				Ctrl: true,
			}
		}

		return &events.KeyPressEvent{
			Key:    string(buf),
			Escape: false,
			Ctrl:   false,
			Shift:  false,
			Meta:   false,
		}
	}

	return nil
}

func (a *App) onUpdate() {
	if atomic.LoadInt32(&a.running) == 0 {
		return
	}
	atomic.StoreInt32(&a.needsRender, 1)
	logDebug("APP", "onUpdate called, triggering render")

	// Try to render immediately if possible.
	// render() uses TryLock so it's safe to call frequently.
	go a.render()
}

func (a *App) render() {
	if !a.mu.TryLock() {
		atomic.StoreInt32(&a.needsRender, 1)
		return
	}
	defer a.mu.Unlock()

	for {
		atomic.StoreInt32(&a.needsRender, 0)

		if atomic.LoadInt32(&a.running) == 0 {
			return
		}

		// Ensure the hooks context used by this app is global while we render
		hooks.SetContext(a.hooksCtx)

		a.hooksCtx.Reset()
		tree := a.root(nil)
		a.hooksCtx.Finalize()

		a.treeState.Update(tree)

		// Render to buffer (returns diff)
		output := a.renderer.Render(tree)

		// If headless, we don't output to stdout
		if a.headless {
			// Flush effects even in headless mode
			a.hooksCtx.FlushLayoutEffects()
			a.hooksCtx.FlushEffects()
			if atomic.LoadInt32(&a.needsRender) == 0 {
				break
			}
			continue
		}

		// Use the renderer's output directly. It already contains optimal cursor movements (diffing).
		if output != "" {
			// Write directly to stdout with explicit flush
			// Hide cursor during update
			fmt.Fprint(os.Stdout, output)

			// Handle Cursor State (if requested by hooks)
			if a.hooksCtx.CursorVisible {
				screenY := a.hooksCtx.CursorY
				screenX := a.hooksCtx.CursorX

				_, h := a.renderer.GetSize()
				if screenY >= 0 && screenY < h {
					fmt.Fprintf(os.Stdout, "\x1b[%d;%dH\x1b[?25h", screenY+1, screenX+1)
				}
			}
			os.Stdout.Sync()
		}

		// Flush effects after paint
		// WARNING: Hooks might trigger updates!
		// To prevent infinite loops or deep recursion, we check needsRender
		a.hooksCtx.FlushLayoutEffects()
		a.hooksCtx.FlushEffects()

		if atomic.LoadInt32(&a.needsRender) == 0 {
			break
		}

		// Prevent CPU pegging if something is constantly updating
		time.Sleep(1 * time.Millisecond)
	}
}

// RunHeadless renders the app once and returns the output as a string.
// Useful for CI, screenshots and tests where a TTY is not available.
func (a *App) RunHeadless() (string, error) {
	// Ensure the hooks context used by this app is global
	hooks.SetContext(a.hooksCtx)
	a.hooksCtx.SetUpdateCallback(a.onUpdate)

	// Ensure hooks context is in a clean state
	a.hooksCtx.Reset()
	tree := a.root(nil)
	a.hooksCtx.Finalize()

	// Flush effects even in headless mode
	a.hooksCtx.FlushLayoutEffects()
	a.hooksCtx.FlushEffects()

	a.treeState.Update(tree)

	output := a.renderer.Render(tree)
	return output, nil
}

func (a *App) Exit() {
	logDebug("APP", "Exit() called")

	if atomic.LoadInt32(&a.exiting) == 1 {
		return
	}
	atomic.StoreInt32(&a.exiting, 1)
	atomic.StoreInt32(&a.running, 0)
	a.hooksCtx.Cleanup()
	a.teardownTerminal()

	// Stop debug server if it was started
	if a.debug {
		debug.StopDebugServer()
	}

	logDebug("APP", "Exiting now")
	if logFile != nil {
		logFile.Close()
	}

	os.Exit(0)
}

func (a *App) Scroll(delta int) {
	newScrollY := atomic.LoadInt64(&a.scrollY) + int64(delta)

	// Clamp scroll
	contentHeight := a.renderer.GetContentHeight()
	_, termHeight := a.renderer.GetSize()

	maxScroll := int64(contentHeight - termHeight)
	if maxScroll < 0 {
		maxScroll = 0
	}

	if newScrollY < 0 {
		newScrollY = 0
	}
	if newScrollY > maxScroll {
		newScrollY = maxScroll
	}

	if newScrollY != atomic.LoadInt64(&a.scrollY) {
		atomic.StoreInt64(&a.scrollY, newScrollY)
		a.renderer.SetViewScroll(0, int(newScrollY))
		a.onUpdate()
	}
}

func (a *App) enterAltScreen() {
	if a.altScreen {
		return
	}
	fmt.Print("\x1b[?1049h") // Enter Alt Screen
	fmt.Print("\x1b[?25l")   // Hide Cursor
	fmt.Print("\x1b[?1000h") // Enable Mouse Click Reporting
	fmt.Print("\x1b[?1003h") // Enable Mouse Movement Reporting (Optional, maybe too noisy?)
	// Let's stick to 1000 (Click) and 1006 (SGR Ext) for now. 1003 is Any Event (motion).
	// User asked for "activation with mouse", implying click.
	// But dragging might be useful.
	// Let's use 1002 (Drag) or just 1000.
	fmt.Print("\x1b[?1006h") // Enable SGR Mouse Mode
	os.Stdout.Sync()
	a.altScreen = true
}

func (a *App) leaveAltScreen() {
	if !a.altScreen {
		return
	}
	fmt.Print("\x1b[?25h")   // Show Cursor
	fmt.Print("\x1b[?1049l") // Leave Alt Screen
	fmt.Print("\x1b[?1000l") // Disable Mouse Click Reporting
	fmt.Print("\x1b[?1003l") // Disable Mouse Movement Reporting
	fmt.Print("\x1b[?1006l") // Disable SGR Mouse Mode
	os.Stdout.Sync()
	a.altScreen = false
}
