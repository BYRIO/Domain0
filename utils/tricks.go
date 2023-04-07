package utils

// IfThen is a generic function to replace the if-else statement
// DO NOT USE *Pointer as the value of trueVal or falseVal
func IfThen[T any](cond bool, trueVal T, falseVal T) T {
	if cond {
		return trueVal
	} else {
		return falseVal
	}
}

// IfThenPtr is a generic function to return ptr value if ptr is not nil, otherwise return default value
func IfThenPtr[T any](ptr *T, defaultval T) T {
	if ptr != nil {
		return *ptr
	} else {
		return defaultval
	}
}
