# ReCLIner

ReCLIner is a next-generation, stateful TrueColor TUI (Terminal User Interface) framework for Go, inspired by React and Ink. It allows you to build complex, interactive terminal applications using declarative components and hooks.

## Features

- **Declarative Components**: Build your UI using a tree of components.
- **Hook-based State Management**: Use `UseState`, `UseEffect`, `UseMemo`, and `UseRef` just like in React.
- **TrueColor Support**: Full 24-bit RGB color support for backgrounds, foregrounds, and gradients.
- **Robust Rendering**: Optimized diff-based rendering engine to minimize terminal flickering and maximize performance.
- **Flexible Layout**: Supports block, inline, and Flexbox-inspired layouts with absolute and fixed positioning.
- **Interactive Components**: Built-in support for Inputs, TextBoxes (with word-wrapping), Buttons, Tabs, Accordions, Progress Bars, and more.
- **Global Viewport Scrolling**: Built-in support for scrolling large application canvases via mouse wheel or keyboard.
- **Mouse & Keyboard Support**: Comprehensive event handling system with hit-testing and input capture.

## Components

ReCLIner comes with a variety of built-in components:

- `Box`: The primary layout container. Supports borders, padding, and flex layouts.
- `Text`: Declarative text rendering with full styling and gradients.
- `Input`: Single-line text input with cursor management and selection.
- `TextBox`: Multi-line text area with automatic word-wrapping and scrollbar support.
- `Tabs`: Easy-to-use tabbed interface for switching between views.
- `ProgressBar`: Customizable progress indicator with label support.
- `Image`: High-quality terminal image rendering using half-blocks and Lanczos3 resizing.
- `Menu`: Context-aware popup menus.

## Getting Started

### Installation

```bash
go get github.com/j-p/recliner
```

### Basic Example

```go
package main

import (
	"github.com/j-p/recliner/app"
	"github.com/j-p/recliner/c"
	"github.com/j-p/recliner/hooks"
	"github.com/j-p/recliner/vdom"
)

func main() {
	appInstance := app.New(func(props any) vdom.Node {
		hc := hooks.GetContext()
		count, setCount := hooks.UseState(hc, 0)

		return c.Box(c.BoxProps{
			Padding: 1,
			BorderStyle: c.BorderStyleRound,
		},
			c.Text(c.TextProps{Content: "Hello, ReCLIner!"}),
			c.Button(c.ButtonProps{
				Label: "Click Me",
				OnClick: func(e events.MouseEvent) {
					if e.Action == events.MouseActionPress {
						setCount(count + 1)
					}
				},
			}),
			c.Text(c.TextProps{Content: fmt.Sprintf("Count: %d", count)}),
		)
	})

	if err := appInstance.Run(); err != nil {
		panic(err)
	}
}
```

## Dashboard Demo

The project includes a comprehensive dashboard demo (`main.go`) that showcases:
- System metrics (CPU/MEM usage).
- Chuck Norris joke API integration.
- Subprocess output capture.
- Complex nested layouts and tab navigation.

To run the demo:
```bash
go run main.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
