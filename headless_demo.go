//go:build demo
// +build demo

package main

import (
	"fmt"

	"github.com/jiiihpeeh/recliner/app"
	"github.com/jiiihpeeh/recliner/components"
	"github.com/jiiihpeeh/recliner/vdom"
)

var longText = `Line 1: The quick brown fox jumps over the lazy dog.
Line 2: Pack my box with five dozen liquor jugs.
Line 3: How vexingly quick daft zebras jump!
Line 4: Sphinx of black quartz, judge my vow.
Line 5: The five boxing wizards jump quickly.
Line 6: Jackdaws love my big sphinx of quartz.
Line 7: Waltz, bad nymph, for quick jigs vex.
Line 8: Quick zephyrs blow, vexing daft Jim.
Line 9: Two driven jocks help fax my big quiz.
Line 10: Bright vixens jump; dozy fowl quack.
` // 10 lines

func main() {
	components.Register()

	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		// Render a scrollable TextBox with width 30 and height 6
		return components.TextBox(map[string]any{
			"value":      longText,
			"width":      30,
			"height":     6,
			"scrollable": true,
		})
	}, app.AppOptions{Headless: true, NoAlt: true})

	out, err := appInstance.RunHeadless()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(out)
}
