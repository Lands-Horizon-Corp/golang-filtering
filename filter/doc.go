// Package filter provides utilities for filtering, sorting, and paginating data sets.
package filter

// FilterHandler is the main struct that handles filtering operations for a specific data type T.
type FilterHandler[T any] struct {
	getters map[string]func(*T) any
}

// NewFilter creates a new filter handler that automatically generates getters using reflection
func NewFilter[T any]() *FilterHandler[T] {
	getters := generateGetters[T]()
	return &FilterHandler[T]{
		getters: getters,
	}
}
