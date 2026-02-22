package utils

import (
	"bytes"
	"os/exec"
)

// WriteClipboard writes text to the system clipboard.
// It tries wl-copy (Wayland), then xclip, then xsel.
func WriteClipboard(text string) error {
	// Try Wayland
	if _, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command("wl-copy")
		cmd.Stdin = bytes.NewReader([]byte(text))
		return cmd.Run()
	}

	// Try xclip
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = bytes.NewReader([]byte(text))
		return cmd.Run()
	}

	// Try xsel
	if _, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command("xsel", "--clipboard", "--input")
		cmd.Stdin = bytes.NewReader([]byte(text))
		return cmd.Run()
	}

	return exec.ErrNotFound
}

// ReadClipboard reads text from the system clipboard.
// It tries wl-paste (Wayland), then xclip, then xsel.
func ReadClipboard() (string, error) {
	// Try Wayland
	if _, err := exec.LookPath("wl-paste"); err == nil {
		out, err := exec.Command("wl-paste", "--no-newline").Output()
		return string(out), err
	}

	// Try xclip
	if _, err := exec.LookPath("xclip"); err == nil {
		out, err := exec.Command("xclip", "-selection", "clipboard", "-o").Output()
		return string(out), err
	}

	// Try xsel
	if _, err := exec.LookPath("xsel"); err == nil {
		out, err := exec.Command("xsel", "--clipboard", "--output").Output()
		return string(out), err
	}

	return "", exec.ErrNotFound
}
