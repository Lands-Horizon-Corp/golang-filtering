package filter

import (
	"example/models"
	"fmt"
	"time"
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
	type sortGetter struct {
		getter func(*T) any
		order  FilterSortOrder
	}

	filteredData := []*T{}
	for _, item := range data {
		matches := filterRoot.Logic == FilterLogicAnd

		for _, fg := range validFilters {
			value := fg.getter(item)
			var match bool
			var err error

			switch fg.filter.FilterDataType {
			case FilterDataTypeNumber:
				match, _, err = f.applyFilterNumber(value, fg.filter)
			case FilterDataTypeText:
				match, _, err = f.applyFilterText(value, fg.filter)
			case FilterDataTypeDate:
				match, _, err = f.applyFilterDate(value, fg.filter)
			case FilterDataTypeBool:
				match, _, err = f.applyFilterBool(value, fg.filter)
			case FilterDataTypeTime:
				match, _, err = f.applyFilterTime(value, fg.filter)
			default:
				err = fmt.Errorf("unsupported data type: %s", fg.filter.FilterDataType)
			}

			if err != nil {
				return nil, err
			}

			if match != (filterRoot.Logic == FilterLogicAnd) {
				matches = match
				break
			}
		}

		if matches {
			filteredData = append(filteredData, item)
		}

	}
	result.Data = filteredData
	return &result, nil
}

func (f *FilterHandler[T]) applyFilterNumber(value any, filter Filter) (bool, float64, error) {
	var num float64

	switch v := value.(type) {
	case int:
		num = float64(v)
	case int32:
		num = float64(v)
	case int64:
		num = float64(v)
	case float32:
		num = float64(v)
	case float64:
		num = v
	default:
		return false, 0, fmt.Errorf("invalid number type for field %s", filter.Field)
	}
	return true, num, nil
}

func (f *FilterHandler[T]) applyFilterText(value any, filter Filter) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "", fmt.Errorf("invalid text type for field %s", filter.Field)
	}

	// TODO: Apply filter logic based on filter.Mode
	return true, str, nil
}

func (f *FilterHandler[T]) applyFilterDate(value any, filter Filter) (bool, time.Time, error) {
	t, ok := value.(time.Time)
	if !ok {
		return false, time.Time{}, fmt.Errorf("invalid date type for field %s", filter.Field)
	}

	// TODO: Apply filter logic based on filter.Mode
	return true, t, nil
}

func (f *FilterHandler[T]) applyFilterBool(value any, filter Filter) (bool, bool, error) {
	b, ok := value.(bool)
	if !ok {
		return false, false, fmt.Errorf("invalid boolean type for field %s", filter.Field)
	}

	// TODO: Apply filter logic based on filter.Mode
	return true, b, nil
}

func (f *FilterHandler[T]) applyFilterTime(value any, filter Filter) (bool, time.Time, error) {
	t, ok := value.(time.Time)
	if !ok {
		return false, time.Time{}, fmt.Errorf("invalid time type for field %s", filter.Field)
	}

	// TODO: Apply filter logic based on filter.Mode
	return true, t, nil
}
