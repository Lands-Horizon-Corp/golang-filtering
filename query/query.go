package query

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (f *Pagination[T]) arrQuery(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
) *gorm.DB {
	db = f.applyJoinsForFilters(db, filters)
	db = f.applySQLFilters(db, filters)
	if len(sorts) > 0 {
		db = f.applySort(db, sorts)
	} else {
		db = db.Order("created_at DESC")
	}
	return db
}

func (f *Pagination[T]) structuredQuery(
	db *gorm.DB,
	filterRoot StructuredFilter,
) *gorm.DB {
	db = autoJoinRelatedTables(db, filterRoot.FieldFilters, filterRoot.SortFields)
	if len(filterRoot.Preload) > 0 {
		for _, preloadField := range filterRoot.Preload {
			db = db.Preload(preloadField)
		}
	}
	if len(filterRoot.FieldFilters) > 0 {
		db = f.applysGorm(db, filterRoot)
	}
	hasNestedFields := false
	for _, filter := range filterRoot.FieldFilters {
		if strings.Contains(filter.Field, ".") {
			hasNestedFields = true
			break
		}
	}
	if !hasNestedFields {
		for _, sortField := range filterRoot.SortFields {
			if strings.Contains(sortField.Field, ".") {
				hasNestedFields = true
				break
			}
		}
	}
	var mainTableName string
	if hasNestedFields {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(new(T)); err == nil {
			mainTableName = stmt.Schema.Table
		}
	}
	if len(filterRoot.SortFields) > 0 {
		for _, sortField := range filterRoot.SortFields {
			if !strings.Contains(sortField.Field, ".") && !f.fieldExists(db, sortField.Field) {
				continue
			}
			order := "ASC"
			if sortField.Order == SortOrderDesc {
				order = "DESC"
			}
			field := sortField.Field
			if strings.Contains(field, ".") {
				parts := strings.Split(field, ".")
				if len(parts) >= 2 {
					parts[0] = toPascalCase(parts[0])
					field = fmt.Sprintf(`"%s"."%s"`, parts[0], parts[1])
					for i := 2; i < len(parts); i++ {
						field = fmt.Sprintf(`%s."%s"`, field, parts[i])
					}
				}
			} else if mainTableName != "" {
				field = fmt.Sprintf(`"%s"."%s"`, mainTableName, field)
			}
			db = db.Order(fmt.Sprintf("%s %s", field, order))
		}
	} else {
		if mainTableName != "" {
			db = db.Order(fmt.Sprintf(`"%s"."created_at" DESC`, mainTableName))
		} else {
			db = db.Order("created_at DESC")
		}
	}
	return db
}

func (f *Pagination[T]) applysGorm(db *gorm.DB, filterRoot StructuredFilter) *gorm.DB {
	if len(filterRoot.FieldFilters) == 0 {
		return db
	}

	hasNestedFields := false
	for _, filter := range filterRoot.FieldFilters {

		if strings.Contains(filter.Field, ".") {
			hasNestedFields = true
			break
		}

	}

	var mainTableName string
	if hasNestedFields {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(new(T)); err == nil {
			mainTableName = stmt.Schema.Table
		}
	}
	if filterRoot.Logic == LogicAnd {
		for _, filter := range filterRoot.FieldFilters {

			if strings.Contains(filter.Field, ".") || f.fieldExists(db, filter.Field) {

				condition, values := f.buildConditionWithTableName(filter, mainTableName)
				if condition != "" {
					db = db.Where(condition, values...)
				}
			}
		}
	} else {
		var orConditions []string
		var orValues []any
		for _, filter := range filterRoot.FieldFilters {
			if strings.Contains(filter.Field, ".") || f.fieldExists(db, filter.Field) {
				condition, values := f.buildConditionWithTableName(filter, mainTableName)
				if condition != "" {
					orConditions = append(orConditions, condition)
					orValues = append(orValues, values...)
				}
			}
		}
		if len(orConditions) > 0 {
			db = db.Where(strings.Join(orConditions, " OR "), orValues...)
		}
	}
	return db
}

func (f *Pagination[T]) buildConditionWithTableName(filter FieldFilter, mainTableName string) (string, []any) {
	field := filter.Field
	value := filter.Value
	isNestedField := strings.Contains(field, ".")
	if isNestedField {
		parts := strings.Split(field, ".")
		if len(parts) >= 2 {
			parts[0] = toPascalCase(parts[0])
			field = fmt.Sprintf(`"%s"."%s"`, parts[0], parts[1])
			for i := 2; i < len(parts); i++ {
				field = fmt.Sprintf(`%s."%s"`, field, parts[i])
			}
		}
	} else if mainTableName != "" {
		field = fmt.Sprintf(`"%s"."%s"`, mainTableName, field)
	}

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

func (f *Pagination[T]) buildNumberCondition(field string, mode Mode, value any) (string, []any) {
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
	case ModeRange, ModeInside:
		rangeVal, err := parseRangeNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{rangeVal.From, rangeVal.To}
	case ModeOutside:
		rangeVal, err := parseRangeNumber(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s < ? OR %s > ?", field, field), []any{rangeVal.From, rangeVal.To}
	}
	return "", nil
}

func (f *Pagination[T]) buildTextCondition(field string, mode Mode, value any) (string, []any) {
	if mode == ModeRange {

		rangeVal, ok := value.(Range)
		if !ok {
			return "", nil
		}
		fromStr, err := parseText(rangeVal.From)
		if err != nil {
			return "", nil
		}
		toStr, err := parseText(rangeVal.To)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s BETWEEN ? AND ?", field), []any{fromStr, toStr}
	}
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
	case ModeGT:
		return fmt.Sprintf("%s > ?", field), []any{str}
	case ModeGTE, ModeAfter:
		return fmt.Sprintf("%s >= ?", field), []any{str}
	case ModeLT, ModeBefore:
		return fmt.Sprintf("%s < ?", field), []any{str}
	case ModeLTE:
		return fmt.Sprintf("%s <= ?", field), []any{str}
	}
	return "", nil
}

func (f *Pagination[T]) buildBoolCondition(field string, mode Mode, value any) (string, []any) {
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

func (f *Pagination[T]) buildDateCondition(field string, mode Mode, value any) (string, []any) {
	switch mode {
	case ModeInside:
		rangeVal, err := parseRangeDateTime(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s >= ? AND %s <= ?", field, field), []any{rangeVal.From, rangeVal.To}
	case ModeOutside:
		rangeVal, err := parseRangeDateTime(value)
		if err != nil {
			return "", nil
		}
		return fmt.Sprintf("%s < ? OR %s > ?", field, field), []any{rangeVal.From, rangeVal.To}
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
			return fmt.Sprintf("%s >= ?", field), []any{t}
		} else {
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			return fmt.Sprintf("%s >= ?", field), []any{endOfDay}
		}
	case ModeRange:
		rangeVal, err := parseRangeDateTime(value)
		if err != nil {
			return "", nil
		}
		hasTimeFrom := hasTimeComponent(rangeVal.From)
		hasTimeTo := hasTimeComponent(rangeVal.To)

		if hasTimeFrom && hasTimeTo {
			// Both dates have time components, use exact timestamps
			return fmt.Sprintf("%s >= ? AND %s <= ?", field, field), []any{rangeVal.From, rangeVal.To}
		} else {
			// Date-only range: include entire days from start of From day to end of To day
			startOfFromDay := time.Date(rangeVal.From.Year(), rangeVal.From.Month(), rangeVal.From.Day(), 0, 0, 0, 0, rangeVal.From.Location())
			endOfToDay := time.Date(rangeVal.To.Year(), rangeVal.To.Month(), rangeVal.To.Day(), 23, 59, 59, 999999999, rangeVal.To.Location())
			return fmt.Sprintf("%s >= ? AND %s <= ?", field, field), []any{startOfFromDay, endOfToDay}
		}
	}
	return "", nil
}

func (f *Pagination[T]) buildTimeCondition(field string, mode Mode, value any) (string, []any) {
	switch mode {
	case ModeInside:
		rangeVal, err := parseRangeTime(value)
		if err != nil {
			return "", nil
		}
		fromStr := rangeVal.From.Format("15:04:05")
		toStr := rangeVal.To.Format("15:04:05")
		return fmt.Sprintf("time(%s) BETWEEN ? AND ?", field), []any{fromStr, toStr}
	case ModeOutside:
		rangeVal, err := parseRangeTime(value)
		if err != nil {
			return "", nil
		}
		fromStr := rangeVal.From.Format("15:04:05")
		toStr := rangeVal.To.Format("15:04:05")
		return fmt.Sprintf("time(%s) < ? OR time(%s) > ?", field, field), []any{fromStr, toStr}
	case ModeEqual:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) = ?", field), []any{timeStr}
	case ModeNotEqual:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) != ?", field), []any{timeStr}
	case ModeGT:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) > ?", field), []any{timeStr}
	case ModeGTE, ModeAfter:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) >= ?", field), []any{timeStr}
	case ModeLT, ModeBefore:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) < ?", field), []any{timeStr}
	case ModeLTE:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		timeStr := t.Format("15:04:05")
		return fmt.Sprintf("time(%s) <= ?", field), []any{timeStr}
	case ModeRange:
		rangeVal, err := parseRangeTime(value)
		if err != nil {
			return "", nil
		}
		fromStr := rangeVal.From.Format("15:04:05")
		toStr := rangeVal.To.Format("15:04:05")
		return fmt.Sprintf("time(%s) BETWEEN ? AND ?", field), []any{fromStr, toStr}
	}
	return "", nil
}

func (r *Pagination[T]) applySQLFilters(db *gorm.DB, filters []ArrFilterSQL) *gorm.DB {
	for _, f := range filters {
		switch f.Op {
		case ModeEqual:
			db = db.Where(f.Field+" = ?", f.Value)
		case ModeGT:
			db = db.Where(f.Field+" > ?", f.Value)
		case ModeGTE:
			db = db.Where(f.Field+" >= ?", f.Value)
		case ModeLT:
			db = db.Where(f.Field+" < ?", f.Value)
		case ModeLTE:
			db = db.Where(f.Field+" <= ?", f.Value)
		case ModeNotEqual:
			db = db.Where(f.Field+" <> ?", f.Value)
		case ModeInside:
			db = db.Where(f.Field+" IN (?)", f.Value)
		case ModeOutside:
			db = db.Where(f.Field+" NOT IN (?)", f.Value)
		case ModeContains:
			db = db.Where(f.Field+" LIKE ?", f.Value)
		case ModeIsEmpty:
			db = db.Where(f.Field + " IS NULL")
		case ModeIsNotEmpty:
			db = db.Where(f.Field + " IS NOT NULL")
		}
	}
	return db
}

func (f *Pagination[T]) applySort(db *gorm.DB, sortFields []ArrFilterSortSQL) *gorm.DB {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil || stmt.Schema == nil {
		return db
	}
	for _, sort := range sortFields {
		field := sort.Field
		order := strings.ToUpper(strings.TrimSpace(string(sort.Order)))
		if order != "DESC" {
			order = "ASC"
		}
		if strings.Contains(field, ".") {
			parts := strings.Split(field, ".")
			if len(parts) >= 2 {
				if !f.fieldExists(db, parts[0]) {
					continue
				}
				field = fmt.Sprintf(`"%s"."%s"`, toSnakeCase(parts[0]), parts[1])
				for i := 2; i < len(parts); i++ {
					field = fmt.Sprintf(`%s."%s"`, field, parts[i])
				}
			}
		} else {
			if f.fieldExists(db, field) {
				field = fmt.Sprintf(`"%s"`, field)
			} else {
				continue
			}
		}
		db = db.Order(fmt.Sprintf("%s %s", field, order))
	}
	return db
}

func (f *Pagination[T]) fieldExists(db *gorm.DB, field string) bool {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil || stmt.Schema == nil {
		return false
	}
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		currentSchema := stmt.Schema
		for i, part := range parts {
			fObj := currentSchema.LookUpField(part)
			if fObj == nil {
				fObj = currentSchema.LookUpField(toSnakeCase(part))
			}
			if fObj == nil {
				return false
			}
			if i < len(parts)-1 {
				rel, ok := currentSchema.Relationships.Relations[part]
				if !ok || rel.FieldSchema == nil {
					return false
				}
				currentSchema = rel.FieldSchema
			}
		}
		return true
	}
	if stmt.Schema.LookUpField(field) != nil {
		return true
	}
	if stmt.Schema.LookUpField(toSnakeCase(field)) != nil {
		return true
	}
	if stmt.Schema.LookUpField(strings.ToLower(field)) != nil {
		return true
	}

	return false
}

func (f *Pagination[T]) applyJoinsForFilters(db *gorm.DB, filters []ArrFilterSQL) *gorm.DB {
	joinMap := make(map[string]bool)

	for _, filter := range filters {
		if strings.Contains(filter.Field, ".") {
			parts := strings.Split(filter.Field, ".")
			if len(parts) >= 2 {
				if !f.fieldExists(db, parts[0]) {
					continue
				}
				relationName := toSnakeCase(parts[0])
				if !joinMap[relationName] {
					db = db.Joins(relationName)
					joinMap[relationName] = true
				}
			}
		}
	}
	return db
}
