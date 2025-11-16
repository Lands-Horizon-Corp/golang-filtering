// Package filter provides utilities for filtering, sorting, and paginating data sets.
package filter

import (
	"fmt"
	"reflect"
	"sync"
)

// Handler is the main struct that handles filtering operations for a specific data type T.
type Handler[T any] struct {
	getters map[string]func(*T) any
}

type GolangFilteringConfig struct {
	MaxDepth *int
}

// Global cache for getters to prevent regenerating them for the same type
var (
	getterCacheMu sync.RWMutex
	getterCache   = make(map[string]interface{})
)

// NewFilter creates a new filter handler that automatically generates getters using reflection
// WARNING: Higher MaxDepth values (>3) can cause significant memory usage and performance issues
// with deeply nested or circular struct references. Default is 1 (no nested fields).
// Getters are cached globally to prevent regeneration for the same type.
func NewFilter[T any](config GolangFilteringConfig) *Handler[T] {
	depth := 1
	if config.MaxDepth != nil {
		depth = *config.MaxDepth
		// Safety limit: prevent excessive memory usage
		if depth > 5 {
			depth = 5 // Maximum safe depth
		}
		if depth < 0 {
			depth = 1
		}
	}

	// Generate cache key based on type and depth
	var zero T
	typeInfo := reflect.TypeOf(zero)
	cacheKey := fmt.Sprintf("%v:%d", typeInfo, depth)

	// Check cache first
	getterCacheMu.RLock()
	if cached, exists := getterCache[cacheKey]; exists {
		if cachedGetters, ok := cached.(map[string]func(*T) any); ok {
			getterCacheMu.RUnlock()
			return &Handler[T]{
				getters: cachedGetters,
			}
		}
	}
	getterCacheMu.RUnlock()

	// Generate getters if not cached
	getterCacheMu.Lock()
	defer getterCacheMu.Unlock()

	// Double-check after acquiring write lock
	if cached, exists := getterCache[cacheKey]; exists {
		if cachedGetters, ok := cached.(map[string]func(*T) any); ok {
			return &Handler[T]{
				getters: cachedGetters,
			}
		}
	}

	getters := generateGetters[T](depth)
	getterCache[cacheKey] = getters

	return &Handler[T]{
		getters: getters,
	}
}

// ExportGetters returns the getters map for inspection/debugging
func (h *Handler[T]) ExportGetters() map[string]func(*T) any {
	return h.getters
}
