package main

import (
	"fmt"
	"time"

	"github.com/j-p/recliner/app"
	"github.com/j-p/recliner/c"
	"github.com/j-p/recliner/hooks"
	"github.com/j-p/recliner/vdom"
)

func main() {
	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		hc := hooks.GetContext()

		// Pulse animation (auto-starts)
		pulse := hooks.UseAnimation(hc, true, hooks.AnimationOptions{
			Duration: 2 * time.Second,
			Easing:   hooks.EasingInOutQuad,
		})

		// Movement animation
		slide := hooks.UseAnimation(hc, true, hooks.AnimationOptions{
			Duration: 1 * time.Second,
			Easing:   hooks.EasingBounce,
			Delay:    500 * time.Millisecond,
		})

		// Interpolated values
		width := hooks.LerpInt(10, 50, pulse)
		left := hooks.LerpInt(0, 30, slide)

		// Color transition simulation
		bgColor := "blue"
		if pulse > 0.5 {
			bgColor = "magenta"
		}

		return c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleSingle,
			Style:       c.StyleProps{Width: 80, Height: 20, Background: "black"},
		},
			c.Box(c.BoxProps{
				BorderStyle: c.BorderStyleDouble,
				BorderColor: "white",
				Style: c.StyleProps{
					Position:   c.PositionAbsolute,
					Top:        5,
					Left:       left,
					Width:      width,
					Height:     5,
					Background: bgColor,
				},
			},
				c.Text(c.TextProps{Content: fmt.Sprintf("Progress: %.2f", pulse)}),
			),
			c.Box(c.BoxProps{
				Style: c.StyleProps{Position: c.PositionAbsolute, Top: 12, Left: 5},
			},
				c.Text(c.TextProps{Content: "This demo shows:\n1. Width pulsing (InOutQuad)\n2. Bouncing entrance (Bounce + Delay)\n3. State-based color switching"}),
			),
		)
	}, app.AppOptions{Debug: true}) // Use debug to see frames in logs

	// Since we can't easily see continuous animation in a headless run,
	// we'll simulate a few frames.
	fmt.Println("Simulating Animation Frames...")
	for i := 0; i < 5; i++ {
		out, _ := appInstance.RunHeadless()
		fmt.Printf("Frame %d:\n%s\n", i, out)
		time.Sleep(200 * time.Millisecond)
	}
}
