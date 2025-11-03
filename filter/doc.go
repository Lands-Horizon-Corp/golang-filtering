// Package filter provides utilities for filtering, sorting, and paginating data sets.
package filter

type FilterHandler[T any] struct {
	getters map[string]func(*T) any
	key     string
}

func NewFilter[T any]() *FilterHandler[T] {
	return &FilterHandler[T]{}
}
