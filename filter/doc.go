// Package filter provides utilities for filtering, sorting, and paginating data sets.
package filter

// Handler is the main struct that handles filtering operations for a specific data type T.
type Handler[T any] struct {
	getters map[string]func(*T) any
}

// New creates a new filter handler that automatically generates getters using reflection
func NewFilter[T any]() *Handler[T] {
	getters := generateGetters[T]()
	return &Handler[T]{
		getters: getters,
	}
}
