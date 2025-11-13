package filter

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	sanitizepkg "github.com/kennygrant/sanitize"
)

var dateTimeLayouts = []string{
	time.RFC3339,                     // "2006-01-02T15:04:05Z07:00"
	time.RFC3339Nano,                 // "2006-01-02T15:04:05.999999999Z07:00"
	time.RFC1123,                     // "Mon, 02 Jan 2006 15:04:05 MST"
	time.RFC1123Z,                    // "Mon, 02 Jan 2006 15:04:05 -0700"
	time.RFC822,                      // "02 Jan 06 15:04 MST"
	time.RFC822Z,                     // "02 Jan 06 15:04 -0700"
	time.RFC850,                      // "Monday, 02-Jan-06 15:04:05 MST"
	time.ANSIC,                       // "Mon Jan _2 15:04:05 2006"
	time.UnixDate,                    // "Mon Jan _2 15:04:05 MST 2006"
	time.RubyDate,                    // "Mon Jan 02 15:04:05 -0700 2006"
	"2006-01-02T15:04:05Z",           // ISO with Z
	"2006-01-02T15:04:05",            // ISO without zone
	"2006-01-02 15:04:05",            // Space separator
	"2006-01-02T15:04:05.999999999",  // With nanoseconds, no zone
	"01/02/2006 15:04:05",            // US MM/DD/YYYY
	"02/01/2006 15:04:05",            // EU DD/MM/YYYY
	"2006-01-02T15:04:05-07:00",      // With offset
	"Mon Jan 02 2006 15:04:05 -0700", // Variation with space and offset
	"2006/01/02 15:04:05",            // New: YYYY/MM/DD HH:MM:SS (addresses "2025/11/02 19:26:31")
	"2006/01/02T15:04:05",            // New: YYYY/MM/DDTHH:MM:SS
	"2006/01/02 15:04:05Z07:00",      // New: With offset
	"2006/01/02 15:04:05 MST",        // New: With named zone
	"2006-01-02",                     // New: Fallback for date-only as midnight
	"2006/01/02",                     // New: Slashed date-only
	"01/02/2006",                     // New: US date-only
	"02/01/2006",                     // New: EU date-only
}

var timeLayouts = []string{
	time.Kitchen,         // "3:04PM"
	"15:04:05",           // HH:MM:SS 24-hour
	"15:04",              // HH:MM
	"15:04:05.999999999", // With nanoseconds
	"3:04:05 PM",         // 12-hour with seconds
	"3:04 PM",            // 12-hour
	"15:04:05Z07:00",     // With offset
	"15:04:05 MST",       // With named zone
	"3:04:05 PM MST",     // 12-hour with named zone
	"15:04:05-07:00",     // New: Offset without Z
}

func parseNumber(value any) (float64, error) {
	// Handle nil values from nested pointers
	if value == nil {
		return 0, nil
	}
	var num float64
	switch v := value.(type) {
	case int:
		num = float64(v)
	case uint:
		num = float64(v)
	case int8:
		num = float64(v)
	case uint8:
		num = float64(v)
	case int16:
		num = float64(v)
	case uint16:
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
		return 0, fmt.Errorf("invalid number type for field %s", value)
	}
	return num, nil
}

func parseText(value any) (string, error) {
	// Handle nil values from nested pointers
	if value == nil {
		return "", nil
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid text type for field %s", value)
	}
	// Don't sanitize - GORM's parameterized queries handle SQL injection protection
	// Sanitizing converts spaces to hyphens which breaks text searches
	return str, nil
}

func parseTime(value any) (time.Time, error) {
	// Handle nil values from nested pointers
	if value == nil {
		return time.Time{}, nil
	}
	var t time.Time
	var err error

	switch v := value.(type) {
	case time.Time:
		t = v
	case string:
		var parsed bool
		for _, layout := range timeLayouts {
			t, err = time.Parse(layout, v)
			if err == nil {
				parsed = true
				break
			}
		}
		if !parsed {
			// Fallback to datetime layouts if time-only fails
			for _, layout := range dateTimeLayouts {
				t, err = time.Parse(layout, v)
				if err == nil {
					break
				}
			}
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid time format: %v", v)
			}
		}
	default:
		// Try to extract time.Time from embedded structs (e.g., custom time types)
		if timeVal := reflect.ValueOf(value); timeVal.Kind() == reflect.Struct {
			// Look for an embedded time.Time field
			if timeField := timeVal.FieldByName("Time"); timeField.IsValid() && timeField.Type() == reflect.TypeOf(time.Time{}) {
				t = timeField.Interface().(time.Time)
			} else {
				return time.Time{}, fmt.Errorf("invalid type for time: %T", value)
			}
		} else {
			return time.Time{}, fmt.Errorf("invalid type for time: %T", value)
		}
	}

	// Normalize to time-only in UTC
	timeOnly := time.Date(0, time.January, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	return timeOnly, nil
}

func parseDateTime(value any) (time.Time, error) {
	// Handle nil values from nested pointers
	if value == nil {
		return time.Time{}, nil
	}
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		for _, layout := range dateTimeLayouts {
			t, err := time.Parse(layout, v)
			if err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("invalid datetime format: %v", v)
	default:
		return time.Time{}, fmt.Errorf("invalid type for datetime: %T", value)
	}
}

func parseRangeNumber(value any) (RangeNumber, error) {
	var rng Range

	// Handle struct type (when used directly in Go code)
	if r, ok := value.(Range); ok {
		rng = r
	} else if m, ok := value.(map[string]interface{}); ok {
		// Handle map type (when parsed from JSON)
		fromVal, hasFrom := m["from"]
		toVal, hasTo := m["to"]
		if !hasFrom || !hasTo {
			return RangeNumber{}, fmt.Errorf("range must have both 'from' and 'to' fields")
		}
		rng = Range{From: fromVal, To: toVal}
	} else {
		return RangeNumber{}, fmt.Errorf("invalid range type for field %v (type: %T)", value, value)
	}
	from, err := parseNumber(rng.From)
	if err != nil {
		return RangeNumber{}, err
	}
	to, err := parseNumber(rng.To)
	if err != nil {
		return RangeNumber{}, err
	}
	return RangeNumber{
		From: from,
		To:   to,
	}, nil
}

func parseRangeDateTime(value any) (RangeDate, error) {
	var rng Range

	// Handle struct type (when used directly in Go code)
	if r, ok := value.(Range); ok {
		rng = r
	} else if m, ok := value.(map[string]interface{}); ok {
		// Handle map type (when parsed from JSON)
		fromVal, hasFrom := m["from"]
		toVal, hasTo := m["to"]
		if !hasFrom || !hasTo {
			return RangeDate{}, fmt.Errorf("range must have both 'from' and 'to' fields")
		}
		rng = Range{From: fromVal, To: toVal}
	} else {
		return RangeDate{}, fmt.Errorf("invalid range type for field %v (type: %T)", value, value)
	}
	from, err := parseDateTime(rng.From)
	if err != nil {
		return RangeDate{}, err
	}
	to, err := parseDateTime(rng.To)
	if err != nil {
		return RangeDate{}, err
	}
	if from.After(to) {
		return RangeDate{}, fmt.Errorf("range from date cannot be after to date")
	}
	return RangeDate{
		From: from,
		To:   to,
	}, nil
}

func parseRangeTime(value any) (RangeDate, error) {
	var rng Range

	// Handle struct type (when used directly in Go code)
	if r, ok := value.(Range); ok {
		rng = r
	} else if m, ok := value.(map[string]interface{}); ok {
		// Handle map type (when parsed from JSON)
		fromVal, hasFrom := m["from"]
		toVal, hasTo := m["to"]
		if !hasFrom || !hasTo {
			return RangeDate{}, fmt.Errorf("range must have both 'from' and 'to' fields")
		}
		rng = Range{From: fromVal, To: toVal}
	} else {
		return RangeDate{}, fmt.Errorf("invalid range type for field %v (type: %T)", value, value)
	}
	from, err := parseTime(rng.From)
	if err != nil {
		return RangeDate{}, err
	}
	to, err := parseTime(rng.To)
	if err != nil {
		return RangeDate{}, err
	}

	// Validate that from <= to
	if from.After(to) {
		return RangeDate{}, fmt.Errorf("range from time cannot be after to time")
	}

	return RangeDate{
		From: from,
		To:   to,
	}, nil
}

func parseBool(value any) (bool, error) {
	// Handle nil values from nested pointers
	if value == nil {
		return false, nil
	}
	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("invalid boolean type for field %s", value)
	}
	return b, nil
}

func hasTimeComponent(t time.Time) bool {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0 {
		return false
	}
	return true
}

func compareValues(a, b any) int {
	// Try to parse both values to standardized types
	numA, errA := parseNumber(a)
	numB, errB := parseNumber(b)
	if errA == nil && errB == nil {
		if numA < numB {
			return -1
		} else if numA > numB {
			return 1
		}
		return 0
	}

	strA, errA := parseText(a)
	strB, errB := parseText(b)
	if errA == nil && errB == nil {
		return strings.Compare(strA, strB)
	}

	boolA, errA := parseBool(a)
	boolB, errB := parseBool(b)
	if errA == nil && errB == nil {
		if boolA == boolB {
			return 0
		}
		if !boolA && boolB {
			return -1
		}
		return 1
	}

	// Try datetime comparison
	timeA, errA := parseDateTime(a)
	timeB, errB := parseDateTime(b)
	if errA == nil && errB == nil {
		if timeA.Before(timeB) {
			return -1
		} else if timeA.After(timeB) {
			return 1
		}
		return 0
	}

	// Try time-only comparison
	timeOnlyA, errA := parseTime(a)
	timeOnlyB, errB := parseTime(b)
	if errA == nil && errB == nil {
		if timeOnlyA.Before(timeOnlyB) {
			return -1
		} else if timeOnlyA.After(timeOnlyB) {
			return 1
		}
		return 0
	}

	// Fallback: cannot compare
	return 0
}

// generateGetters automatically generates field getters using reflection
func generateGetters[T any](maxDepth int) map[string]func(*T) any {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	getters := make(map[string]func(*T) any)
	if t.Kind() != reflect.Struct {
		return getters
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		fieldName := field.Name
		key := fieldName
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			tagValue := strings.Split(jsonTag, ",")[0]
			if tagValue != "" && tagValue != "-" {
				key = tagValue
			}
		}
		lowerKey := strings.ToLower(fieldName)
		fieldIndex := i
		getter := func(v *T) any {
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Pointer {
				val = val.Elem()
			}
			return val.Field(fieldIndex).Interface()
		}

		getters[key] = getter
		if key != lowerKey {
			getters[lowerKey] = getter
		}

		// Handle nested structs (both direct and pointer types)
		// Use configurable depth limit to avoid circular references
		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct && maxDepth > 1 {
			generateNestedGetters(getters, field, fieldIndex, key, field.Type.Kind() == reflect.Pointer, 1, maxDepth)
		}
	}

	return getters
}

// generateNestedGetters generates getters for nested struct fields with depth limit
func generateNestedGetters[T any](
	getters map[string]func(*T) any,
	parentField reflect.StructField,
	parentIndex int,
	parentKey string,
	isPointer bool,
	depth int,
	maxDepth int,
) {
	if depth > maxDepth {
		return // Stop at maximum depth to avoid circular references
	}

	nestedType := parentField.Type
	if isPointer {
		nestedType = nestedType.Elem()
	}

	for i := 0; i < nestedType.NumField(); i++ {
		nestedField := nestedType.Field(i)

		if !nestedField.IsExported() {
			continue
		}

		nestedFieldName := nestedField.Name
		nestedKey := nestedFieldName

		// Check for json tag on nested field
		if jsonTag := nestedField.Tag.Get("json"); jsonTag != "" {
			tagValue := strings.Split(jsonTag, ",")[0]
			if tagValue != "" && tagValue != "-" {
				nestedKey = tagValue
			}
		}

		// Create composite key: parent.nested
		compositeKey := parentKey + "." + nestedKey
		compositeLowerKey := parentKey + "." + strings.ToLower(nestedFieldName)

		// Create getter for nested field
		nestedIndex := i
		nestedGetter := func(v *T) any {
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Pointer {
				val = val.Elem()
			}
			parentVal := val.Field(parentIndex)

			// Handle pointer to struct
			if isPointer {
				if parentVal.IsNil() {
					return nil
				}
				parentVal = parentVal.Elem()
			}

			return parentVal.Field(nestedIndex).Interface()
		}

		getters[compositeKey] = nestedGetter
		if compositeKey != compositeLowerKey {
			getters[compositeLowerKey] = nestedGetter
		}

		// Recursively handle deeply nested structs with depth limit
		nestedFieldType := nestedField.Type
		isNestedPointer := false
		if nestedFieldType.Kind() == reflect.Pointer {
			nestedFieldType = nestedFieldType.Elem()
			isNestedPointer = true
		}
		if nestedFieldType.Kind() == reflect.Struct && depth < maxDepth {
			generateNestedGettersRecursive(getters, nestedField, parentIndex, nestedIndex, compositeKey, isPointer, isNestedPointer, depth+1, maxDepth)
		}
	}
}

// generateNestedGettersRecursive handles deeply nested struct fields with depth limit
func generateNestedGettersRecursive[T any](getters map[string]func(*T) any, parentField reflect.StructField, rootIndex, parentIndex int, parentKey string, rootIsPointer, parentIsPointer bool, depth int, maxDepth int) {
	if depth > maxDepth {
		return // Stop at maximum depth
	}

	nestedType := parentField.Type
	if parentIsPointer {
		nestedType = nestedType.Elem()
	}

	for i := 0; i < nestedType.NumField(); i++ {
		nestedField := nestedType.Field(i)

		if !nestedField.IsExported() {
			continue
		}

		nestedFieldName := nestedField.Name
		nestedKey := nestedFieldName

		if jsonTag := nestedField.Tag.Get("json"); jsonTag != "" {
			tagValue := strings.Split(jsonTag, ",")[0]
			if tagValue != "" && tagValue != "-" {
				nestedKey = tagValue
			}
		}

		compositeKey := parentKey + "." + nestedKey
		compositeLowerKey := parentKey + "." + strings.ToLower(nestedFieldName)

		nestedIndex := i
		nestedGetter := func(v *T) any {
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Pointer {
				val = val.Elem()
			}

			// Navigate to root parent
			rootVal := val.Field(rootIndex)
			if rootIsPointer {
				if rootVal.IsNil() {
					return nil
				}
				rootVal = rootVal.Elem()
			}

			// Navigate to immediate parent
			parentVal := rootVal.Field(parentIndex)
			if parentIsPointer {
				if parentVal.IsNil() {
					return nil
				}
				parentVal = parentVal.Elem()
			}

			return parentVal.Field(nestedIndex).Interface()
		}

		getters[compositeKey] = nestedGetter
		if compositeKey != compositeLowerKey {
			getters[compositeLowerKey] = nestedGetter
		}
	}
}

func Sanitize(input string) string {
	// Use kennygrant/sanitize package which handles:
	// - HTML/XSS sanitization
	// - SQL injection prevention
	// - Script tag removal
	// - Control character removal
	// All without manual pattern matching or regex

	// First, sanitize HTML and remove all potentially dangerous content
	sanitized := sanitizepkg.HTML(input)

	// Additionally sanitize as plain text to remove any remaining tags
	sanitized = sanitizepkg.Name(sanitized)

	// Note: GORM's parameterized queries provide the primary SQL injection protection.
	// This sanitization provides defense in depth for the application layer.

	return strings.TrimSpace(sanitized)
}

// fieldExists checks if a field (including nested fields) exists in the getters map
func (f *Handler[T]) fieldExists(field string) bool {
	if f.getters == nil {
		return false
	}

	// Check direct field access
	if _, exists := f.getters[field]; exists {
		return true
	}

	// Check lowercase version
	if _, exists := f.getters[strings.ToLower(field)]; exists {
		return true
	}

	return false
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

// escapeCSVField properly escapes a field value for CSV format
// This implementation follows RFC 4180 standard but replaces newlines with spaces for better compatibility
func escapeCSVField(field string) string {
	// Replace newlines with spaces for better single-line CSV readability
	field = strings.ReplaceAll(field, "\n", " ")
	field = strings.ReplaceAll(field, "\r", " ")
	// Clean up multiple spaces
	for strings.Contains(field, "  ") {
		field = strings.ReplaceAll(field, "  ", " ")
	}
	field = strings.TrimSpace(field)

	// Check if field contains special characters that require quoting
	needsQuoting := strings.Contains(field, ",") ||
		strings.Contains(field, "\"")

	if needsQuoting {
		// Escape existing quotes by doubling them (RFC 4180 standard)
		escaped := strings.ReplaceAll(field, "\"", "\"\"")
		return "\"" + escaped + "\""
	}
	return field
}

// escapeCSVFieldWithOptions provides additional control over CSV field escaping
func escapeCSVFieldWithOptions(field string, replaceNewlines bool) string {
	// Option to replace newlines with spaces for better single-line readability
	if replaceNewlines {
		field = strings.ReplaceAll(field, "\n", " ")
		field = strings.ReplaceAll(field, "\r", " ")
		// Clean up multiple spaces
		for strings.Contains(field, "  ") {
			field = strings.ReplaceAll(field, "  ", " ")
		}
		field = strings.TrimSpace(field)
	}

	// Check if field contains special characters that require quoting
	needsQuoting := strings.Contains(field, ",") ||
		strings.Contains(field, "\n") ||
		strings.Contains(field, "\r") ||
		strings.Contains(field, "\"")

	if needsQuoting {
		// Escape existing quotes by doubling them (RFC 4180 standard)
		escaped := strings.ReplaceAll(field, "\"", "\"\"")
		return "\"" + escaped + "\""
	}
	return field
}
