package main

import (
	"fmt"
	"strings"

	"github.com/jiiihpeeh/recliner/render"
	"github.com/jiiihpeeh/recliner/vdom"
)

func main() {
	r := render.NewRenderer()

	// Create a text element with style
	el := &vdom.Element{
		Type:      "text",
		InnerText: "Hello Red World",
		Style: vdom.Style{
			Foreground: "red",
		},
	}

	// Create a dummy node with children to trigger drawElement default path
	// But first let's test the text element directly

	output := r.Render(el)
	// We expect ANSI codes for red
	if strings.Contains(output, "31") {
		fmt.Println("Success: Contains red color code")
	} else {
		fmt.Println("Failure: Does not contain red color code")
	}
	fmt.Printf("Output: %q\n", output)
}
