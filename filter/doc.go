// Package filter provides utilities for filtering, sorting, and paginating data sets.
package filter

type FilterHandler[T any] struct{}

func NewFilter[T any]() *FilterHandler[T] {
	return &FilterHandler[T]{}
}
