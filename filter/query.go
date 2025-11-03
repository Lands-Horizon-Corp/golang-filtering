package filter

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func (f *FilterHandler[T]) FilterDataQuery(
	data []*T,
	filterRoot FilterRoot,
	pageIndex int,
	pageSize int,
) (*PaginationResult[T], error) {
	result := PaginationResult[T]{
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}

	// Set defaults if not provided
	if result.PageIndex <= 0 {
		result.PageIndex = 1
	}
	if result.PageSize <= 0 {
		result.PageSize = 30
	}

	if len(data) == 0 {
		result.Data = data // Reuse the empty slice
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

	numCPU := runtime.NumCPU()
	chunkSize := (len(data) + numCPU - 1) / numCPU

	// Pre-allocate result slices with exact capacity to avoid reallocations
	resultChunks := make([][]*T, numCPU)
	for i := range numCPU {
		resultChunks[i] = make([]*T, 0, chunkSize)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var filterErr error

	// Atomic counter for progress tracking
	var processedCount int64

	for i := range numCPU {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			start := workerID * chunkSize
			end := min(start+chunkSize, len(data))
			if start >= len(data) {
				return
			}

			localFiltered := resultChunks[workerID] // Reuse pre-allocated slice

			for _, item := range data[start:end] {
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
						mu.Lock()
						if filterErr == nil {
							filterErr = err
						}
						mu.Unlock()
						return
					}
					if match != (filterRoot.Logic == FilterLogicAnd) {
						matches = match
						break
					}
				}
				if matches {
					localFiltered = append(localFiltered, item) // Only append pointers, no data cloning
				}
				atomic.AddInt64(&processedCount, 1)
			}
			resultChunks[workerID] = localFiltered
		}(i)
	}

	wg.Wait()

	if filterErr != nil {
		return nil, filterErr
	}

	// Calculate total size first
	totalSize := 0
	for _, chunk := range resultChunks {
		totalSize += len(chunk)
	}

	// Pre-allocate exactly the size needed - no reallocation
	filteredData := make([]*T, 0, totalSize)
	for _, chunk := range resultChunks {
		filteredData = append(filteredData, chunk...) // Only copying pointers, not data
	}

	// Sort after filtering
	if len(filterRoot.SortFields) > 0 {
		sort.Slice(filteredData, func(i, j int) bool {
			return f.compareItems(filteredData[i], filteredData[j], filterRoot.SortFields) < 0
		})
	}

	// Apply pagination
	result.TotalSize = len(filteredData)
	result.TotalPage = (result.TotalSize + result.PageSize - 1) / result.PageSize

	// Calculate start and end indices for the requested page
	startIdx := (result.PageIndex - 1) * result.PageSize
	endIdx := startIdx + result.PageSize

	// Handle out of bounds
	if startIdx >= len(filteredData) {
		result.Data = make([]*T, 0) // Empty slice with zero allocation
		return &result, nil
	}

	if endIdx > len(filteredData) {
		endIdx = len(filteredData)
	}

	// Return only the requested page - this is a slice view, not a copy
	// No data cloning, just sharing pointers to the same underlying data
	result.Data = filteredData[startIdx:endIdx]
	return &result, nil
}

func (f *FilterHandler[T]) compareItems(a, b *T, sortFields []SortField) int {
	for _, sortField := range sortFields {
		getter, exists := f.getters[sortField.Field]
		if !exists {
			continue
		}
		valA := getter(a)
		valB := getter(b)
		cmp := compareValues(valA, valB)
		if sortField.Order == FilterSortOrderDesc {
			cmp = -cmp
		}

		if cmp != 0 {
			return cmp
		}
	}
	return 0
}

// applyFilterNumber applies a number filter and returns whether the value matches the filter
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

// applyFilterText applies a text filter and returns whether the value matches the filter
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

// applyFilterBool applies a boolean filter and returns whether the value matches the filter
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

// applyFilterDate applies a date filter and returns whether the value matches the filter
func (f *FilterHandler[T]) applyFilterDate(value any, filter Filter) (bool, time.Time, error) {
	data, err := parseDateTime(value)
	if err != nil {
		return false, time.Time{}, err
	}
	hasTime := hasTimeComponent(data)

	switch filter.Mode {
	case FilterModeEqual:
		filterVal, err := parseDateTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		if hasTime {
			return data.Equal(filterVal), data, nil
		} else {
			startOfDay := time.Date(data.Year(), data.Month(), data.Day(), 0, 0, 0, 0, data.Location())
			endOfDay := time.Date(data.Year(), data.Month(), data.Day(), 23, 59, 59, 999999999, data.Location())
			return !filterVal.Before(startOfDay) && !filterVal.After(endOfDay), data, nil
		}
	case FilterModeNotEqual:
		filterVal, err := parseDateTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		if hasTime {
			return !data.Equal(filterVal), data, nil
		} else {
			startOfDay := time.Date(data.Year(), data.Month(), data.Day(), 0, 0, 0, 0, data.Location())
			endOfDay := time.Date(data.Year(), data.Month(), data.Day(), 23, 59, 59, 999999999, data.Location())
			return filterVal.Before(startOfDay) || filterVal.After(endOfDay), data, nil
		}
	case FilterModeContains:
		return false, data, fmt.Errorf("contains filter not supported for date field %s", filter.Field)
	case FilterModeNotContains:
		return false, data, fmt.Errorf("not contains filter not supported for date field %s", filter.Field)
	case FilterModeStartsWith:
		return false, data, fmt.Errorf("starts with filter not supported for date field %s", filter.Field)
	case FilterModeEndsWith:
		return false, data, fmt.Errorf("ends with filter not supported for date field %s", filter.Field)
	case FilterModeIsEmpty:
		return false, data, fmt.Errorf("is empty filter not supported for date field %s", filter.Field)
	case FilterModeIsNotEmpty:
		return false, data, fmt.Errorf("is not empty filter not supported for date field %s", filter.Field)
	case FilterModeGT:
		return false, data, fmt.Errorf("greater than filter not supported for date field %s", filter.Field)
	case FilterModeGTE:
		filterVal, err := parseDateTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		if hasTime {
			return data.Equal(filterVal) || data.After(filterVal), data, nil
		} else {
			startOfDay := time.Date(filterVal.Year(), filterVal.Month(), filterVal.Day(), 0, 0, 0, 0, filterVal.Location())
			return data.Equal(startOfDay) || data.After(startOfDay), data, nil
		}
	case FilterModeLT:
		filterVal, err := parseDateTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		if hasTime {
			return data.Before(filterVal), data, nil
		} else {
			startOfDay := time.Date(filterVal.Year(), filterVal.Month(), filterVal.Day(), 0, 0, 0, 0, filterVal.Location())
			return data.Before(startOfDay), data, nil
		}
	case FilterModeLTE:
		filterVal, err := parseDateTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		if hasTime {
			return data.Equal(filterVal) || data.Before(filterVal), data, nil
		} else {
			endOfDay := time.Date(filterVal.Year(), filterVal.Month(), filterVal.Day(), 23, 59, 59, 999999999, filterVal.Location())
			return data.Equal(endOfDay) || data.Before(endOfDay), data, nil
		}
	case FilterModeRange:
		rangeVal, err := parseRangeDateTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		if hasTime {
			return !data.Before(rangeVal.From) && !data.After(rangeVal.To), data, nil
		} else {
			startOfDay := time.Date(rangeVal.From.Year(), rangeVal.From.Month(), rangeVal.From.Day(), 0, 0, 0, 0, rangeVal.From.Location())
			endOfDay := time.Date(rangeVal.To.Year(), rangeVal.To.Month(), rangeVal.To.Day(), 23, 59, 59, 999999999, rangeVal.To.Location())
			return !data.Before(startOfDay) && !data.After(endOfDay), data, nil
		}
	case FilterModeBefore:
		filterVal, err := parseDateTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		if hasTime {
			return data.Before(filterVal), data, nil
		} else {
			startOfDay := time.Date(filterVal.Year(), filterVal.Month(), filterVal.Day(), 0, 0, 0, 0, filterVal.Location())
			return data.Before(startOfDay), data, nil
		}
	case FilterModeAfter:
		filterVal, err := parseDateTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		if hasTime {
			return data.After(filterVal), data, nil
		} else {
			// After the end of the day
			endOfDay := time.Date(filterVal.Year(), filterVal.Month(), filterVal.Day(), 23, 59, 59, 999999999, filterVal.Location())
			return data.After(endOfDay), data, nil
		}
	default:
		return false, data, fmt.Errorf("unsupported filter mode: %s", filter.Mode)
	}
}

// applyFilterTime applies a time filter and returns whether the value matches the filter
func (f *FilterHandler[T]) applyFilterTime(value any, filter Filter) (bool, time.Time, error) {
	data, err := parseTime(value)
	if err != nil {
		return false, time.Time{}, err
	}
	switch filter.Mode {
	case FilterModeEqual:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return data.Equal(filterVal), data, nil

	case FilterModeNotEqual:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !data.Equal(filterVal), data, nil

	case FilterModeGTE, FilterModeAfter:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !data.Before(filterVal), data, nil

	case FilterModeLTE:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !data.After(filterVal), data, nil

	case FilterModeLT, FilterModeBefore:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return data.Before(filterVal), data, nil

	case FilterModeGT:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return data.After(filterVal), data, nil

	case FilterModeRange:
		rangeVal, err := parseRangeTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !data.Before(rangeVal.From) && !data.After(rangeVal.To), data, nil

	case FilterModeContains, FilterModeNotContains, FilterModeStartsWith, FilterModeEndsWith,
		FilterModeIsEmpty, FilterModeIsNotEmpty:
		return false, data, fmt.Errorf("filter mode %s not supported for time field %s", filter.Mode, filter.Field)

	default:
		return false, data, fmt.Errorf("unsupported filter mode: %s", filter.Mode)
	}
}
