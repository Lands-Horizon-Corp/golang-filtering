package filter

import (
	"fmt"
	"reflect"
	"strings"
	"time"
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
		return 0, fmt.Errorf("invalid number type for field %s", value)
	}
	return num, nil
}

func parseText(value any) (string, error) {
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid text type for field %s", value)
	}
	return strings.ToLower(strings.TrimSpace(str)), nil
}

func parseTime(value any) (time.Time, error) {
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
		return time.Time{}, fmt.Errorf("invalid type for time: %T", value)
	}

	// Normalize to time-only in UTC
	timeOnly := time.Date(0, time.January, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	return timeOnly, nil
}

func parseDateTime(value any) (time.Time, error) {
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
	rng, ok := value.(FilterRange)
	if !ok {
		return RangeNumber{}, fmt.Errorf("invalid range type for field %s", value)
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
	rng, ok := value.(FilterRange)
	if !ok {
		return RangeDate{}, fmt.Errorf("invalid range type for field %s", value)
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
	rng, ok := value.(FilterRange)
	if !ok {
		return RangeDate{}, fmt.Errorf("invalid range type for field %s", value)
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
func generateGetters[T any]() map[string]func(*T) any {
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
		if field.Type.Kind() == reflect.Struct {
			generateNestedGetters(getters, field, fieldIndex, key)
		}
	}

	return getters
}

// generateNestedGetters generates getters for nested struct fields
func generateNestedGetters[T any](getters map[string]func(*T) any, parentField reflect.StructField, parentIndex int, parentKey string) {
	nestedType := parentField.Type

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
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			parentVal := val.Field(parentIndex)
			return parentVal.Field(nestedIndex).Interface()
		}

		getters[compositeKey] = nestedGetter
		if compositeKey != compositeLowerKey {
			getters[compositeLowerKey] = nestedGetter
		}
	}
}
