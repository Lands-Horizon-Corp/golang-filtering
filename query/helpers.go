package query

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"gorm.io/gorm"
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

func autoJoinRelatedTables(db *gorm.DB, filters []FieldFilter, sortFields []SortField) *gorm.DB {
	joinedTables := make(map[string]bool)
	for _, filter := range filters {
		if strings.Contains(filter.Field, ".") {
			parts := strings.Split(filter.Field, ".")
			if len(parts) >= 2 {
				tableName := toPascalCase(parts[0])
				if !joinedTables[tableName] {
					db = db.Joins(tableName)
					joinedTables[tableName] = true
				}
			}
		}
	}
	for _, sortField := range sortFields {
		if strings.Contains(sortField.Field, ".") {
			parts := strings.Split(sortField.Field, ".")
			if len(parts) >= 2 {
				tableName := toPascalCase(parts[0])
				if !joinedTables[tableName] {
					db = db.Joins(tableName)
					joinedTables[tableName] = true
				}
			}
		}
	}
	return db
}

func parseNumber(value any) (float64, error) {
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
	if value == nil {
		return "", nil
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid text type for field %s", value)
	}
	return str, nil
}

func parseTime(value any) (time.Time, error) {
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
		if timeVal := reflect.ValueOf(value); timeVal.Kind() == reflect.Struct {
			if timeField := timeVal.FieldByName("Time"); timeField.IsValid() && timeField.Type() == reflect.TypeOf(time.Time{}) {
				t = timeField.Interface().(time.Time)
			} else {
				return time.Time{}, fmt.Errorf("invalid type for time: %T", value)
			}
		} else {
			return time.Time{}, fmt.Errorf("invalid type for time: %T", value)
		}
	}
	timeOnly := time.Date(0, time.January, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	return timeOnly, nil
}

func parseDateTime(value any) (time.Time, error) {
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
	if r, ok := value.(Range); ok {
		rng = r
	} else if m, ok := value.(map[string]any); ok {
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
	if r, ok := value.(Range); ok {
		rng = r
	} else if m, ok := value.(map[string]interface{}); ok {
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
	if r, ok := value.(Range); ok {
		rng = r
	} else if m, ok := value.(map[string]any); ok {
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
	if from.After(to) {
		return RangeDate{}, fmt.Errorf("range from time cannot be after to time")
	}

	return RangeDate{
		From: from,
		To:   to,
	}, nil
}

func parseBool(value any) (bool, error) {
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

func csvCreation[T any](data []*T, getter func(*T) map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)
	if len(data) == 0 {
		return buf.Bytes(), nil
	}
	firstRowFields := getter(data[0])
	fieldNames := make([]string, 0, len(firstRowFields))
	for k := range firstRowFields {
		fieldNames = append(fieldNames, k)
	}
	if err := csvWriter.Write(fieldNames); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}
	for _, item := range data {
		itemFields := getter(item)
		record := make([]string, len(fieldNames))
		for i, fieldName := range fieldNames {
			if value, exists := itemFields[fieldName]; exists {
				record[i] = fmt.Sprintf("%v", value)
			} else {
				record[i] = ""
			}
		}
		if err := csvWriter.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush CSV: %w", err)
	}
	return buf.Bytes(), nil
}

func toPascalCase(s string) string {
	if len(s) == 0 {
		return s
	}
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
