package util

func Ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func MinF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func GetPropString(props any, key string) string {
	v, _ := GetProp[string](props, key)
	return v
}
