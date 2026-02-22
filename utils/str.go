package utils

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// Str is a UTF-8 aware string wrapper that behaves like JavaScript strings.
// It treats indices as rune counts rather than byte offsets.
type Str struct {
	runes []rune
}

// NewStr creates a new Str from a native Go string.
func NewStr(s string) Str {
	return Str{runes: []rune(s)}
}

// String returns the native Go string representation.
func (s Str) String() string {
	return string(s.runes)
}

// Value alias for String()
func (s Str) Value() string {
	return string(s.runes)
}

// Length returns the number of characters (runes).
// JS equivalent: .length
func (s Str) Length() int {
	return len(s.runes)
}

// At returns the character at the specified index.
// Allows negative integers to count back from the last character.
// JS equivalent: .at()
func (s Str) At(index int) string {
	l := len(s.runes)
	if index < 0 {
		index = l + index
	}
	if index < 0 || index >= l {
		return ""
	}
	return string(s.runes[index])
}

// CharAt returns the character at the specified index.
// Returns empty string if out of range.
// JS equivalent: .charAt()
func (s Str) CharAt(index int) string {
	if index < 0 || index >= len(s.runes) {
		return ""
	}
	return string(s.runes[index])
}

// Slice extracts a section of a string and returns it as a new string.
// Supported signatures: Slice(), Slice(start), Slice(start, end).
// JS equivalent: .slice()
func (s Str) Slice(args ...int) Str {
	l := len(s.runes)
	start := 0
	end := l

	if len(args) > 0 {
		start = args[0]
		if start < 0 {
			start = l + start
		}
	}

	if len(args) > 1 {
		end = args[1]
		if end < 0 {
			end = l + end
		}
	}

	// Clamp values
	if start < 0 {
		start = 0
	}
	if start > l {
		start = l
	}
	if end < 0 {
		end = 0
	}
	if end > l {
		end = l
	}

	if start >= end {
		return Str{runes: []rune{}}
	}

	return Str{runes: s.runes[start:end]}
}

// Substring returns the part of the string between the start and end indexes.
// Swaps start and end if start > end.
// JS equivalent: .substring()
func (s Str) Substring(start int, endArgs ...int) Str {
	l := len(s.runes)
	end := l
	if len(endArgs) > 0 {
		end = endArgs[0]
	}

	// Clamp first (JS substring treats negative as 0)
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start > l {
		start = l
	}
	if end > l {
		end = l
	}

	if start > end {
		start, end = end, start
	}

	return Str{runes: s.runes[start:end]}
}

// Includes determines whether the string contains the characters of a specified string.
// JS equivalent: .includes()
func (s Str) Includes(searchString string) bool {
	return strings.Contains(string(s.runes), searchString)
}

// IndexOf returns the index (rune-based) of the first occurrence of the specified value.
// Returns -1 if not found.
// JS equivalent: .indexOf()
func (s Str) IndexOf(searchValue string) int {
	searchRunes := []rune(searchValue)
	if len(searchRunes) == 0 {
		return 0
	}

	// Naive implementation on runes
	l := len(s.runes)
	sl := len(searchRunes)

	for i := 0; i <= l-sl; i++ {
		match := true
		for j := 0; j < sl; j++ {
			if s.runes[i+j] != searchRunes[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// LastIndexOf returns the index (rune-based) of the last occurrence of the specified value.
// JS equivalent: .lastIndexOf()
func (s Str) LastIndexOf(searchValue string) int {
	searchRunes := []rune(searchValue)
	if len(searchRunes) == 0 {
		return -1 // JS implementation behavior for LastIndexOf("") depends, but often length.
	}

	l := len(s.runes)
	sl := len(searchRunes)

	for i := l - sl; i >= 0; i-- {
		match := true
		for j := 0; j < sl; j++ {
			if s.runes[i+j] != searchRunes[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// ToLowerCase returns the calling string value converted to lower case.
// JS equivalent: .toLowerCase()
func (s Str) ToLowerCase() Str {
	lower := strings.ToLower(string(s.runes))
	return Str{runes: []rune(lower)}
}

// ToUpperCase returns the calling string value converted to upper case.
// JS equivalent: .toUpperCase()
func (s Str) ToUpperCase() Str {
	upper := strings.ToUpper(string(s.runes))
	return Str{runes: []rune(upper)}
}

// Trim removes whitespace from both ends of the string.
// JS equivalent: .trim()
func (s Str) Trim() Str {
	trimmed := strings.TrimSpace(string(s.runes))
	return Str{runes: []rune(trimmed)}
}

// Split divides the string into an ordered list of substrings.
// JS equivalent: .split()
func (s Str) Split(separator string) []Str {
	parts := strings.Split(string(s.runes), separator)
	result := make([]Str, len(parts))
	for i, p := range parts {
		result[i] = NewStr(p)
	}
	return result
}

// Repeat constructs and returns a new string which contains the specified number of copies.
// JS equivalent: .repeat()
func (s Str) Repeat(count int) Str {
	if count < 0 {
		return Str{runes: []rune{}} // JS throws RangeError, return empty safe here
	}
	if count == 0 {
		return Str{runes: []rune{}}
	}
	return NewStr(strings.Repeat(string(s.runes), count))
}

// PadStart pads the current string with another string until the resulting string reaches the given length.
// JS equivalent: .padStart()
func (s Str) PadStart(targetLength int, padString ...string) Str {
	l := len(s.runes)
	if l >= targetLength {
		return s
	}

	pad := " "
	if len(padString) > 0 {
		pad = padString[0]
	}

	needed := targetLength - l
	padRunes := []rune(pad)
	if len(padRunes) == 0 {
		return s // Can't pad with empty string
	}

	var sb []rune
	for len(sb) < needed {
		sb = append(sb, padRunes...)
	}
	// Truncate to exact needed
	sb = sb[:needed]

	// Prepend
	return Str{runes: append(sb, s.runes...)}
}

// PadEnd pads the current string with another string until the resulting string reaches the given length.
// JS equivalent: .padEnd()
func (s Str) PadEnd(targetLength int, padString ...string) Str {
	l := len(s.runes)
	if l >= targetLength {
		return s
	}

	pad := " "
	if len(padString) > 0 {
		pad = padString[0]
	}

	needed := targetLength - l
	padRunes := []rune(pad)
	if len(padRunes) == 0 {
		return s
	}

	var sb []rune
	for len(sb) < needed {
		sb = append(sb, padRunes...)
	}
	// Truncate
	sb = sb[:needed]

	// Append
	return Str{runes: append(s.runes, sb...)}
}

// Wrap splits the string into multiple lines based on the given width.
// It considers whitespace for better line splitting (word wrapping).
func (s Str) Wrap(width int) []Str {
	if width <= 0 {
		return []Str{s}
	}

	// Safety limit for wrapping to prevent hang on massive inputs
	const maxRunes = 50000
	runesToWrap := s.runes
	if len(runesToWrap) > maxRunes {
		runesToWrap = runesToWrap[:maxRunes]
	}

	var result []Str
	var current []rune
	currentWidth := 0
	// ...

	// Helper to add line to results ensuring NO slice sharing
	flush := func() {
		if len(current) > 0 {
			line := make([]rune, len(current))
			copy(line, current)
			result = append(result, Str{runes: line})
			current = nil
			currentWidth = 0
		}
	}

	words := s.splitForWrap()

	for _, word := range words {
		wordRunes := word.runes
		wordWidth := 0
		for _, r := range wordRunes {
			rw := runewidth.RuneWidth(r)
			if rw == 0 {
				rw = 1
			}
			wordWidth += rw
		}

		// If single word is wider than the whole container, character-wrap it
		if wordWidth > width {
			flush()
			for _, r := range wordRunes {
				rw := runewidth.RuneWidth(r)
				if rw == 0 {
					rw = 1
				}
				if currentWidth+rw > width {
					flush()
				}
				current = append(current, r)
				currentWidth += rw
			}
			continue
		}

		if currentWidth+wordWidth > width && len(current) > 0 {
			flush()
			// Trim leading spaces when wrapping a word to a new line
			startIdx := 0
			for startIdx < len(wordRunes) && wordRunes[startIdx] == ' ' {
				startIdx++
			}
			wordRunes = wordRunes[startIdx:]
			wordWidth = 0
			for _, r := range wordRunes {
				rw := runewidth.RuneWidth(r)
				if rw == 0 {
					rw = 1
				}
				wordWidth += rw
			}
		}

		current = append(current, wordRunes...)
		currentWidth += wordWidth
	}

	flush()

	if len(result) == 0 && len(s.runes) == 0 {
		result = append(result, Str{runes: []rune{}})
	}

	return result
}

// splitForWrap divides the string into words and whitespace segments.
// It keeps consecutive spaces together as a single "word" segment to avoid
// unnecessary line breaks between multiple spaces.
func (s Str) splitForWrap() []Str {
	var result []Str
	if len(s.runes) == 0 {
		return result
	}

	var current []rune
	inSpace := s.runes[0] == ' '

	for _, r := range s.runes {
		isSpace := r == ' '
		if isSpace != inSpace {
			if len(current) > 0 {
				result = append(result, Str{runes: current})
			}
			current = []rune{r}
			inSpace = isSpace
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		result = append(result, Str{runes: current})
	}
	return result
}
