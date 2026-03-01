package hooks

import (
	"slices"

	"github.com/jiiihpeeh/recliner/events"
)

type FocusOptions struct {
	ID        string
	AutoFocus bool
}

type FocusResult struct {
	IsFocused bool
	Focus     func()
	Blur      func()
}

// UseFocus registers a component as focusable and returns its focus state.
// If options.ID is empty, a random ID is generated (not implemented here, user must provide ID for stability).
func (hc *HooksContext) UseFocus(opts FocusOptions) FocusResult {
	hc.mu.Lock()
	// Register this ID for the current render pass so manager knows it exists
	if opts.ID != "" {
		exists := slices.Contains(hc.FocusableIDs, opts.ID)
		if !exists {
			hc.FocusableIDs = append(hc.FocusableIDs, opts.ID)
		}
	}

	// AutoFocus logic: if no focus set, set this one
	if opts.AutoFocus && hc.FocusID == "" {
		hc.FocusID = opts.ID
	}

	isFocused := opts.ID != "" && hc.FocusID == opts.ID
	hc.mu.Unlock()

	focus := func() {
		hc.mu.Lock()
		if hc.FocusID != opts.ID {
			hc.FocusID = opts.ID
			hc.mu.Unlock() // Unlock before trigger
			hc.triggerUpdate()
		} else {
			hc.mu.Unlock()
		}
	}

	blur := func() {
		hc.mu.Lock()
		if hc.FocusID == opts.ID {
			hc.FocusID = ""
			hc.mu.Unlock() // Unlock before trigger
			hc.triggerUpdate()
		} else {
			hc.mu.Unlock()
		}
	}

	return FocusResult{
		IsFocused: isFocused,
		Focus:     focus,
		Blur:      blur,
	}
}

type FocusManager struct {
	Focus     func(id string)
	Blur      func()
	FocusNext func()
	FocusPrev func()
	HandleKey func(events.KeyPressEvent)
}

type FocusManagerOptions struct {
	ClearKey string
}

// UseFocusManager returns methods to control focus globally.
func (hc *HooksContext) UseFocusManager(opts ...FocusManagerOptions) FocusManager {
	clearKey := "ctrl+g"
	if len(opts) > 0 && opts[0].ClearKey != "" {
		clearKey = opts[0].ClearKey
	}
	return FocusManager{
		Focus: func(id string) {
			hc.mu.Lock()
			if hc.FocusID != id {
				hc.FocusID = id
				hc.mu.Unlock()
				hc.triggerUpdate()
			} else {
				hc.mu.Unlock()
			}
		},
		Blur: func() {
			hc.mu.Lock()
			if hc.FocusID != "" {
				hc.FocusID = ""
				hc.mu.Unlock()
				hc.triggerUpdate()
			} else {
				hc.mu.Unlock()
			}
		},
		FocusNext: func() {
			hc.mu.Lock()
			if len(hc.FocusableIDs) == 0 {
				hc.mu.Unlock()
				return
			}

			// find current index
			currentIndex := -1
			for i, id := range hc.FocusableIDs {
				if id == hc.FocusID {
					currentIndex = i
					break
				}
			}

			nextIndex := (currentIndex + 1) % len(hc.FocusableIDs)
			newID := hc.FocusableIDs[nextIndex]

			if hc.FocusID != newID {
				hc.FocusID = newID
				hc.mu.Unlock()
				hc.triggerUpdate()
				return
			}
			hc.mu.Unlock()
		},
		FocusPrev: func() {
			hc.mu.Lock()
			if len(hc.FocusableIDs) == 0 {
				hc.mu.Unlock()
				return
			}

			currentIndex := -1
			for i, id := range hc.FocusableIDs {
				if id == hc.FocusID {
					currentIndex = i
					break
				}
			}

			prevIndex := (currentIndex - 1 + len(hc.FocusableIDs)) % len(hc.FocusableIDs)
			newID := hc.FocusableIDs[prevIndex]

			if hc.FocusID != newID {
				hc.FocusID = newID
				hc.mu.Unlock()
				hc.triggerUpdate()
				return
			}
			hc.mu.Unlock()
		},
		HandleKey: func(ev events.KeyPressEvent) {
			// Clear focus
			if ev.Key == clearKey {
				hc.mu.Lock()
				if hc.FocusID != "" {
					hc.FocusID = ""
					hc.mu.Unlock()
					hc.triggerUpdate()
					return
				}
				hc.mu.Unlock()
				return
			}

			// Tab navigation
			if ev.Key == "tab" {
				if ev.Shift {
					// FocusPrev
					hc.mu.Lock()
					if len(hc.FocusableIDs) == 0 {
						hc.mu.Unlock()
						return
					}
					currentIndex := -1
					for i, id := range hc.FocusableIDs {
						if id == hc.FocusID {
							currentIndex = i
							break
						}
					}
					prevIndex := (currentIndex - 1 + len(hc.FocusableIDs)) % len(hc.FocusableIDs)
					newID := hc.FocusableIDs[prevIndex]
					if hc.FocusID != newID {
						hc.FocusID = newID
						hc.mu.Unlock()
						hc.triggerUpdate()
						return
					}
					hc.mu.Unlock()
				} else {
					// FocusNext
					hc.mu.Lock()
					if len(hc.FocusableIDs) == 0 {
						hc.mu.Unlock()
						return
					}
					currentIndex := -1
					for i, id := range hc.FocusableIDs {
						if id == hc.FocusID {
							currentIndex = i
							break
						}
					}
					nextIndex := (currentIndex + 1) % len(hc.FocusableIDs)
					newID := hc.FocusableIDs[nextIndex]
					if hc.FocusID != newID {
						hc.FocusID = newID
						hc.mu.Unlock()
						hc.triggerUpdate()
						return
					}
					hc.mu.Unlock()
				}
			}
		},
	}
}
