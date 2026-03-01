package util

import (
	"reflect"
	"strings"
)

func GetProp[T any](props any, key string) (T, bool) {
	var zero T
	if props == nil {
		return zero, false
	}

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

	v := reflect.ValueOf(props)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return zero, false
	}

	field := v.FieldByName(key)
	if !field.IsValid() {
		if len(key) > 0 {
			capKey := string(rune(key[0]-32)) + key[1:]
			if key[0] >= 'a' && key[0] <= 'z' {
				field = v.FieldByName(capKey)
			}
		}
	}

	if !field.IsValid() {
		if key == "id" {
			field = v.FieldByName("ID")
		}
	}

	if !field.IsValid() {
		typ := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := typ.Field(i)
			if strings.EqualFold(f.Name, key) {
				field = v.Field(i)
				break
			}
		}
	}

	if !field.IsValid() {
		propsField := v.FieldByName("Props")
		if propsField.IsValid() {
			if val, ok := GetProp[T](propsField.Interface(), key); ok {
				return val, true
			}
		}

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
