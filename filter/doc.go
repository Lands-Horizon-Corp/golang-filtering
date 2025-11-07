// Package filter provides utilities for filtering, sorting, and paginating data sets.
package filter

// Handler is the main struct that handles filtering operations for a specific data type T.
type Handler[T any] struct {
	getters map[string]func(*T) any
}

type GolangFilteringConfig struct {
	MaxDepth *int
}

// New creates a new filter handler that automatically generates getters using reflection
func NewFilter[T any](config GolangFilteringConfig) *Handler[T] {
	depth := 1
	if config.MaxDepth != nil {
		depth = *config.MaxDepth
	}
	getters := generateGetters[T](depth)
	return &Handler[T]{
		getters: getters,
	}
}
