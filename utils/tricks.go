package utils

func IfThen[T any](cond bool, trueVal T, falseVal T) T {
	if cond {
		return trueVal
	} else {
		return falseVal
	}
}
