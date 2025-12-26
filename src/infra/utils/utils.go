package utils

import "reflect"

func PtrIfNotZero[T any](v T) *T {
	var zero T
	if reflect.DeepEqual(v, zero) {
		return nil
	}
	return &v
}

func ToMap[T comparable](s []T) map[T]struct{} {
	m := make(map[T]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
