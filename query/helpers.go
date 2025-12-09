package query

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

var dateTimeLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05.999999999",
	"01/02/2006 15:04:05",
	"02/01/2006 15:04:05",
	"2006-01-02T15:04:05-07:00",
	"Mon Jan 02 2006 15:04:05 -0700",
	"2006/01/02 15:04:05",
	"2006/01/02T15:04:05",
	"2006/01/02 15:04:05Z07:00",
	"2006/01/02 15:04:05 MST",
	"2006-01-02",
	"2006/01/02",
	"01/02/2006",
	"02/01/2006",
}

var timeLayouts = []string{
	time.Kitchen,
	"15:04:05",
	"15:04",
	"15:04:05.999999999",
	"3:04:05 PM",
	"3:04 PM",
	"15:04:05Z07:00",
	"15:04:05 MST",
	"3:04:05 PM MST",
	"15:04:05-07:00",
}

func autoJoinRelatedTables(db *gorm.DB, filters []FieldFilter, sortFields []SortField) *gorm.DB {
	joinedTables := make(map[string]bool)
	allowedJoinTables := map[string]struct{}{
		// Allowed PascalCase table names (example):
		"User":    {},
		"Profile": {},
		"Order":   {},
		// TODO: Add other table names from your schema as appropriate.
	}
	for _, filter := range filters {
		if strings.Contains(filter.Field, ".") {
			parts := strings.Split(filter.Field, ".")
			if len(parts) >= 2 {
				tableName := toPascalCase(parts[0])
				if _, allowed := allowedJoinTables[tableName]; allowed {
					if !joinedTables[tableName] {
						db = db.Joins(tableName)
						joinedTables[tableName] = true
					}
				} else {
					// Optionally: log or handle attempt to join disallowed table
					// fmt.Printf("Attempt to join disallowed table: %s\n", tableName)
					continue
				}
			}
		}
	}
	for _, sortField := range sortFields {
		if strings.Contains(sortField.Field, ".") {
			parts := strings.Split(sortField.Field, ".")
			if len(parts) >= 2 {
				tableName := toPascalCase(parts[0])
				if _, allowed := allowedJoinTables[tableName]; allowed {
					if !joinedTables[tableName] {
						db = db.Joins(tableName)
						joinedTables[tableName] = true
					}
				} else {
					// Optionally: log or handle attempt to join disallowed table
					// fmt.Printf("Attempt to join disallowed table: %s\n", tableName)
					continue
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
	switch v := value.(type) {
	case RangeNumber:
		return v, nil
	case Range:
		from, err := parseNumber(v.From)
		if err != nil {
			return RangeNumber{}, err
		}
		to, err := parseNumber(v.To)
		if err != nil {
			return RangeNumber{}, err
		}
		return RangeNumber{From: from, To: to}, nil
	case map[string]any:
		fromVal, hasFrom := v["from"]
		toVal, hasTo := v["to"]
		if !hasFrom || !hasTo {
			return RangeNumber{}, fmt.Errorf("range must have both 'from' and 'to' fields")
		}
		from, err := parseNumber(fromVal)
		if err != nil {
			return RangeNumber{}, err
		}
		to, err := parseNumber(toVal)
		if err != nil {
			return RangeNumber{}, err
		}
		return RangeNumber{From: from, To: to}, nil
	default:
		return RangeNumber{}, fmt.Errorf("invalid range type for field %v (type: %T)", value, value)
	}
}

func parseRangeDateTime(value any) (RangeDate, error) {
	switch v := value.(type) {
	case RangeDate:
		// Already a RangeDate, just return
		return v, nil
	case Range:
		from, err := parseDateTime(v.From)
		if err != nil {
			return RangeDate{}, err
		}
		to, err := parseDateTime(v.To)
		if err != nil {
			return RangeDate{}, err
		}
		if from.After(to) {
			return RangeDate{}, fmt.Errorf("range from date cannot be after to date")
		}
		return RangeDate{From: from, To: to}, nil
	case map[string]interface{}:
		fromVal, hasFrom := v["from"]
		toVal, hasTo := v["to"]
		if !hasFrom || !hasTo {
			return RangeDate{}, fmt.Errorf("range must have both 'from' and 'to' fields")
		}
		from, err := parseDateTime(fromVal)
		if err != nil {
			return RangeDate{}, err
		}
		to, err := parseDateTime(toVal)
		if err != nil {
			return RangeDate{}, err
		}
		if from.After(to) {
			return RangeDate{}, fmt.Errorf("range from date cannot be after to date")
		}
		return RangeDate{From: from, To: to}, nil
	default:
		return RangeDate{}, fmt.Errorf("invalid range type for field %v (type: %T)", value, value)
	}
}

func parseRangeTime(value any) (RangeDate, error) {
	switch v := value.(type) {
	case RangeDate:
		// Already a RangeDate, just return
		return v, nil
	case Range:
		from, err := parseTime(v.From)
		if err != nil {
			return RangeDate{}, err
		}
		to, err := parseTime(v.To)
		if err != nil {
			return RangeDate{}, err
		}
		if from.After(to) {
			return RangeDate{}, fmt.Errorf("range from time cannot be after to time")
		}
		return RangeDate{From: from, To: to}, nil
	case map[string]any:
		fromVal, hasFrom := v["from"]
		toVal, hasTo := v["to"]
		if !hasFrom || !hasTo {
			return RangeDate{}, fmt.Errorf("range must have both 'from' and 'to' fields")
		}
		from, err := parseTime(fromVal)
		if err != nil {
			return RangeDate{}, err
		}
		to, err := parseTime(toVal)
		if err != nil {
			return RangeDate{}, err
		}
		if from.After(to) {
			return RangeDate{}, fmt.Errorf("range from time cannot be after to time")
		}
		return RangeDate{From: from, To: to}, nil
	default:
		return RangeDate{}, fmt.Errorf("invalid range type for field %v (type: %T)", value, value)
	}
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

func ToModel[TData any, TResponse any](data *TData, resource func(*TData) *TResponse) *TResponse {
	if data == nil {
		return nil
	}
	return resource(data)
}

func ToModels[TData any, TResponse any](data []*TData, resource func(*TData) *TResponse) []*TResponse {
	if data == nil {
		return []*TResponse{}
	}
	out := make([]*TResponse, 0, len(data))
	for _, item := range data {
		if m := ToModel(item, resource); m != nil {
			out = append(out, m)
		}
	}
	return out
}

func Validate[T any](ctx echo.Context, v *validator.Validate) (*T, error) {
	var req T
	if err := ctx.Bind(&req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}
	if err := v.Struct(req); err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "validation failed: "+err.Error())
	}
	return &req, nil
}

func parseFilters(ctx echo.Context) (StructuredFilter, error) {
	filterParam := ctx.QueryParam("filter")
	if filterParam == "" {
		return StructuredFilter{Logic: LogicAnd}, nil
	}
	filterDecodedRaw, err := url.QueryUnescape(filterParam)
	if err != nil {
		return StructuredFilter{}, fmt.Errorf("unescaping failed: %w", err)
	}
	filterBytes, err := base64.StdEncoding.DecodeString(filterDecodedRaw)
	if err != nil {
		return StructuredFilter{}, fmt.Errorf("base64 decoding failed: %w", err)
	}
	var filterRoot StructuredFilter
	if err := json.Unmarshal(filterBytes, &filterRoot); err != nil {
		return StructuredFilter{}, fmt.Errorf("JSON unmarshalling failed: %w", err)
	}
	if filterRoot.Logic == "" {
		filterRoot.Logic = LogicAnd
	}
	return filterRoot, nil
}

func parseSort(ctx echo.Context) ([]SortField, error) {
	sortParam := ctx.QueryParam("sort")
	if sortParam == "" {
		return nil, nil
	}
	sortDecodedRaw, err := url.QueryUnescape(sortParam)
	if err != nil {
		return nil, fmt.Errorf("unescaping failed: %w", err)
	}
	sortBytes, err := base64.StdEncoding.DecodeString(sortDecodedRaw)
	if err != nil {
		return nil, fmt.Errorf("base64 decoding failed: %w", err)
	}
	var sortFields []SortField
	if err := json.Unmarshal(sortBytes, &sortFields); err != nil {
		return nil, fmt.Errorf("JSON unmarshalling failed: %w", err)
	}
	for i, sf := range sortFields {
		order := strings.ToLower(strings.TrimSpace(string(sf.Order)))
		if order != "asc" && order != "desc" {
			sortFields[i].Order = "asc"
		} else {
			sortFields[i].Order = SortOrder(order)
		}
	}
	return sortFields, nil
}

func parsePageSize(ctx echo.Context) (int, error) {
	pageSize, err := strconv.Atoi(ctx.QueryParam("pageSize"))
	if err != nil {
		return 0, fmt.Errorf("invalid pageSize parameter: %w", err)
	}
	return pageSize, nil
}

func parsePageIndex(ctx echo.Context) (int, error) {
	pageIndex, err := strconv.Atoi(ctx.QueryParam("pageIndex"))
	if err != nil {
		return 0, fmt.Errorf("invalid pageIndex parameter: %w", err)
	}
	return pageIndex, nil
}

func parseQuery(ctx echo.Context) (StructuredFilter, int, int, error) {
	filterRoot, err := parseFilters(ctx)
	if err != nil {
		return StructuredFilter{}, 0, 0, fmt.Errorf("filter processing failed: %w", err)
	}
	sortFields, err := parseSort(ctx)
	if err != nil {
		return StructuredFilter{}, 0, 0, fmt.Errorf("sort processing failed: %w", err)
	}
	filterRoot.SortFields = sortFields
	pageIndex, err := parsePageIndex(ctx)
	if err != nil {
		return StructuredFilter{}, 0, 0, fmt.Errorf("pageIndex processing failed: %w", err)
	}
	pageSize, err := parsePageSize(ctx)
	if err != nil {
		return StructuredFilter{}, 0, 0, fmt.Errorf("pageSize processing failed: %w", err)
	}

	return filterRoot, pageIndex, pageSize, nil
}

func strParseFilters(value string) (StructuredFilter, error) {
	if value == "" {
		return StructuredFilter{Logic: LogicAnd}, nil
	}
	decodedRaw, err := url.QueryUnescape(value)
	if err != nil {
		return StructuredFilter{}, fmt.Errorf("unescaping failed: %w", err)
	}
	filterBytes, err := base64.StdEncoding.DecodeString(decodedRaw)
	if err != nil {
		return StructuredFilter{}, fmt.Errorf("base64 decoding failed: %w", err)
	}
	var filterRoot StructuredFilter
	if err := json.Unmarshal(filterBytes, &filterRoot); err != nil {
		return StructuredFilter{}, fmt.Errorf("JSON unmarshalling failed: %w", err)
	}
	if filterRoot.Logic == "" {
		filterRoot.Logic = LogicAnd
	}
	return filterRoot, nil
}

func strParseSort(value string) ([]SortField, error) {
	if value == "" {
		return nil, nil
	}
	decodedRaw, err := url.QueryUnescape(value)
	if err != nil {
		return nil, fmt.Errorf("unescaping failed: %w", err)
	}
	sortBytes, err := base64.StdEncoding.DecodeString(decodedRaw)
	if err != nil {
		return nil, fmt.Errorf("base64 decoding failed: %w", err)
	}
	var sortFields []SortField
	if err := json.Unmarshal(sortBytes, &sortFields); err != nil {
		return nil, fmt.Errorf("JSON unmarshalling failed: %w", err)
	}
	for i, sf := range sortFields {
		order := strings.ToLower(strings.TrimSpace(string(sf.Order)))
		if order != "asc" && order != "desc" {
			sortFields[i].Order = "asc"
		} else {
			sortFields[i].Order = SortOrder(order)
		}
	}
	return sortFields, nil
}

func strParsePageSize(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	pageSize, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid pageSize parameter: %w", err)
	}
	return pageSize, nil
}

func strParsePageIndex(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	pageIndex, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid pageIndex parameter: %w", err)
	}
	return pageIndex, nil
}

func strParseQuery(value string) (StructuredFilter, int, int, error) {
	parts := strings.Split(value, "|")
	var filterStr, sortStr, pageIndexStr, pageSizeStr string
	if len(parts) > 0 {
		filterStr = parts[0]
	}
	if len(parts) > 1 {
		sortStr = parts[1]
	}
	if len(parts) > 2 {
		pageIndexStr = parts[2]
	}
	if len(parts) > 3 {
		pageSizeStr = parts[3]
	}

	filterRoot, err := strParseFilters(filterStr)
	if err != nil {
		return StructuredFilter{}, 0, 0, fmt.Errorf("filter processing failed: %w", err)
	}
	sortFields, err := strParseSort(sortStr)
	if err != nil {
		return StructuredFilter{}, 0, 0, fmt.Errorf("sort processing failed: %w", err)
	}
	filterRoot.SortFields = sortFields
	pageIndex, err := strParsePageIndex(pageIndexStr)
	if err != nil {
		return StructuredFilter{}, 0, 0, fmt.Errorf("pageIndex processing failed: %w", err)
	}
	pageSize, err := strParsePageSize(pageSizeStr)
	if err != nil {
		return StructuredFilter{}, 0, 0, fmt.Errorf("pageSize processing failed: %w", err)
	}
	return filterRoot, pageIndex, pageSize, nil
}
