package filter

import (
	"bytes"
	"encoding/csv"
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

	// Set defaults if not provided - use 0-based indexing
	if result.PageIndex < 0 {
		result.PageIndex = 0
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
				// If no filters are provided, include all items
				if len(valids) == 0 {
					localed = append(localed, item)
				} else {
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

	// Calculate start and end indices for the requested page (0-based indexing)
	startIdx := result.PageIndex * result.PageSize
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

// DataQueryNoPage performs in-memory filtering with parallel processing without pagination.
// It filters the provided data slice based on the filter configuration and returns all matching results as a simple array.
func (f *Handler[T]) DataQueryNoPage(
	data []*T,
	filterRoot Root,
) ([]*T, error) {
	if len(data) == 0 {
		return data, nil // Return the empty slice directly
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
				// If no filters are provided, include all items
				if len(valids) == 0 {
					localed = append(localed, item)
				} else {
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

	return filteredData, nil
}

// DataQueryNoPageCSV performs in-memory filtering with parallel processing and returns results as CSV bytes.
// It filters the provided data slice based on the filter configuration and exports all matching results as CSV format.
// Field names are automatically used as CSV headers.
func (f *Handler[T]) DataQueryNoPageCSV(
	data []*T,
	filterRoot Root,
) ([]byte, error) {
	// Use DataQueryNoPage to get filtered results
	filteredData, err := f.DataQueryNoPage(data, filterRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to filter data: %w", err)
	}

	// Build CSV content using encoding/csv
	var csvBuffer strings.Builder
	csvWriter := csv.NewWriter(&csvBuffer)

	// Sort field names for deterministic column ordering
	fieldNames := make([]string, 0, len(f.getters))
	for fieldName := range f.getters {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames)

	// Write headers
	if err := csvWriter.Write(fieldNames); err != nil {
		return nil, fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	for _, item := range filteredData {
		record := make([]string, len(fieldNames))
		for i, fieldName := range fieldNames {
			// Get the value using the getter for this field
			getter := f.getters[fieldName]
			value := getter(item)
			record[i] = fmt.Sprintf("%v", value)
		}

		if err := csvWriter.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	// Flush the writer to ensure all data is written
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return []byte(csvBuffer.String()), nil
}

// DataQueryNoPageCSVCustom performs in-memory filtering with parallel processing and returns results as CSV bytes.
// It uses a custom callback function to allow users to define exactly what fields and values to include in the CSV output.
// This provides full control over CSV structure and field mapping on the user side.
//
// Parameters:
//   - data: slice of pointers to the data type T to filter
//   - filterRoot: filter configuration defining conditions, logic, and sorting
//   - customGetter: callback function that takes a data item and returns a map[string]any
//     where keys are column headers and values are the corresponding data
//
// Returns CSV bytes with headers from the customGetter map keys, sorted alphabetically for deterministic ordering.
//
// Example usage:
//
//	csvData, err := handler.DataQueryNoPageCSVCustom(users, filterRoot, func(user *User) map[string]any {
//	    return map[string]any{
//	        "Full Name": user.FirstName + " " + user.LastName,
//	        "Email": user.Email,
//	        "Status": user.IsActive,
//	        "Department": user.Department.Name, // Access nested fields
//	    }
//	})
func (f *Handler[T]) DataQueryNoPageCSVCustom(
	data []*T,
	filterRoot Root,
	customGetter func(*T) map[string]any,
) ([]byte, error) {
	// Use DataQueryNoPage to get filtered results
	filteredData, err := f.DataQueryNoPage(data, filterRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to filter data: %w", err)
	}

	if len(filteredData) == 0 {
		// If no data, we can't determine headers, return empty CSV with no headers
		return []byte(""), nil
	}

	// Get headers from the first item using the custom getter
	firstItemFields := customGetter(filteredData[0])

	// Sort field names for deterministic column ordering
	fieldNames := make([]string, 0, len(firstItemFields))
	for fieldName := range firstItemFields {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames)

	// Build CSV content using encoding/csv
	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)

	// Write headers
	if err := csvWriter.Write(fieldNames); err != nil {
		return nil, fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	for _, item := range filteredData {
		itemFields := customGetter(item)
		record := make([]string, len(fieldNames))

		for i, fieldName := range fieldNames {
			// Get the value for this field from the custom getter result
			if value, exists := itemFields[fieldName]; exists {
				record[i] = fmt.Sprintf("%v", value)
			} else {
				// If field doesn't exist in this item's result, use empty string
				record[i] = ""
			}
		}

		if err := csvWriter.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
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

		// Check if filter range values have time components
		hasTimeFrom := hasTimeComponent(rangeVal.From)
		hasTimeTo := hasTimeComponent(rangeVal.To)

		if hasTimeFrom && hasTimeTo {
			// Both range boundaries have time - do exact timestamp comparison
			return !data.Before(rangeVal.From) && !data.After(rangeVal.To), data, nil
		} else {
			// Date-only range - compare against full day boundaries
			startOfFromDay := time.Date(rangeVal.From.Year(), rangeVal.From.Month(), rangeVal.From.Day(), 0, 0, 0, 0, rangeVal.From.Location())
			endOfToDay := time.Date(rangeVal.To.Year(), rangeVal.To.Month(), rangeVal.To.Day(), 23, 59, 59, 999999999, rangeVal.To.Location())
			return !data.Before(startOfFromDay) && !data.After(endOfToDay), data, nil
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
