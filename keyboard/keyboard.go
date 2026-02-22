package keyboard

// Mapping represents a key->action mapping.
type Mapping map[string]string

// DefaultMapping returns a sensible default mapping.
func DefaultMapping() Mapping {
	return Mapping{
		"up":    "up",
		"down":  "down",
		"left":  "left",
		"right": "right",
		"enter": "enter",
		"esc":   "escape",
	}
}
