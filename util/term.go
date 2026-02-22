package util

import (
	"os"
	"strings"
)

type TermCapabilities struct {
	TrueColor bool
	Sixel     bool
	Kitty     bool
}

func GetTermCapabilities() TermCapabilities {
	term := strings.ToLower(os.Getenv("TERM"))
	colorterm := strings.ToLower(os.Getenv("COLORTERM"))

	caps := TermCapabilities{
		TrueColor: colorterm == "truecolor" || colorterm == "24bit" || term == "xterm-256color", // basic heuristic
	}

	// Sixel detection (heuristic)
	if strings.Contains(term, "xterm") || strings.Contains(term, "mlterm") || strings.Contains(term, "foot") {
		caps.Sixel = true
	}

	// Kitty detection
	if term == "xterm-kitty" || os.Getenv("KITTY_WINDOW_ID") != "" {
		caps.Kitty = true
	}

	return caps
}
