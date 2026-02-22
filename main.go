package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/j-p/recliner/app"
	"github.com/j-p/recliner/c"
	"github.com/j-p/recliner/components"
	"github.com/j-p/recliner/events"
	"github.com/j-p/recliner/hooks"
	"github.com/j-p/recliner/vdom"
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
	fibIndex, setFibIndex := hooks.UseState[int](hooksCtx, 100)

	fibValue := hooks.UseMemo(hooksCtx, func() *big.Int {
		if fibIndex <= 1 {
			return big.NewInt(int64(fibIndex))
		}
		a, b := big.NewInt(0), big.NewInt(1)
		tmp := new(big.Int)
		for i := 2; i <= fibIndex; i++ {
			tmp.Add(a, b)
			a.Set(b)
			b.Set(tmp)
		}
		return new(big.Int).Set(b)
	}, []any{fibIndex})

	effectMsg, _ := hooks.UseState[string](hooksCtx, "Active")
	sysStats := UseSystemStats(hooksCtx)
	inputVal, setInputVal := hooks.UseState[string](hooksCtx, "")
	inputVal2, setInputVal2 := hooks.UseState[string](hooksCtx, "")
	focusMgr := hooksCtx.UseFocusManager()
	showMenu, setShowMenu := hooks.UseState[bool](hooksCtx, false)
	menuX, setMenuX := hooks.UseState[int](hooksCtx, 0)
	menuY, setMenuY := hooks.UseState[int](hooksCtx, 0)
	refreshJoke, setRefreshJoke := hooks.UseState[bool](hooksCtx, false)

	jokeRes := hooks.UseFetch[ChuckNorrisJoke](hooksCtx, fmt.Sprintf(chuckNorrisAPI, boolToInt(refreshJoke)))
	textBoxValue, _ := hooks.UseState[string](hooksCtx, "")
	if jokeRes.Loading {
		textBoxValue = "Loading..."
	} else if jokeRes.Error != nil {
		textBoxValue = "Error"
	} else {
		textBoxValue = jokeRes.Data.Value
	}

	hooksCtx.UseInput(func(key events.KeyPressEvent) {
		focusMgr.HandleKey(key)
		switch key.Key {
		case "up":
			setCount(count + 1)
		case "down":
			setCount(count - 1)
		case "pageup":
			setFibIndex(fibIndex + 1)
		case "pagedown":
			if fibIndex > 0 {
				setFibIndex(fibIndex - 1)
			}
		case "r":
			setRandomMode(!randomMode)
		case "tab":
			if key.Shift {
				focusMgr.FocusPrev()
			} else {
				focusMgr.FocusNext()
			}
		}
	}, []any{count, randomMode, fibIndex})

	return c.Box(c.BoxProps{
		BorderStyle: c.BorderStyleRound,
		BorderColor: "magenta",
		Padding:     1,
		Style:       c.StyleProps{Width: rootWidth, Background: "black"},
		OnClick: func(e events.MouseEvent) {
			if e.Action == events.MouseActionPress && e.Button == events.MouseButtonRight {
				setMenuX(e.ScreenX)
				setMenuY(e.ScreenY)
				setShowMenu(true)
			}
		},
	},
		c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleNone,
			Style:       c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionRow, JustifyContent: c.JustifyContentSpaceBetween, Background: "blue"},
		},
			c.Text(c.TextProps{Content: " RECLINER DASHBOARD ", Style: c.TextStyle().Bold().Color("white")}),
			c.Text(c.TextProps{Content: time.Now().Format(" 15:04:05 "), Style: c.TextStyle().Color("whiteBright")}),
		),
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
			BorderStyle: c.BorderStyleSingle, BorderColor: "cyan",
			Style: c.StyleProps{Display: c.DisplayFlex, FlexDirection: c.FlexDirectionRow, Gap: 2, Background: "black"},
		},
			c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Background: "red"}},
				c.Text(c.TextProps{Content: fmt.Sprintf(" Count: %d ", count), Style: c.TextStyle().Bold().Color("white")}),
			),
			c.Box(c.BoxProps{BorderStyle: c.BorderStyleNone, Style: c.StyleProps{Background: "green"}},
				c.Text(c.TextProps{Content: fmt.Sprintf(" Fib: %s ", fibValue.String()), Style: c.TextStyle().Bold().Color("black")}),
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
				ID: "main-tabs",
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
							c.Text(c.TextProps{Content: "Name:"}),
							c.Input(c.InputProps{ID: "input1", Value: inputVal, OnChange: func(s string) { setInputVal(s) }, Placeholder: "Enter name...", Width: 30}),
							c.Text(c.TextProps{Content: "Email:"}),
							c.Input(c.InputProps{ID: "input2", Value: inputVal2, OnChange: func(s string) { setInputVal2(s) }, Placeholder: "Enter email...", Width: 30}),
						),
					},
				},
			}),
		),
		c.Spacer(1),
		c.Accordion(c.AccordionProps{
			ID: "info-accordion",
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
						c.Text(c.TextProps{Content: "ReCLIner TUI Framework", Style: c.TextStyle().Bold().Color("cyan")}),
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
		ticker := time.NewTicker(1 * time.Second)
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
