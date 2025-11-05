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

// DataQuery performs in-memory filtering with parallel processing.
// It filters the provided data slice based on the filter configuration and returns paginated results.
func (f *Handler[T]) DataQuery(
	data []*T,
	filterRoot Root,
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
		filter FieldFilter
		getter func(*T) any
	}
	valids := make([]filterGetter, 0, len(filterRoot.FieldFilters))
	for _, filter := range filterRoot.FieldFilters {
		if getter, exists := f.getters[filter.Field]; exists {
			valids = append(valids, filterGetter{filter: filter, getter: getter})
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

			localed := resultChunks[workerID] // Reuse pre-allocated slice

			for _, item := range data[start:end] {
				matches := filterRoot.Logic == LogicAnd
				for _, fg := range valids {
					value := fg.getter(item)
					var match bool
					var err error
					switch fg.filter.DataType {
					case DataTypeNumber:
						match, _, err = f.applyNumber(value, fg.filter)
					case DataTypeText:
						match, _, err = f.applyText(value, fg.filter)
					case DataTypeDate:
						match, _, err = f.applyDate(value, fg.filter)
					case DataTypeBool:
						match, _, err = f.applyBool(value, fg.filter)
					case DataTypeTime:
						match, _, err = f.applyTime(value, fg.filter)
					default:
						err = fmt.Errorf("unsupported data type: %s", fg.filter.DataType)
					}
					if err != nil {
						mu.Lock()
						if filterErr == nil {
							filterErr = err
						}
						mu.Unlock()
						return
					}
					if match != (filterRoot.Logic == LogicAnd) {
						matches = match
						break
					}
				}
				if matches {
					localed = append(localed, item) // Only append pointers, no data cloning
				}
				atomic.AddInt64(&processedCount, 1)
			}
			resultChunks[workerID] = localed
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

func (f *Handler[T]) compareItems(a, b *T, sortFields []SortField) int {
	for _, sortField := range sortFields {
		getter, exists := f.getters[sortField.Field]
		if !exists {
			continue
		}
		valA := getter(a)
		valB := getter(b)
		cmp := compareValues(valA, valB)
		if sortField.Order == SortOrderDesc {
			cmp = -cmp
		}

		if cmp != 0 {
			return cmp
		}
	}
	return 0
}

// applyNumber applies a number filter and returns whether the value matches the filter
func (f *Handler[T]) applyNumber(value any, filter FieldFilter) (bool, float64, error) {
	num, err := parseNumber(value)
	if err != nil {
		return false, 0, err
	}
	switch filter.Mode {
	case ModeEqual:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num == value, num, nil
	case ModeNotEqual:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num != value, num, nil
	case ModeContains:
		return false, num, fmt.Errorf("contains filter not supported for number field %s", filter.Field)
	case ModeNotContains:
		return false, num, fmt.Errorf("not contains filter not supported for number field %s", filter.Field)
	case ModeStartsWith:
		return false, num, fmt.Errorf("starts with filter not supported for number field %s", filter.Field)
	case ModeEndsWith:
		return false, num, fmt.Errorf("ends with filter not supported for number field %s", filter.Field)
	case ModeIsEmpty:
		return false, num, fmt.Errorf("is empty filter not supported for number field %s", filter.Field)
	case ModeIsNotEmpty:
		return false, num, fmt.Errorf("is not empty filter not supported for number field %s", filter.Field)
	case ModeGT:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num > value, num, nil
	case ModeGTE:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num >= value, num, nil
	case ModeLT:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num < value, num, nil
	case ModeLTE:
		value, err := parseNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num <= value, num, nil
	case ModeRange:
		value, err := parseRangeNumber(filter.Value)
		if err != nil {
			return false, num, err
		}
		return num >= value.From && num <= value.To, num, nil
	case ModeBefore:
		return false, num, fmt.Errorf("before filter not supported for number field %s", filter.Field)
	case ModeAfter:
		return false, num, fmt.Errorf("after filter not supported for number field %s", filter.Field)
	default:
		return false, num, fmt.Errorf("unsupported filter mode: %s", filter.Mode)
	}
}

// applyText applies a text filter and returns whether the value matches the filter
// All text comparisons are case-insensitive
func (f *Handler[T]) applyText(value any, filter FieldFilter) (bool, string, error) {
	data, err := parseText(value)
	if err != nil {
		return false, "", err
	}

	// Convert to lowercase for case-insensitive comparison
	dataLower := strings.ToLower(data)

	switch filter.Mode {
	case ModeEqual:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return dataLower == strings.ToLower(substr), data, nil
	case ModeNotEqual:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return dataLower != strings.ToLower(substr), data, nil
	case ModeContains:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return strings.Contains(dataLower, strings.ToLower(substr)), data, nil
	case ModeNotContains:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !strings.Contains(dataLower, strings.ToLower(substr)), data, nil
	case ModeStartsWith:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return strings.HasPrefix(dataLower, strings.ToLower(substr)), data, nil
	case ModeEndsWith:
		substr, err := parseText(filter.Value)
		if err != nil {
			return false, data, err
		}
		return strings.HasSuffix(dataLower, strings.ToLower(substr)), data, nil
	case ModeIsEmpty:
		return data == "", data, nil
	case ModeIsNotEmpty:
		return data != "", data, nil
	case ModeGT:
		return false, data, fmt.Errorf("greater than filter not supported for text field %s", filter.Field)
	case ModeGTE:
		return false, data, fmt.Errorf("greater than or equal filter not supported for text field %s", filter.Field)
	case ModeLT:
		return false, data, fmt.Errorf("less than filter not supported for text field %s", filter.Field)
	case ModeLTE:
		return false, data, fmt.Errorf("less than or equal filter not supported for text field %s", filter.Field)
	case ModeRange:
		return false, data, fmt.Errorf("range filter not supported for text field %s", filter.Field)
	case ModeBefore:
		return false, data, fmt.Errorf("before filter not supported for text field %s", filter.Field)
	case ModeAfter:
		return false, data, fmt.Errorf("after filter not supported for text field %s", filter.Field)
	default:
		return false, data, fmt.Errorf("unsupported filter mode: %s", filter.Mode)
	}
}

// applyBool applies a boolean filter and returns whether the value matches the filter
func (f *Handler[T]) applyBool(value any, filter FieldFilter) (bool, bool, error) {
	data, err := parseBool(value)
	if err != nil {
		return false, data, err
	}
	val, err := parseBool(filter.Value)
	if err != nil {
		return false, data, err
	}

	switch filter.Mode {
	case ModeEqual:
		return data == val, data, nil
	case ModeNotEqual:
		return data != val, data, nil
	case ModeContains:
		return false, data, fmt.Errorf("contains filter not supported for boolean field %s", filter.Field)
	case ModeNotContains:
		return false, data, fmt.Errorf("not contains filter not supported for boolean field %s", filter.Field)
	case ModeStartsWith:
		return false, data, fmt.Errorf("starts with filter not supported for boolean field %s", filter.Field)
	case ModeEndsWith:
		return false, data, fmt.Errorf("ends with filter not supported for boolean field %s", filter.Field)
	case ModeIsEmpty:
		return false, data, fmt.Errorf("is empty filter not supported for boolean field %s", filter.Field)
	case ModeIsNotEmpty:
		return false, data, fmt.Errorf("is not empty filter not supported for boolean field %s", filter.Field)
	case ModeGT:
		return false, data, fmt.Errorf("greater than filter not supported for boolean field %s", filter.Field)
	case ModeGTE:
		return false, data, fmt.Errorf("greater than or equal filter not supported for boolean field %s", filter.Field)
	case ModeLT:
		return false, data, fmt.Errorf("less than filter not supported for boolean field %s", filter.Field)
	case ModeLTE:
		return false, data, fmt.Errorf("less than or equal filter not supported for boolean field %s", filter.Field)
	case ModeRange:
		return false, data, fmt.Errorf("range filter not supported for boolean field %s", filter.Field)
	case ModeBefore:
		return false, data, fmt.Errorf("before filter not supported for boolean field %s", filter.Field)
	case ModeAfter:
		return false, data, fmt.Errorf("after filter not supported for boolean field %s", filter.Field)
	default:
		return false, data, fmt.Errorf("unsupported filter mode: %s", filter.Mode)
	}
}

// applyDate applies a date filter and returns whether the value matches the filter
func (f *Handler[T]) applyDate(value any, filter FieldFilter) (bool, time.Time, error) {
	data, err := parseDateTime(value)
	if err != nil {
		return false, time.Time{}, err
	}
	hasTime := hasTimeComponent(data)

	switch filter.Mode {
	case ModeEqual:
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
	case ModeNotEqual:
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
	case ModeContains:
		return false, data, fmt.Errorf("contains filter not supported for date field %s", filter.Field)
	case ModeNotContains:
		return false, data, fmt.Errorf("not contains filter not supported for date field %s", filter.Field)
	case ModeStartsWith:
		return false, data, fmt.Errorf("starts with filter not supported for date field %s", filter.Field)
	case ModeEndsWith:
		return false, data, fmt.Errorf("ends with filter not supported for date field %s", filter.Field)
	case ModeIsEmpty:
		return false, data, fmt.Errorf("is empty filter not supported for date field %s", filter.Field)
	case ModeIsNotEmpty:
		return false, data, fmt.Errorf("is not empty filter not supported for date field %s", filter.Field)
	case ModeGT:
		return false, data, fmt.Errorf("greater than filter not supported for date field %s", filter.Field)
	case ModeGTE:
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
	case ModeLT:
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
	case ModeLTE:
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
	case ModeRange:
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
	case ModeBefore:
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
	case ModeAfter:
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

// applyTime applies a time filter and returns whether the value matches the filter
func (f *Handler[T]) applyTime(value any, filter FieldFilter) (bool, time.Time, error) {
	data, err := parseTime(value)
	if err != nil {
		return false, time.Time{}, err
	}
	switch filter.Mode {
	case ModeEqual:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return data.Equal(filterVal), data, nil

	case ModeNotEqual:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !data.Equal(filterVal), data, nil

	case ModeGTE, ModeAfter:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !data.Before(filterVal), data, nil

	case ModeLTE:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !data.After(filterVal), data, nil

	case ModeLT, ModeBefore:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return data.Before(filterVal), data, nil

	case ModeGT:
		filterVal, err := parseTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return data.After(filterVal), data, nil

	case ModeRange:
		rangeVal, err := parseRangeTime(filter.Value)
		if err != nil {
			return false, data, err
		}
		return !data.Before(rangeVal.From) && !data.After(rangeVal.To), data, nil

	case ModeContains, ModeNotContains, ModeStartsWith, ModeEndsWith,
		ModeIsEmpty, ModeIsNotEmpty:
		return false, data, fmt.Errorf("filter mode %s not supported for time field %s", filter.Mode, filter.Field)

	default:
		return false, data, fmt.Errorf("unsupported filter mode: %s", filter.Mode)
	}
}
