package hooks

// UseCursor registers a request to set cursor visibility and position for this render pass.
// If multiple components call this, the last one called will likely determine the cursor state unless there's logic to handle precedence.
// Currently simpler: last write wins.
func (hc *HooksContext) UseCursor(visible bool, x, y int) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	// Only update if visible is true, or if we want to explicitly hide it?
	// If multiple components try to control the cursor, it's tricky.
	// We assume only the active/focused component calls UseCursor.
	// Therefore, we trust the caller.
	hc.CursorVisible = visible
	if visible {
		hc.CursorX = x
		hc.CursorY = y
	}
}
