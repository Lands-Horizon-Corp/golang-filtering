package filter

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// DataGorm performs database-level filtering using GORM queries.
// It generates SQL WHERE clauses based on the filter configuration and returns paginated results.
func (f *Handler[T]) DataGorm(
	db *gorm.DB,
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

	// Build the query
	query := db.Model(new(T))

	// Apply filters
	if len(filterRoot.FieldFilters) > 0 {
		query = f.applysGorm(query, filterRoot)
	}

	// Get total count before pagination
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}
	result.TotalSize = int(totalCount)
	result.TotalPage = (result.TotalSize + result.PageSize - 1) / result.PageSize

	// Apply sorting
	if len(filterRoot.SortFields) > 0 {
		for _, sortField := range filterRoot.SortFields {
			order := "ASC"
			if sortField.Order == SortOrderDesc {
				order = "DESC"
			}
			query = query.Order(fmt.Sprintf("%s %s", sortField.Field, order))
		}
	}

	// Apply pagination
	offset := (result.PageIndex - 1) * result.PageSize
	query = query.Offset(offset).Limit(result.PageSize)

	// Execute query
	var data []*T
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}

	result.Data = data
	return &result, nil
}

func (f *Handler[T]) applysGorm(db *gorm.DB, filterRoot Root) *gorm.DB {
	if len(filterRoot.FieldFilters) == 0 {
		return db
	}

	if filterRoot.Logic == LogicAnd {
		for _, filter := range filterRoot.FieldFilters {
			db = f.applyGorm(db, filter)
		}
	} else {
		var orConditions []string
		var orValues []any

		for _, filter := range filterRoot.FieldFilters {
			condition, values := f.buildCondition(filter)
			if condition != "" {
				orConditions = append(orConditions, condition)
				orValues = append(orValues, values...)
			}
		}
		if len(orConditions) > 0 {
			db = db.Where(strings.Join(orConditions, " OR "), orValues...)
		}
	}
	return db
}

// applyGorm applies a single filter to the GORM query
func (f *Handler[T]) applyGorm(db *gorm.DB, filter FieldFilter) *gorm.DB {
	condition, values := f.buildCondition(filter)
	if condition != "" {
		db = db.Where(condition, values...)
	}
	return db
}

// buildCondition builds SQL condition and values for a filter
func (f *Handler[T]) buildCondition(filter FieldFilter) (string, []any) {
	field := filter.Field
	value := filter.Value

	switch filter.DataType {
	case DataTypeNumber:
		return f.buildNumberCondition(field, filter.Mode, value)
	case DataTypeText:
		return f.buildTextCondition(field, filter.Mode, value)
	case DataTypeBool:
		return f.buildBoolCondition(field, filter.Mode, value)
	case DataTypeDate:
		return f.buildDateCondition(field, filter.Mode, value)
	case DataTypeTime:
		return f.buildTimeCondition(field, filter.Mode, value)
	default:
		return "", nil
	}
}

// buildNumberCondition builds SQL condition for number filters
func (f *Handler[T]) buildNumberCondition(field string, mode Mode, value any) (string, []any) {
	switch mode {
	case ModeEqual:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s = ?", field), []any{num}
	case ModeNotEqual:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s != ?", field), []any{num}
	case ModeGT:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s > ?", field), []any{num}
	case ModeGTE:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s >= ?", field), []any{num}
	case ModeLT:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s < ?", field), []any{num}
	case ModeLTE:
		num, err := parseNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s <= ?", field), []any{num}
	case ModeRange:
		rangeVal, err := parseRangeNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{rangeVal.From, rangeVal.To}
	}
	return "", nil
}

// buildTextCondition builds SQL condition for text filters
func (f *Handler[T]) buildTextCondition(field string, mode Mode, value any) (string, []any) {
	str, err := parseText(value)
	if err != nil {
		return "", nil
	}

	switch mode {
	case ModeEqual:
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", field), []any{str}
	case ModeNotEqual:
		return fmt.Sprintf("LOWER(%s) != LOWER(?)", field), []any{str}
	case ModeContains:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), []any{"%" + str + "%"}
	case ModeNotContains:
		return fmt.Sprintf("LOWER(%s) NOT LIKE LOWER(?)", field), []any{"%" + str + "%"}
	case ModeStartsWith:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), []any{str + "%"}
	case ModeEndsWith:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), []any{"%" + str}
	case ModeIsEmpty:
		return fmt.Sprintf("(%s IS NULL OR %s = '')", field, field), []any{}
	case ModeIsNotEmpty:
		return fmt.Sprintf("(%s IS NOT NULL AND %s != '')", field, field), []any{}
	}
	return "", nil
}

// buildBoolCondition builds SQL condition for boolean filters
func (f *Handler[T]) buildBoolCondition(field string, mode Mode, value any) (string, []any) {
	boolVal, err := parseBool(value)
	if err != nil {
		return "", nil
	}
	switch mode {
	case ModeEqual:
		return fmt.Sprintf("%s = ?", field), []any{boolVal}
	case ModeNotEqual:
		return fmt.Sprintf("%s != ?", field), []any{boolVal}
	}
	return "", nil
}

// buildDateCondition builds SQL condition for date/datetime filters
func (f *Handler[T]) buildDateCondition(field string, mode Mode, value any) (string, []any) {
	switch mode {
	case ModeEqual:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s = ?", field), []any{t}
		}
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{startOfDay, endOfDay}
	case ModeNotEqual:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s != ?", field), []any{t}
		}
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
		return fmt.Sprintf("(%s < ? OR %s > ?)", field, field), []any{startOfDay, endOfDay}
	case ModeGTE:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s >= ?", field), []any{t}
		} else {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			return fmt.Sprintf("%s >= ?", field), []any{startOfDay}
		}
	case ModeLT:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s < ?", field), []any{t}
		} else {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			return fmt.Sprintf("%s < ?", field), []any{startOfDay}
		}
	case ModeLTE:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s <= ?", field), []any{t}
		} else {
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			return fmt.Sprintf("%s <= ?", field), []any{endOfDay}
		}
	case ModeBefore:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s < ?", field), []any{t}
		} else {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			return fmt.Sprintf("%s < ?", field), []any{startOfDay}
		}
	case ModeAfter:
		t, err := parseDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTime := hasTimeComponent(t)
		if hasTime {
			return fmt.Sprintf("%s > ?", field), []any{t}
		} else {
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			return fmt.Sprintf("%s > ?", field), []any{endOfDay}
		}
	case ModeRange:
		rangeVal, err := parseRangeDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTimeFrom := hasTimeComponent(rangeVal.From)
		hasTimeTo := hasTimeComponent(rangeVal.To)

		if hasTimeFrom && hasTimeTo {
			return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{rangeVal.From, rangeVal.To}
		} else {
			startOfDay := time.Date(rangeVal.From.Year(), rangeVal.From.Month(), rangeVal.From.Day(), 0, 0, 0, 0, rangeVal.From.Location())
			endOfDay := time.Date(rangeVal.To.Year(), rangeVal.To.Month(), rangeVal.To.Day(), 23, 59, 59, 999999999, rangeVal.To.Location())
			return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{startOfDay, endOfDay}
		}
	}
	return "", nil
}

// buildTimeCondition builds SQL condition for time filters
func (f *Handler[T]) buildTimeCondition(field string, mode Mode, value any) (string, []any) {
	switch mode {
	case ModeEqual:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s = ?", field), []any{t}
	case ModeNotEqual:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s != ?", field), []any{t}
	case ModeGT:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s > ?", field), []any{t}
	case ModeGTE, ModeAfter:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s >= ?", field), []any{t}
	case ModeLT, ModeBefore:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s < ?", field), []any{t}
	case ModeLTE:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s <= ?", field), []any{t}
	case ModeRange:
		rangeVal, err := parseRangeTime(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{rangeVal.From, rangeVal.To}
	}
	return "", nil
}
