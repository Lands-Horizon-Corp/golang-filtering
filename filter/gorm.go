package filter

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type FilterHandler[T any] struct{}

// NewFilter creates a new filter handler
func NewFilter[T any]() *FilterHandler[T] {
	return &FilterHandler[T]{}
}

// FilterDataGorm applies filters, sorting, and pagination using GORM
func (f *FilterHandler[T]) FilterDataGorm(
	db *gorm.DB,
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

	// Build the query
	query := db.Model(new(T))

	// Apply filters
	if len(filterRoot.Filters) > 0 {
		query = f.applyFiltersGorm(query, filterRoot)
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
			if sortField.Order == FilterSortOrderDesc {
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

// applyFiltersGorm builds GORM query conditions from FilterRoot
func (f *FilterHandler[T]) applyFiltersGorm(db *gorm.DB, filterRoot FilterRoot) *gorm.DB {
	if len(filterRoot.Filters) == 0 {
		return db
	}

	if filterRoot.Logic == FilterLogicAnd {
		// AND logic - chain all conditions
		for _, filter := range filterRoot.Filters {
			db = f.applyFilterGorm(db, filter)
		}
	} else {
		// OR logic - use db.Where with OR
		var orConditions []string
		var orValues []any

		for _, filter := range filterRoot.Filters {
			condition, values := f.buildFilterCondition(filter)
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

// applyFilterGorm applies a single filter to the GORM query
func (f *FilterHandler[T]) applyFilterGorm(db *gorm.DB, filter Filter) *gorm.DB {
	condition, values := f.buildFilterCondition(filter)
	if condition != "" {
		db = db.Where(condition, values...)
	}
	return db
}

// buildFilterCondition builds SQL condition and values for a filter
func (f *FilterHandler[T]) buildFilterCondition(filter Filter) (string, []any) {
	field := filter.Field
	value := filter.Value

	switch filter.FilterDataType {
	case FilterDataTypeNumber:
		return f.buildNumberCondition(field, filter.Mode, value)
	case FilterDataTypeText:
		return f.buildTextCondition(field, filter.Mode, value)
	case FilterDataTypeBool:
		return f.buildBoolCondition(field, filter.Mode, value)
	case FilterDataTypeDate:
		return f.buildDateCondition(field, filter.Mode, value)
	case FilterDataTypeTime:
		return f.buildTimeCondition(field, filter.Mode, value)
	default:
		return "", nil
	}
}

// buildNumberCondition builds SQL condition for number filters
func (f *FilterHandler[T]) buildNumberCondition(field string, mode FilterMode, value any) (string, []any) {
	switch mode {
	case FilterModeEqual:
		return fmt.Sprintf("%s = ?", field), []any{value}
	case FilterModeNotEqual:
		return fmt.Sprintf("%s != ?", field), []any{value}
	case FilterModeGT:
		return fmt.Sprintf("%s > ?", field), []any{value}
	case FilterModeGTE:
		return fmt.Sprintf("%s >= ?", field), []any{value}
	case FilterModeLT:
		return fmt.Sprintf("%s < ?", field), []any{value}
	case FilterModeLTE:
		return fmt.Sprintf("%s <= ?", field), []any{value}
	case FilterModeRange:
		if rng, ok := value.(FilterRange); ok {
			return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{rng.From, rng.To}
		}
	}
	return "", nil
}

// buildTextCondition builds SQL condition for text filters
func (f *FilterHandler[T]) buildTextCondition(field string, mode FilterMode, value any) (string, []any) {
	str, ok := value.(string)
	if !ok {
		return "", nil
	}

	switch mode {
	case FilterModeEqual:
		return fmt.Sprintf("LOWER(%s) = LOWER(?)", field), []any{str}
	case FilterModeNotEqual:
		return fmt.Sprintf("LOWER(%s) != LOWER(?)", field), []any{str}
	case FilterModeContains:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), []any{"%" + str + "%"}
	case FilterModeNotContains:
		return fmt.Sprintf("LOWER(%s) NOT LIKE LOWER(?)", field), []any{"%" + str + "%"}
	case FilterModeStartsWith:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), []any{str + "%"}
	case FilterModeEndsWith:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field), []any{"%" + str}
	case FilterModeIsEmpty:
		return fmt.Sprintf("(%s IS NULL OR %s = '')", field, field), []any{}
	case FilterModeIsNotEmpty:
		return fmt.Sprintf("(%s IS NOT NULL AND %s != '')", field, field), []any{}
	}
	return "", nil
}

// buildBoolCondition builds SQL condition for boolean filters
func (f *FilterHandler[T]) buildBoolCondition(field string, mode FilterMode, value any) (string, []any) {
	switch mode {
	case FilterModeEqual:
		return fmt.Sprintf("%s = ?", field), []any{value}
	case FilterModeNotEqual:
		return fmt.Sprintf("%s != ?", field), []any{value}
	}
	return "", nil
}

// buildDateCondition builds SQL condition for date/datetime filters
func (f *FilterHandler[T]) buildDateCondition(field string, mode FilterMode, value any) (string, []any) {
	switch mode {
	case FilterModeEqual:
		if t, ok := value.(time.Time); ok {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{startOfDay, endOfDay}
		}
	case FilterModeNotEqual:
		if t, ok := value.(time.Time); ok {
			startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			return fmt.Sprintf("(%s < ? OR %s > ?)", field, field), []any{startOfDay, endOfDay}
		}
	case FilterModeGTE:
		return fmt.Sprintf("%s >= ?", field), []any{value}
	case FilterModeLT:
		return fmt.Sprintf("%s < ?", field), []any{value}
	case FilterModeLTE:
		return fmt.Sprintf("%s <= ?", field), []any{value}
	case FilterModeBefore:
		return fmt.Sprintf("%s < ?", field), []any{value}
	case FilterModeAfter:
		return fmt.Sprintf("%s > ?", field), []any{value}
	case FilterModeRange:
		if rng, ok := value.(FilterRange); ok {
			return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{rng.From, rng.To}
		}
	}
	return "", nil
}

// buildTimeCondition builds SQL condition for time filters
func (f *FilterHandler[T]) buildTimeCondition(field string, mode FilterMode, value any) (string, []any) {
	switch mode {
	case FilterModeEqual:
		return fmt.Sprintf("%s = ?", field), []any{value}
	case FilterModeNotEqual:
		return fmt.Sprintf("%s != ?", field), []any{value}
	case FilterModeGT:
		return fmt.Sprintf("%s > ?", field), []any{value}
	case FilterModeGTE, FilterModeAfter:
		return fmt.Sprintf("%s >= ?", field), []any{value}
	case FilterModeLT, FilterModeBefore:
		return fmt.Sprintf("%s < ?", field), []any{value}
	case FilterModeLTE:
		return fmt.Sprintf("%s <= ?", field), []any{value}
	case FilterModeRange:
		if rng, ok := value.(FilterRange); ok {
			return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{rng.From, rng.To}
		}
	}
	return "", nil
}
