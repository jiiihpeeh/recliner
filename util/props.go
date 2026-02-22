package util

import (
	"reflect"
)

func GetProp[T any](props any, key string) (T, bool) {
	var zero T
	if props == nil {
		return zero, false
	}

	// Try common map types first
	switch m := props.(type) {
	case map[string]any:
		if v, ok := m[key]; ok {
			if val, ok := v.(T); ok {
				return val, true
			}
		}
		if p, ok := m["props"]; ok {
			if val, ok := GetProp[T](p, key); ok {
				return val, true
			}
		}
		return zero, false
	}

	// Try struct via reflection
	v := reflect.ValueOf(props)
	// ... (rest of reflection logic)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return zero, false
	}

	// Try exact match
	field := v.FieldByName(key)
	if !field.IsValid() {
		// Try capitalized match (e.g. "label" -> "Label")
		if len(key) > 0 {
			capKey := string(rune(key[0]-32)) + key[1:]
			if key[0] >= 'a' && key[0] <= 'z' {
				field = v.FieldByName(capKey)
			}
		}
	}

	if !field.IsValid() {
		// If it's a struct and has a "Props" field, try looking inside that
		propsField := v.FieldByName("Props")
		if propsField.IsValid() {
			if val, ok := GetProp[T](propsField.Interface(), key); ok {
				return val, true
			}
		}

		// If it's a struct and has a "Children" field and key is "children"
		if key == "children" {
			childrenField := v.FieldByName("Children")
			if childrenField.IsValid() {
				if val, ok := childrenField.Interface().(T); ok {
					return val, true
				}
			}
		}

		return zero, false
	}

	if val, ok := field.Interface().(T); ok {
		return val, true
	}

	return zero, false
}
