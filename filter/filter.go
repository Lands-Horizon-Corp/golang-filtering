package filter

import (
	"example/models"
	"fmt"
)

type FilterHandler[T any] struct {
	getters map[string]func(*T) any
	key     string
}

func NewFilter[T any](key string) (*FilterHandler[T], error) {
	gettersInterface, ok := models.ModelFieldGetters[key]
	if !ok {
		return nil, fmt.Errorf("no field getters found for key: %s", key)
	}
	getters, ok := gettersInterface.(map[string]func(*T) any)
	if !ok {
		return nil, fmt.Errorf("invalid getter map type for key: %s", key)
	}
	return &FilterHandler[T]{
		getters: getters,
		key:     key,
	}, nil
}

func (f *FilterHandler[T]) FilterData(data []*T, filterRoot FilterRoot) (*PaginationResult[T], error) {
	result := PaginationResult[T]{
		Data:      data,
		PageIndex: 1,
		PageSize:  30,
	}
	if len(data) == 0 {
		return &result, nil
	}
	type filterGetter struct {
		filter Filter
		getter func(*T) any
	}
	validFilters := make([]filterGetter, 0, len(filterRoot.Filters))
	for _, filter := range filterRoot.Filters {
		if getter, exists := f.getters[filter.Field]; exists {
			validFilters = append(validFilters, filterGetter{filter: filter, getter: getter})
		}
	}
	for _, item := range data {
		fmt.Println("---")
		for _, fg := range validFilters {
			value := fg.getter(item)
			fmt.Printf("Field: %s, Value: %v, FilterMode: %s, FilterValue: %v\n",
				fg.filter.Field, value, fg.filter.Mode, fg.filter.Value)
		}
	}
	return &result, nil
}
