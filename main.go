package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/jiiihpeeh/recliner/app"
	"github.com/jiiihpeeh/recliner/c"
	"github.com/jiiihpeeh/recliner/components"
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/vdom"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	defaultWidth    = 50
	innerBoxWidth   = 46
	magicMultiplier = 37
	magicModulus    = 100
	chuckNorrisAPI  = "https://api.chucknorris.io/jokes/random?i=%d"
)

type ChuckNorrisJoke struct {
	ID    string `json:"id"`
	Value string `json:"value"`
	URL   string `json:"url"`
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug mode with socket-based logging")
	flag.Parse()

	components.Register()

	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		return createApp(props, *debug)
	}, app.AppOptions{
		Debug: *debug,
	})

	if err := appInstance.Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}

func createApp(props any, debugMode bool) vdom.Node {
	hooksCtx := hooks.GetContext()

	w, _ := hooks.UseWindowSize(hooksCtx)
	if w < 10 {
		w = 10
	}
	rootWidth := w - 4
	// Available width inside the root box (which has padding 1 and border 1 on each side)
	innerRootWidth := rootWidth - 4
	if innerRootWidth < 1 {
		innerRootWidth = 1
	}

	count, setCount := hooks.UseState[int](hooksCtx, 0)
	randomMode, setRandomMode := hooks.UseState[bool](hooksCtx, false)

	effectMsg, _ := hooks.UseState[string](hooksCtx, "Active")
	sysStats := UseSystemStats(hooksCtx)
	inputVal, setInputVal := hooks.UseState[string](hooksCtx, "")
	inputVal2, setInputVal2 := hooks.UseState[string](hooksCtx, "")
	cmdInput, setCmdInput := hooks.UseState[string](hooksCtx, "ls -la")
	cmdOutput, setCmdOutput := hooks.UseState[string](hooksCtx, "")
	focusMgr := hooksCtx.UseFocusManager()
	showMenu, setShowMenu := hooks.UseState[bool](hooksCtx, false)
	showModal, setShowModal := hooks.UseState[bool](hooksCtx, false)
	menuX, setMenuX := hooks.UseState[int](hooksCtx, 0)
	menuY, setMenuY := hooks.UseState[int](hooksCtx, 0)
	refreshJoke, setRefreshJoke := hooks.UseState[bool](hooksCtx, false)
	lastKeys, setLastKeys := hooks.UseState[[]string](hooksCtx, []string{})

	// Radio group states
	radioValH, setRadioValH := hooks.UseState[string](hooksCtx, "opt1")
	radioValV, setRadioValV := hooks.UseState[string](hooksCtx, "opt1")

	// Checkbox states
	check1, setCheck1 := hooks.UseState[bool](hooksCtx, true)
	check2, setCheck2 := hooks.UseState[bool](hooksCtx, false)
	check3, setCheck3 := hooks.UseState[bool](hooksCtx, false)

	// Checkbox group state
	checkboxGroupOpts, setCheckboxGroupOpts := hooks.UseState[[]c.CheckboxOption](hooksCtx, []c.CheckboxOption{
		{ID: "opt1", Label: "Option 1", Value: true},
		{ID: "opt2", Label: "Option 2", Value: false},
		{ID: "opt3", Label: "Option 3", Value: false},
		{ID: "opt4", Label: "Option 4", Value: true},
	})

	// Modal input states
	modalUsername, setModalUsername := hooks.UseState[string](hooksCtx, "")
	modalEmail, setModalEmail := hooks.UseState[string](hooksCtx, "")

	// Navbar state
	activeNav, setActiveNav := hooks.UseState[string](hooksCtx, "home")

	// Pre-create inputs to ensure hook order stability even if they are not rendered
	nameInput := c.Input(c.InputProps{ID: "input1", Value: inputVal, OnChange: func(s string) { setInputVal(s) }, Placeholder: "Enter name...", Width: 30, BorderStyle: c.BorderStyleSingle, BorderColor: "white"})
	emailInput := c.Input(c.InputProps{ID: "input2", Value: inputVal2, OnChange: func(s string) { setInputVal2(s) }, Placeholder: "Enter email...", Width: 30, BorderStyle: c.BorderStyleSingle, BorderColor: "white"})
	cmdInField := c.Input(c.InputProps{ID: "cmd-in", Value: cmdInput, OnChange: setCmdInput, Width: 20, BorderStyle: c.BorderStyleSingle, BorderColor: "white"})

	jokeRes := hooks.UseFetch[ChuckNorrisJoke](hooksCtx, fmt.Sprintf("https://api.chucknorris.io/jokes/random?t=%v", refreshJoke))
	textBoxValue, _ := hooks.UseState[string](hooksCtx, "")
	if jokeRes.Loading && jokeRes.Data.Value == "" {
		textBoxValue = "Loading chuck norris joke..."
	} else if jokeRes.Error != nil {
		textBoxValue = fmt.Sprintf("Error loading joke: %v", jokeRes.Error)
	} else {
		textBoxValue = jokeRes.Data.Value
	}

	hooksCtx.UseInput(func(key events.KeyPressEvent) {
		app.DebugLog("GLOBAL_INPUT", "Key: "+key.Key)
		focusMgr.HandleKey(key)

		// Track last 5 keys for Hooks Context demo
		newKeys := append([]string{key.Key}, lastKeys...)
		if len(newKeys) > 5 {
			newKeys = newKeys[:5]
		}
		setLastKeys(newKeys)

		switch key.Key {
		case "up":
			setCount(count + 1)
		case "down":
			setCount(count - 1)
		case "r":
			setRandomMode(!randomMode)
		}
	}, []any{count, randomMode, lastKeys})

	return c.Box(c.BoxProps{
		BorderStyle: c.BorderStyleRound,
		BorderColor: "magenta",
		Padding:     1,
		Style:       c.StyleProps{Width: rootWidth, Background: "black"},
		OnClick: func(e events.MouseEvent) {
			if e.Action == events.MouseActionPress {
				if e.Button == events.MouseButtonRight {
					setMenuX(e.ScreenX)
					setMenuY(e.ScreenY)
					setShowMenu(true)
				} else {
					// Clear focus on background click
					focusMgr.Blur()
				}
			}
		},
	},
		c.Navbar(c.NavbarProps{
			Title:    "RECLINER",
			ActiveID: activeNav,
			OnSelect: setActiveNav,
			Items: []c.NavItem{
				{ID: "home", Label: "Home"},
				{ID: "about", Label: "About"},
				{ID: "contact", Label: "Contact"},
			},
		}),
		c.Spacer(1),
		c.Text(c.TextProps{Content: " Active: " + activeNav + " ", Style: c.TextStyle().BG("gray").Color("black")}),
		c.Spacer(1),
		c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleNone,
		},
			c.Text(c.TextProps{Content: " [ Welcome to the next-gen TrueColor TUI framework ] ", Style: c.TextStyle().Bold().Gradient("#00ffff", "#ff00ff")}),
		),
		c.Spacer(1),
		c.Text(c.TextProps{Content: fmt.Sprintf(" Status: %s ", effectMsg), Style: c.TextStyle().BG("gray").Color("black")}),
		c.Spacer(1),
		c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleClassic, BorderColor: "cyan",
			Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionRow, Gap: 2, Background: "black"},
		},
			c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Background: "red"}},
				c.Text(c.TextProps{Content: fmt.Sprintf(" Count: %d ", count), Style: c.TextStyle().Bold().Color("white")}),
			),
			c.Text(c.TextProps{Content: " Mode: " + func() string {
				if randomMode {
					return "RANDOM"
				}
				return "NORMAL"
			}() + " ", Style: c.TextStyle().BG("yellow").Color("black").Bold()}),
		),
		c.Spacer(1),
		c.Text(c.TextProps{Content: " SYSTEM METRICS ", Style: c.TextStyle().Bold().BG("white").Color("black")}),
		c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleSingle, BorderColor: "blue", Padding: 1,
			Style: c.StyleProps{Display: c.DisplayFlex, Gap: 2, Background: "black"},
		},
			c.Box(c.BoxProps{BorderStyle: c.BorderStyleSingle, Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionColumn, AlignItems: c.AlignItemsCenter, Width: 15}},
				c.Text(c.TextProps{Content: "CPU USAGE", Style: c.TextStyle().Dim()}),
				c.Text(c.TextProps{Content: fmt.Sprintf("%.1f%%", sysStats.CPUUsage), Style: c.TextStyle().Bold().Color("greenBright")}),
			),
			c.Box(c.BoxProps{BorderStyle: c.BorderStyleSingle, Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionColumn, AlignItems: c.AlignItemsCenter, Width: 15}},
				c.Text(c.TextProps{Content: "MEM USAGE", Style: c.TextStyle().Dim()}),
				c.Text(c.TextProps{Content: fmt.Sprintf("%.1f%%", sysStats.MemUsage), Style: c.TextStyle().Bold().Color("yellowBright")}),
			),
		),
		c.Spacer(1),
		c.Text(c.TextProps{Content: " INTERACTIVE COMPONENTS ", Style: c.TextStyle().Bold().BG("white").Color("black")}),
		c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleDouble, BorderColor: "white", Padding: 1,
			Style: c.StyleProps{
				Display:       c.DisplayFlex,
				FlexDirection: c.FlexDirectionColumn,
				Gap:           1,
				Background:    "black",
				BorderGradient: &c.RadialGradient{
					From:    "#ffff00",
					To:      "#ff0000",
					CenterX: 0.5,
					CenterY: 0.5,
					Radius:  1.5,
				},
			},
		},
			c.Tabs(c.TabsProps{
				ID:            "main-tabs",
				DefaultActive: "tab-joke",
				Variant:       c.TabsVariantUnicode,
				Items: []c.TabItem{

					{
						ID: "tab-joke", Title: "Daily Joke",
						Content: c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionColumn, Gap: 1}},
							c.Box(c.BoxProps{
								BorderStyle: c.BorderStyleNone,
								Padding:     0,
								Style: c.StyleProps{
									Display:       c.DisplayFlex,
									Gap:           1,
									FlexDirection: c.FlexDirectionRow,
									AlignItems:    c.AlignItemsCenter,
								},
							},
								c.TextBox(c.TextBoxProps{Value: textBoxValue, Placeholder: "Loading...", Width: innerRootWidth - 25, Height: 5, ReadOnly: true, Scrollable: true, ID: "textbox"}),
								c.Button(c.ButtonProps{Label: "Refresh", Variant: c.ButtonVariantFilled, Style: c.ButtonStylePrimary, OnClick: func(e events.MouseEvent) { setRefreshJoke(!refreshJoke) }, ID: "refresh-joke-button"}),
								c.Button(c.ButtonProps{Label: "Settings", Variant: c.ButtonVariantFilled, Style: c.ButtonStyleSecondary, OnClick: func(e events.MouseEvent) { setShowModal(true) }, ID: "modal-button"}),
							),
							c.Spacer(1),
							c.Text(c.TextProps{Content: "Loading Progress:", Style: c.TextStyle().Dim()}),
							c.ProgressBar(c.ProgressBarProps{
								Value: func() float64 {
									if jokeRes.Loading {
										return 0.3
									}
									return 1.0
								}(),
								Width:         innerRootWidth - 10,
								LabelPosition: "inside-right",
								FromColor:     "cyan",
								ToColor:       "blue",
							}),
						),
					},
					{
						ID:    "tab-image",
						Title: "Image Demo",
						Content: c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionColumn, AlignItems: c.AlignItemsCenter, Gap: 1}},
							c.Text(c.TextProps{Content: "Go Gopher (via Half-Blocks):", Style: c.TextStyle().Bold()}),
							c.Image(c.ImageProps{
								Src:    "https://go.dev/blog/gopher/gopher.png",
								Width:  30,
								Height: 15,
							}),
						),
					},
					{
						ID: "tab-inputs", Title: "Form Inputs",
						Content: c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionColumn, Gap: 1}},
							c.Text(c.TextProps{Content: "Command Input (Testing):"}),
							cmdInField,
							c.Spacer(1),
							c.Text(c.TextProps{Content: "Name:"}),
							nameInput,
							c.Text(c.TextProps{Content: "Email:"}),
							emailInput,
							c.Spacer(1),
							c.Text(c.TextProps{Content: "Select an option (Horizontal):", Style: c.TextStyle().Bold()}),
							c.RadioGroup(c.RadioGroupProps{
								ID:        "radio-group-h",
								Value:     radioValH,
								OnChange:  setRadioValH,
								Direction: c.FlexDirectionRow,
								Gap:       2,
								Options: []c.RadioOption{
									{Label: "Option 1", Value: "opt1"},
									{Label: "Option 2", Value: "opt2"},
									{Label: "Option 3", Value: "opt3"},
								},
							}),
							c.Text(c.TextProps{Content: "Selected: " + radioValH, Style: c.TextStyle().Color("gray")}),

							c.Spacer(1),

							c.Text(c.TextProps{Content: "Select an option (Vertical):", Style: c.TextStyle().Bold()}),
							c.RadioGroup(c.RadioGroupProps{
								ID:        "radio-group-v",
								Value:     radioValV,
								OnChange:  setRadioValV,
								Direction: c.FlexDirectionColumn,
								Options: []c.RadioOption{
									{Label: "First Choice", Value: "opt1"},
									{Label: "Second Choice", Value: "opt2"},
									{Label: "Third Choice", Value: "opt3"},
								},
							}),
							c.Text(c.TextProps{Content: "Selected: " + radioValV, Style: c.TextStyle().Color("gray")}),
							c.Spacer(1),
							c.Text(c.TextProps{Content: "Checkboxes:", Style: c.TextStyle().Bold()}),
							c.CheckBox(c.CheckBoxProps{
								Label:    c.Text(c.TextProps{Content: "Enable notifications"}),
								Checked:  check1,
								OnChange: setCheck1,
								ID:       "cb1",
							}),
							c.CheckBox(c.CheckBoxProps{
								Label:    c.Text(c.TextProps{Content: "Receive weekly newsletter"}),
								Checked:  check2,
								OnChange: setCheck2,
								ID:       "cb2",
							}),
							c.CheckBox(c.CheckBoxProps{
								Label:    c.Text(c.TextProps{Content: "I agree to terms and conditions"}),
								Checked:  check3,
								OnChange: setCheck3,
								ID:       "cb3",
							}),
							c.Text(c.TextProps{Content: fmt.Sprintf("Selected: %v, %v, %v", check1, check2, check3), Style: c.TextStyle().Color("gray")}),
							c.Spacer(1),
							c.Text(c.TextProps{Content: "Checkbox Group (Tri-state):", Style: c.TextStyle().Bold()}),
							c.CheckboxGroup(c.CheckboxGroupProps{
								Label:    "Select Options",
								Options:  checkboxGroupOpts,
								OnChange: setCheckboxGroupOpts,
							}),
						),
					},
					{
						ID: "tab-hooks", Title: "Hooks Context",
						Content: c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionColumn, Gap: 1}},
							c.Text(c.TextProps{Content: "Standard Handle File Descriptors:", Style: c.TextStyle().Bold().Color("cyan")}),
							c.Text(c.TextProps{Content: fmt.Sprintf("• Stdin:  %d (present: %v)", hooksCtx.Stdin.Fd(), hooksCtx.Stdin != nil)}),
							c.Text(c.TextProps{Content: fmt.Sprintf("• Stdout: %d (present: %v)", hooksCtx.Stdout.Fd(), hooksCtx.Stdout != nil)}),
							c.Text(c.TextProps{Content: fmt.Sprintf("• Stderr: %d (present: %v)", hooksCtx.Stderr.Fd(), hooksCtx.Stderr != nil)}),
							c.Spacer(1),
							c.Text(c.TextProps{Content: "Terminal Info:", Style: c.TextStyle().Bold().Color("cyan")}),
							c.Text(c.TextProps{Content: fmt.Sprintf("• Dimensions: %d x %d", hooksCtx.WindowWidth, hooksCtx.WindowHeight)}),
							c.Text(c.TextProps{Content: fmt.Sprintf("• App present: %v", hooksCtx.App != nil)}),
							c.Spacer(1),
							c.Text(c.TextProps{Content: "Keystroke Test (Stdin):", Style: c.TextStyle().Bold().Color("cyan")}),
							c.Text(c.TextProps{Content: fmt.Sprintf("• Last 5 Keys: %v", strings.Join(lastKeys, ", "))}),
							c.Spacer(1),
							c.Button(c.ButtonProps{
								Label: "Log to Stderr",
								OnClick: func(e events.MouseEvent) {
									if e.Action == events.MouseActionPress {
										fmt.Fprintln(hooksCtx.Stderr, "Manual log from hooks context to stderr at "+time.Now().Format("15:04:05"))
										app.DebugLog("HOOKS_DEMO", "Wrote to Stderr")
									}
								},
							}),
						),
					},
					{
						ID: "tab-terminal", Title: "Terminal Capture",
						Content: c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionColumn, Gap: 1}},
							c.Text(c.TextProps{Content: "Subprocess Output Capture:", Style: c.TextStyle().Bold().Color("yellow")}),
							c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionRow, Gap: 1, AlignItems: c.AlignItemsCenter}},
								c.Text(c.TextProps{Content: "Cmd:"}),
								cmdInField,
								c.Button(c.ButtonProps{
									Label: "Run",
									OnClick: func(e events.MouseEvent) {
										if e.Action == events.MouseActionPress {
											setCmdOutput("Running...")
											go func() {
												args := strings.Split(cmdInput, " ")
												cmd := exec.Command(args[0], args[1:]...)
												out, err := cmd.CombinedOutput()
												outputStr := string(out)
												if len(outputStr) > 10000 {
													outputStr = outputStr[:10000] + "\n... (truncated)"
												}
												if err != nil {
													setCmdOutput(fmt.Sprintf("Error: %v\n%s", err, outputStr))
												} else {
													setCmdOutput(outputStr)
												}
											}()
										}
									},
								}),
							),
							c.Spacer(1),
							c.TextBox(c.TextBoxProps{
								ID: "cmd-out", Value: cmdOutput, Width: innerRootWidth - 6, Height: 10,
								ReadOnly: true, Scrollable: true,
							}),
						),
					},
				},
			}),
		),
		c.Spacer(1),
		c.Accordion(c.AccordionProps{
			ID:      "info-accordion",
			Variant: c.AccordionVariantUnicode,
			Items: []c.AccordionItem{
				{
					ID:    "shortcuts",
					Title: "Shortcuts",
					Content: c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone},
						c.Text(c.TextProps{Content: "• Up/Down: Count", Style: c.TextStyle().Dim()}),
						c.Text(c.TextProps{Content: "• [R]: Random Mode", Style: c.TextStyle().Dim()}),
					),
				},
				{
					ID:    "about",
					Title: "About",
					Content: c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone},
						c.Text(c.TextProps{Content: "About", Style: c.TextStyle().Bold().Color("cyan")}),
						c.Link(c.LinkProps{
							Href:  "https://github.com/jiiihpeeh/recliner",
							Style: c.StyleProps{MarginTop: 1},
						},
							c.Text(c.TextProps{Content: "Visit GitHub Repository", Style: c.TextStyle().Color("blue").Underline()}),
						),
					),
				},
			},
		}),
		c.Spacer(1),
		c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Width: innerRootWidth, Background: "gray"}},
			c.Text(c.TextProps{Content: " [Tab] Focus | [R] Random | [PageUp/Dn] Fib | [RightClick] Menu ", Style: c.TextStyle().Color("black")}),
		),

		func() vdom.Node {
			if showMenu {
				_, h := hooks.UseWindowSize(hooksCtx)
				return c.Box(c.BoxProps{
					Style: c.StyleProps{Position: c.PositionFixed, Top: 0, Left: 0, Width: rootWidth, Height: h},
					OnClick: func(e events.MouseEvent) {
						if e.Action == events.MouseActionPress {
							setShowMenu(false)
						}
					},
				})
			}
			return nil
		}(),
		func() vdom.Node {
			if showMenu {
				return c.Menu(c.MenuProps{
					Title: "Main Menu", Items: []string{"Resume", "Settings", "Quit"},
					X: menuX, Y: menuY,
					OnClose:  func() { setShowMenu(false) },
					OnSelect: func(item string) { setShowMenu(false) },
				})
			}
			return nil
		}(),
		func() vdom.Node {
			if showModal {
				hooksCtx.UseEffect(func() func() {
					hooksCtx.UseFocusManager().Focus("modal-username")
					return nil
				}, []any{showModal})
				return c.Modal(c.ModalProps{
					Title:           "User Settings",
					ConfirmText:     "Save",
					CancelText:      "Cancel",
					BackgroundColor: "orange",
					Children: []vdom.Node{
						c.Text(c.TextProps{Content: "Username:"}),
						c.Input(c.InputProps{ID: "modal-username", Value: modalUsername, OnChange: setModalUsername, Placeholder: "Enter username...", Width: 25, BorderStyle: c.BorderStyleSingle, BorderColor: "white", AutoFocus: true}),
						c.Spacer(1),
						c.Text(c.TextProps{Content: "Email:"}),
						c.Input(c.InputProps{ID: "modal-email", Value: modalEmail, OnChange: setModalEmail, Placeholder: "Enter email...", Width: 25, BorderStyle: c.BorderStyleSingle, BorderColor: "white"}),
						c.Spacer(1),
						c.Text(c.TextProps{Content: "This is a sample modal dialog with form elements.", Style: c.TextStyle().Dim()}),
					},
					OnClose:   func() { setShowModal(false) },
					OnConfirm: func() { setShowModal(false) },
					OnCancel:  func() { setShowModal(false) },
				})
			}
			return nil
		}(),
	)
}

type SystemStats struct {
	CPUUsage float64
	MemUsage float64
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func UseSystemStats(hc *hooks.HooksContext) SystemStats {
	stats, setStats := hooks.UseState[SystemStats](hc, SystemStats{})
	hc.UseEffect(func() func() {
		ticker := time.NewTicker(2 * time.Second)
		done := make(chan struct{})
		updateStats := func() {
			v, _ := mem.VirtualMemory()
			c, _ := cpu.Percent(0, false)
			cpuVal := 0.0
			if len(c) > 0 {
				cpuVal = c[0]
			}
			memVal := 0.0
			if v != nil {
				memVal = v.UsedPercent
			}
			setStats(SystemStats{CPUUsage: cpuVal, MemUsage: memVal})
		}
		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					updateStats()
				}
			}
		}()
		go updateStats()
		return func() { ticker.Stop(); close(done) }
	}, []any{})
	return stats
}
