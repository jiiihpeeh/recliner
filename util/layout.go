package util

// CenterPad returns a string padded to width centered around s.
func CenterPad(s string, width int) string {
	if width <= len(s) {
		return s
	}
	pad := width - len(s)
	left := pad / 2
	right := pad - left
	leftPad := make([]byte, left)
	rightPad := make([]byte, right)
	for i := range leftPad {
		leftPad[i] = ' '
	}
	for i := range rightPad {
		rightPad[i] = ' '
	}
	return string(leftPad) + s + string(rightPad)
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func StableSort[T any](items []T, less func(i, j int) bool) {
	if len(items) <= 1 {
		return
	}
	// Simple bubble sort for stable sorting in a library/utility context
	// Usually TUI trees are not that deep/wide to require more complex logic here
	for i := 0; i < len(items)-1; i++ {
		for j := 0; j < len(items)-i-1; j++ {
			if less(j+1, j) {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}
}
