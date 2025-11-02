package filter

import (
	"example/models"
	"fmt"
	"strings"
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
	num, err := parseNumber(value)
	if err != nil {
		return false, 0, err
	}
	switch filter.Mode {
	case FilterModeEqual:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num == value, num, nil
	case FilterModeNotEqual:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num != value, num, nil
	case FilterModeContains:
		return false, num, fmt.Errorf("contains filter not supported for number field %s", filter.Field)
	case FilterModeNotContains:
		return false, num, fmt.Errorf("not contains filter not supported for number field %s", filter.Field)
	case FilterModeStartsWith:
		return false, num, fmt.Errorf("starts with filter not supported for number field %s", filter.Field)
	case FilterModeEndsWith:
		return false, num, fmt.Errorf("ends with filter not supported for number field %s", filter.Field)
	case FilterModeIsEmpty:
		return false, num, fmt.Errorf("is empty filter not supported for number field %s", filter.Field)
	case FilterModeIsNotEmpty:
		return false, num, fmt.Errorf("is not empty filter not supported for number field %s", filter.Field)
	case FilterModeGT:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num > value, num, nil
	case FilterModeGTE:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num >= value, num, nil
	case FilterModeLT:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num < value, num, nil
	case FilterModeLTE:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num <= value, num, nil
	case FilterModeRange:
		value, err := parseRangeNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num >= value.From && num <= value.To, num, nil
	case FilterModeBefore:
		return false, num, fmt.Errorf("before filter not supported for number field %s", filter.Field)
	case FilterModeAfter:
		return false, num, fmt.Errorf("after filter not supported for number field %s", filter.Field)
	default:
		return false, num, fmt.Errorf("unsupported filter mode: %s", filter.Mode)
	}
}

func (f *FilterHandler[T]) applyFilterText(value any, filter Filter) (bool, string, error) {
	data, err := parseText(value)
	if err != nil {
		return false, "", err
	}

	switch filter.Mode {
	case FilterModeEqual:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return data == substr, data, nil
	case FilterModeNotEqual:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return data != substr, data, nil
	case FilterModeContains:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return strings.Contains(data, substr), data, nil
	case FilterModeNotContains:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !strings.Contains(data, substr), data, nil
	case FilterModeStartsWith:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return strings.HasPrefix(data, substr), data, nil
	case FilterModeEndsWith:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return strings.HasSuffix(data, substr), data, nil
	case FilterModeIsEmpty:
		return data == "", data, nil
	case FilterModeIsNotEmpty:
		return data != "", data, nil
	case FilterModeGT:
		return false, data, fmt.Errorf("greater than filter not supported for text field %s", filter.Field)
	case FilterModeGTE:
		return false, data, fmt.Errorf("greater than or equal filter not supported for text field %s", filter.Field)
	case FilterModeLT:
		return false, data, fmt.Errorf("less than filter not supported for text field %s", filter.Field)
	case FilterModeLTE:
		return false, data, fmt.Errorf("less than or equal filter not supported for text field %s", filter.Field)
	case FilterModeRange:
		return false, data, fmt.Errorf("range filter not supported for text field %s", filter.Field)
	case FilterModeBefore:
		return false, data, fmt.Errorf("before filter not supported for text field %s", filter.Field)
	case FilterModeAfter:
		return false, data, fmt.Errorf("after filter not supported for text field %s", filter.Field)
	default:
		return false, data, fmt.Errorf("unsupported filter mode: %s", filter.Mode)
	}
}

func (f *FilterHandler[T]) applyFilterBool(value any, filter Filter) (bool, bool, error) {
	data, err := parseBool(value)
	if err != nil {
		return false, data, err
	}
	val, err := parseBool(filter.Value)
	if err != nil {
		return false, data, err
	}

	switch filter.Mode {
	case FilterModeEqual:
		return data == val, data, nil
	case FilterModeNotEqual:
		return data != val, data, nil
	case FilterModeContains:
		return false, data, fmt.Errorf("contains filter not supported for boolean field %s", filter.Field)
	case FilterModeNotContains:
		return false, data, fmt.Errorf("not contains filter not supported for boolean field %s", filter.Field)
	case FilterModeStartsWith:
		return false, data, fmt.Errorf("starts with filter not supported for boolean field %s", filter.Field)
	case FilterModeEndsWith:
		return false, data, fmt.Errorf("ends with filter not supported for boolean field %s", filter.Field)
	case FilterModeIsEmpty:
		return false, data, fmt.Errorf("is empty filter not supported for boolean field %s", filter.Field)
	case FilterModeIsNotEmpty:
		return false, data, fmt.Errorf("is not empty filter not supported for boolean field %s", filter.Field)
	case FilterModeGT:
		return false, data, fmt.Errorf("greater than filter not supported for boolean field %s", filter.Field)
	case FilterModeGTE:
		return false, data, fmt.Errorf("greater than or equal filter not supported for boolean field %s", filter.Field)
	case FilterModeLT:
		return false, data, fmt.Errorf("less than filter not supported for boolean field %s", filter.Field)
	case FilterModeLTE:
		return false, data, fmt.Errorf("less than or equal filter not supported for boolean field %s", filter.Field)
	case FilterModeRange:
		return false, data, fmt.Errorf("range filter not supported for boolean field %s", filter.Field)
	case FilterModeBefore:
		return false, data, fmt.Errorf("before filter not supported for boolean field %s", filter.Field)
	case FilterModeAfter:
		return false, data, fmt.Errorf("after filter not supported for boolean field %s", filter.Field)
	default:
		return false, data, fmt.Errorf("unsupported filter mode: %s", filter.Mode)
	}
}

func (f *FilterHandler[T]) applyFilterDate(value any, filter Filter) (bool, time.Time, error) {
	t, ok := value.(time.Time)
	if !ok {
		return false, time.Time{}, fmt.Errorf("invalid date type for field %s", filter.Field)
	}

	switch filter.Mode {
	case FilterModeEqual:
	case FilterModeNotEqual:
	case FilterModeContains:
	case FilterModeNotContains:
	case FilterModeStartsWith:
	case FilterModeEndsWith:
	case FilterModeIsEmpty:
	case FilterModeIsNotEmpty:
	case FilterModeGT:
	case FilterModeGTE:
	case FilterModeLT:
	case FilterModeLTE:
	case FilterModeRange:
	case FilterModeBefore:
	case FilterModeAfter:
	}
	return true, t, nil
}

func (f *FilterHandler[T]) applyFilterTime(value any, filter Filter) (bool, time.Time, error) {
	t, ok := value.(time.Time)
	if !ok {
		return false, time.Time{}, fmt.Errorf("invalid time type for field %s", filter.Field)
	}

	switch filter.Mode {
	case FilterModeEqual:
	case FilterModeNotEqual:
	case FilterModeContains:
	case FilterModeNotContains:
	case FilterModeStartsWith:
	case FilterModeEndsWith:
	case FilterModeIsEmpty:
	case FilterModeIsNotEmpty:
	case FilterModeGT:
	case FilterModeGTE:
	case FilterModeLT:
	case FilterModeLTE:
	case FilterModeRange:
	case FilterModeBefore:
	case FilterModeAfter:
	}
	return true, t, nil
}
