package utils

import "reflect"

func PtrIfNotZero[T any](v T) *T {
	var zero T
	if reflect.DeepEqual(v, zero) {
		return nil
	}
	return &v
}
