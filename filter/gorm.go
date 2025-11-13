package filter

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ApplyPresetConditions applies struct fields as WHERE conditions to the db query.
// This is a helper to easily apply preset filters from a struct.
//
// Example usage:
//
//	type AccountTag struct {
//	    OrganizationID uint
//	    BranchID       uint
//	}
//
//	tag := &AccountTag{OrganizationID: 1, BranchID: 2}
//	db = filter.ApplyPresetConditions(db, tag)
//	result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
func ApplyPresetConditions(db *gorm.DB, conditions any) *gorm.DB {
	if conditions == nil {
		return db
	}
	return db.Where(conditions)
}

// DataGorm performs database-level filtering using GORM queries.
// It generates SQL WHERE clauses based on the filter configuration and returns paginated results.
// The db parameter can have existing WHERE conditions (e.g., organization_id, branch_id),
// and DataGorm will apply additional filters from filterRoot on top of those.
//
// Example with preset conditions using struct:
//
//	type AccountTag struct {
//	    OrganizationID uint
//	    BranchID       uint
//	}
//	tag := &AccountTag{OrganizationID: user.OrganizationID, BranchID: *user.BranchID}
//	db = filter.ApplyPresetConditions(db, tag)
//	result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
//
// Example with preset conditions using Where:
//
//	presetDB := db.Where("organization_id = ? AND branch_id = ?", orgID, branchID)
//	result, err := handler.DataGorm(presetDB, filterRoot, pageIndex, pageSize)
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

	// Set defaults if not provided - use 0-based indexing
	if result.PageIndex < 0 {
		result.PageIndex = 0
	}
	if result.PageSize <= 0 {
		result.PageSize = 30
	}

	// Build the query - db may already have WHERE conditions, they will be preserved
	query := db.Model(new(T))

	// Auto-join related tables based on field filters and sort fields
	query = f.autoJoinRelatedTables(query, filterRoot.FieldFilters, filterRoot.SortFields)

	// Apply preloads (GORM only feature)
	if len(filterRoot.Preload) > 0 {
		for _, preloadField := range filterRoot.Preload {
			query = query.Preload(preloadField)
		}
	}

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

	// Check if any filters or sorts use nested fields (for table name disambiguation)
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

	// Get the main table name for disambiguation
	var mainTableName string
	if hasNestedFields {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(new(T)); err == nil {
			mainTableName = stmt.Schema.Table
		}
	}

	// Apply sorting
	if len(filterRoot.SortFields) > 0 {
		for _, sortField := range filterRoot.SortFields {
			// For simple fields, check if they exist. For nested fields, let GORM handle them.
			if !strings.Contains(sortField.Field, ".") && !f.fieldExists(sortField.Field) {
				// Silently ignore non-existent simple sort fields
				continue
			}

			order := "ASC"
			if sortField.Order == SortOrderDesc {
				order = "DESC"
			}
			// Normalize nested field names: "member_profile.name" -> "MemberProfile.name"
			field := sortField.Field
			if strings.Contains(field, ".") {
				parts := strings.Split(field, ".")
				if len(parts) >= 2 {
					parts[0] = f.toPascalCase(parts[0])
					// Quote identifiers to preserve case
					field = fmt.Sprintf(`"%s"."%s"`, parts[0], parts[1])
					for i := 2; i < len(parts); i++ {
						field = fmt.Sprintf(`%s."%s"`, field, parts[i])
					}
				}
			} else if mainTableName != "" {
				// For non-nested fields, prefix with main table name to avoid ambiguity
				field = fmt.Sprintf(`"%s"."%s"`, mainTableName, field)
			}
			query = query.Order(fmt.Sprintf("%s %s", field, order))
		}
	}

	// Apply pagination (0-based indexing)
	offset := result.PageIndex * result.PageSize
	query = query.Offset(int(offset)).Limit(int(result.PageSize))

	// Execute query
	var data []*T
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}

	result.Data = data
	return &result, nil
}

// DataGormNoPage performs database-level filtering using GORM queries without pagination.
// It generates SQL WHERE clauses based on the filter configuration and returns all matching results as a simple array.
// The db parameter can have existing WHERE conditions (e.g., organization_id, branch_id),
// and DataGormNoPage will apply additional filters from filterRoot on top of those.
//
// Example with preset conditions using struct:
//
//	type AccountTag struct {
//	    OrganizationID uint
//	    BranchID       uint
//	}
//	tag := &AccountTag{OrganizationID: user.OrganizationID, BranchID: *user.BranchID}
//	db = filter.ApplyPresetConditions(db, tag)
//	results, err := handler.DataGormNoPage(db, filterRoot)
//
// Example with preset conditions using Where:
//
//	presetDB := db.Where("organization_id = ? AND branch_id = ?", orgID, branchID)
//	results, err := handler.DataGormNoPage(presetDB, filterRoot)
func (f *Handler[T]) DataGormNoPage(
	db *gorm.DB,
	filterRoot Root,
) ([]*T, error) {
	// Build the query - db may already have WHERE conditions, they will be preserved
	query := db.Model(new(T))

	// Auto-join related tables based on field filters and sort fields
	query = f.autoJoinRelatedTables(query, filterRoot.FieldFilters, filterRoot.SortFields)

	// Apply preloads (GORM only feature)
	if len(filterRoot.Preload) > 0 {
		for _, preloadField := range filterRoot.Preload {
			query = query.Preload(preloadField)
		}
	}

	// Apply filters
	if len(filterRoot.FieldFilters) > 0 {
		query = f.applysGorm(query, filterRoot)
	}

	// Check if any filters or sorts use nested fields (for table name disambiguation)
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

	// Get the main table name for disambiguation
	var mainTableName string
	if hasNestedFields {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(new(T)); err == nil {
			mainTableName = stmt.Schema.Table
		}
	}

	// Apply sorting
	if len(filterRoot.SortFields) > 0 {
		for _, sortField := range filterRoot.SortFields {
			// For simple fields, check if they exist. For nested fields, let GORM handle them.
			if !strings.Contains(sortField.Field, ".") && !f.fieldExists(sortField.Field) {
				// Silently ignore non-existent simple sort fields
				continue
			}

			order := "ASC"
			if sortField.Order == SortOrderDesc {
				order = "DESC"
			}
			// Normalize nested field names: "member_profile.name" -> "MemberProfile.name"
			field := sortField.Field
			if strings.Contains(field, ".") {
				parts := strings.Split(field, ".")
				if len(parts) >= 2 {
					parts[0] = f.toPascalCase(parts[0])
					// Quote identifiers to preserve case
					field = fmt.Sprintf(`"%s"."%s"`, parts[0], parts[1])
					for i := 2; i < len(parts); i++ {
						field = fmt.Sprintf(`%s."%s"`, field, parts[i])
					}
				}
			} else if mainTableName != "" {
				// For non-nested fields, prefix with main table name to avoid ambiguity
				field = fmt.Sprintf(`"%s"."%s"`, mainTableName, field)
			}
			query = query.Order(fmt.Sprintf("%s %s", field, order))
		}
	}

	// Execute query without pagination
	var data []*T
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}

	return data, nil
}

// GormNoPaginationCSV performs database-level filtering using GORM queries and returns results as CSV bytes.
// It generates SQL WHERE clauses based on the filter configuration and exports all matching results as CSV format.
// Field names are automatically used as CSV headers.
// The db parameter can have existing WHERE conditions (e.g., organization_id, branch_id),
// and GormNoPaginationCSV will apply additional filters from filterRoot on top of those.
//
// Example with preset conditions using struct:
//
//	type AccountTag struct {
//	    OrganizationID uint
//	    BranchID       uint
//	}
//	tag := &AccountTag{OrganizationID: user.OrganizationID, BranchID: *user.BranchID}
//	db = filter.ApplyPresetConditions(db, tag)
//	csvData, err := handler.GormNoPaginationCSV(db, filterRoot)
//
// Example with preset conditions using Where:
//
//	presetDB := db.Where("organization_id = ? AND branch_id = ?", orgID, branchID)
//	csvData, err := handler.GormNoPaginationCSV(presetDB, filterRoot)
func (f *Handler[T]) GormNoPaginationCSV(
	db *gorm.DB,
	filterRoot Root,
) ([]byte, error) {
	// Use DataGormNoPage to get filtered results
	filteredData, err := f.DataGormNoPage(db, filterRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to filter data: %w", err)
	}

	// Write headers using field names from the getters map
	fieldNames := make([]string, 0, len(f.getters))
	for fieldName := range f.getters {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames) // Ensure deterministic column ordering

	// Build CSV content using encoding/csv
	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)

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

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// DataGormWithPreset is a convenience method that combines ApplyPresetConditions and DataGorm.
// It accepts preset conditions as a struct and applies them before filtering.
//
// Example usage:
//
//	type AccountTag struct {
//	    OrganizationID uint `gorm:"column:organization_id"`
//	    BranchID       uint `gorm:"column:branch_id"`
//	}
//
//	tag := &AccountTag{
//	    OrganizationID: user.OrganizationID,
//	    BranchID:       *user.BranchID,
//	}
//	result, err := handler.DataGormWithPreset(db, tag, filterRoot, pageIndex, pageSize)
func (f *Handler[T]) DataGormWithPreset(
	db *gorm.DB,
	presetConditions any,
	filterRoot Root,
	pageIndex int,
	pageSize int,
) (*PaginationResult[T], error) {
	// Apply preset conditions to db
	if presetConditions != nil {
		db = db.Where(presetConditions)
	}

	// Call regular DataGorm with the modified db
	return f.DataGorm(db, filterRoot, pageIndex, pageSize)
}

// DataGormNoPageWithPreset is a convenience method that combines ApplyPresetConditions and DataGormNoPage.
// It accepts preset conditions as a struct and applies them before filtering, returning results without pagination.
//
// Example usage:
//
//	type AccountTag struct {
//	    OrganizationID uint `gorm:"column:organization_id"`
//	    BranchID       uint `gorm:"column:branch_id"`
//	}
//
//	tag := &AccountTag{
//	    OrganizationID: user.OrganizationID,
//	    BranchID:       *user.BranchID,
//	}
//	results, err := handler.DataGormNoPageWithPreset(db, tag, filterRoot)
func (f *Handler[T]) DataGormNoPageWithPreset(
	db *gorm.DB,
	presetConditions any,
	filterRoot Root,
) ([]*T, error) {
	// Apply preset conditions to db
	if presetConditions != nil {
		db = db.Where(presetConditions)
	}

	// Call DataGormNoPage with the modified db
	return f.DataGormNoPage(db, filterRoot)
}

// GormNoPaginationCSVWithPreset is a convenience method that combines ApplyPresetConditions and GormNoPaginationCSV.
// It accepts preset conditions as a struct and applies them before filtering, returning results as CSV without pagination.
//
// Example usage:
//
//	type AccountTag struct {
//	    OrganizationID uint `gorm:"column:organization_id"`
//	    BranchID       uint `gorm:"column:branch_id"`
//	}
//
//	tag := &AccountTag{
//	    OrganizationID: user.OrganizationID,
//	    BranchID:       *user.BranchID,
//	}
//	csvData, err := handler.GormNoPaginationCSVWithPreset(db, tag, filterRoot)
func (f *Handler[T]) GormNoPaginationCSVWithPreset(
	db *gorm.DB,
	presetConditions any,
	filterRoot Root,
) ([]byte, error) {
	// Apply preset conditions to db
	if presetConditions != nil {
		db = db.Where(presetConditions)
	}

	// Call GormNoPaginationCSV with the modified db
	return f.GormNoPaginationCSV(db, filterRoot)
}

// GormNoPaginationCSVCustom performs database-level filtering and returns results as CSV bytes.
// It uses a custom callback function to allow users to define exactly what fields and values to include in the CSV output.
// This provides full control over CSV structure and field mapping on the user side.
//
// Parameters:
//   - db: GORM database instance with any preset conditions
//   - filterRoot: filter configuration defining conditions, logic, and sorting
//   - customGetter: callback function that takes a data item and returns a map[string]any
//     where keys are column headers and values are the corresponding data
//
// Returns CSV bytes with headers from the customGetter map keys, sorted alphabetically for deterministic ordering.
//
// Example usage:
//
//	csvData, err := handler.GormNoPaginationCSVCustom(db, filterRoot, func(user *User) map[string]any {
//	    return map[string]any{
//	        "Full Name": user.FirstName + " " + user.LastName,
//	        "Email": user.Email,
//	        "Status": user.IsActive,
//	        "Department": user.Department.Name, // Access nested fields if preloaded
//	    }
//	})
func (f *Handler[T]) GormNoPaginationCSVCustom(
	db *gorm.DB,
	filterRoot Root,
	customGetter func(*T) map[string]any,
) ([]byte, error) {
	// Apply filters to database query
	filteredDB := f.applysGorm(db, filterRoot)

	// Apply sorting
	if len(filterRoot.SortFields) > 0 {
		for _, sortField := range filterRoot.SortFields {
			// For simple fields, check if they exist. For nested fields, let GORM handle them.
			if !strings.Contains(sortField.Field, ".") && !f.fieldExists(sortField.Field) {
				// Silently ignore non-existent simple sort fields
				continue
			}

			order := "ASC"
			if sortField.Order == SortOrderDesc {
				order = "DESC"
			}
			filteredDB = filteredDB.Order(fmt.Sprintf("%s %s", sortField.Field, order))
		}
	}

	// Execute query to get all matching records
	var results []*T
	if err := filteredDB.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	if len(results) == 0 {
		// If no data, return empty CSV
		return []byte(""), nil
	}

	// Get headers from the first item using the custom getter
	firstItemFields := customGetter(results[0])

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
	for _, item := range results {
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

// GormNoPaginationCSVCustomWithPreset is a convenience method that combines preset conditions with GormNoPaginationCSVCustom.
// It applies preset conditions to the database query before filtering and CSV export using a custom getter function.
//
// Parameters:
//   - db: GORM database instance
//   - presetConditions: struct or map with preset WHERE conditions to apply before filtering
//   - filterRoot: filter configuration defining conditions, logic, and sorting
//   - customGetter: callback function that defines custom CSV field mapping
//
// Example usage:
//
//	type UserFilter struct {
//	    OrganizationID uint
//	    BranchID       uint
//	}
//
//	presetConditions := &UserFilter{
//	    OrganizationID: user.OrganizationID,
//	    BranchID:       *user.BranchID,
//	}
//
//	csvData, err := handler.GormNoPaginationCSVCustomWithPreset(db, presetConditions, filterRoot, func(user *User) map[string]any {
//	    return map[string]any{
//	        "Employee ID": user.ID,
//	        "Full Name":   fmt.Sprintf("%s %s", user.FirstName, user.LastName),
//	        "Department":  user.Department.Name,
//	    }
//	})
func (f *Handler[T]) GormNoPaginationCSVCustomWithPreset(
	db *gorm.DB,
	presetConditions any,
	filterRoot Root,
	customGetter func(*T) map[string]any,
) ([]byte, error) {
	// Apply preset conditions to db
	if presetConditions != nil {
		db = db.Where(presetConditions)
	}

	// Call GormNoPaginationCSVCustom with the modified db
	return f.GormNoPaginationCSVCustom(db, filterRoot, customGetter)
}

func (f *Handler[T]) applysGorm(db *gorm.DB, filterRoot Root) *gorm.DB {
	if len(filterRoot.FieldFilters) == 0 {
		return db
	}

	// Check if any filters use nested fields (which trigger JOINs)
	hasNestedFields := false
	for _, filter := range filterRoot.FieldFilters {
		if strings.Contains(filter.Field, ".") {
			hasNestedFields = true
			break
		}
	}

	// Get the main table name for disambiguation
	var mainTableName string
	if hasNestedFields {
		// Get table name from GORM
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(new(T)); err == nil {
			mainTableName = stmt.Schema.Table
		}
	}

	if filterRoot.Logic == LogicAnd {
		for _, filter := range filterRoot.FieldFilters {
			// For simple fields, check if they exist. For nested fields, let GORM handle them.
			if strings.Contains(filter.Field, ".") || f.fieldExists(filter.Field) {
				db = f.applyGormWithTableName(db, filter, mainTableName)
			}
			// Silently ignore non-existent simple fields
		}
	} else {
		var orConditions []string
		var orValues []any

		for _, filter := range filterRoot.FieldFilters {
			// For simple fields, check if they exist. For nested fields, let GORM handle them.
			if strings.Contains(filter.Field, ".") || f.fieldExists(filter.Field) {
				condition, values := f.buildConditionWithTableName(filter, mainTableName)
				if condition != "" {
					orConditions = append(orConditions, condition)
					orValues = append(orValues, values...)
				}
			}
			// Silently ignore non-existent fields
		}
		if len(orConditions) > 0 {
			db = db.Where(strings.Join(orConditions, " OR "), orValues...)
		}
	}
	return db
}

// toPascalCase converts snake_case or lowercase to PascalCase
// Examples: "member_profile" -> "MemberProfile", "currency" -> "Currency"
func (f *Handler[T]) toPascalCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Split by underscore for snake_case
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// applyGormWithTableName applies a single filter with table name disambiguation
func (f *Handler[T]) applyGormWithTableName(db *gorm.DB, filter FieldFilter, mainTableName string) *gorm.DB {
	condition, values := f.buildConditionWithTableName(filter, mainTableName)
	if condition != "" {
		db = db.Where(condition, values...)
	}
	return db
}

// buildConditionWithTableName builds SQL condition with optional table name prefix for non-nested fields
func (f *Handler[T]) buildConditionWithTableName(filter FieldFilter, mainTableName string) (string, []any) {
	field := filter.Field
	value := filter.Value

	// Check if this is a nested field
	isNestedField := strings.Contains(field, ".")

	// For nested fields, we need to normalize the relationship name to match GORM's struct field names
	// Example: "currency.currency_code" should become "Currency.currency_code"
	// We also need to quote identifiers to preserve case for PostgreSQL
	if isNestedField {
		parts := strings.Split(field, ".")
		if len(parts) >= 2 {
			// Convert the first part (relationship name) to PascalCase to match struct field name
			// GORM uses the struct field name for JOINs
			parts[0] = f.toPascalCase(parts[0])
			// Quote identifiers to preserve case in PostgreSQL
			// Format: "RelationName"."field_name"
			field = fmt.Sprintf(`"%s"."%s"`, parts[0], parts[1])
			// For more than 2 parts, append remaining parts
			for i := 2; i < len(parts); i++ {
				field = fmt.Sprintf(`%s."%s"`, field, parts[i])
			}
		}
	} else if mainTableName != "" {
		// For non-nested fields, prefix with main table name to avoid ambiguity when JOINs are present
		// Quote both table and field names
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
	// Handle Range mode separately since value is a Range struct, not a string
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

	// For all other modes, parse value as text
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
		// Support for text comparison (useful for time strings like "08:00:00")
		return fmt.Sprintf("%s > ?", field), []any{str}
	case ModeGTE, ModeAfter:
		// Support for text comparison (useful for time strings like "08:00:00")
		return fmt.Sprintf("%s >= ?", field), []any{str}
	case ModeLT, ModeBefore:
		// Support for text comparison (useful for time strings like "08:00:00")
		return fmt.Sprintf("%s < ?", field), []any{str}
	case ModeLTE:
		// Support for text comparison (useful for time strings like "08:00:00")
		return fmt.Sprintf("%s <= ?", field), []any{str}
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

// buildTimeCondition builds SQL condition for time filters
func (f *Handler[T]) buildTimeCondition(field string, mode Mode, value any) (string, []any) {
	switch mode {
	case ModeEqual:
		t, err := parseTime(value)
		if err != nil {
			return "", nil
		}
		// Format time as HH:MM:SS for SQLite TEXT comparison
		// Use time() function to extract time from datetime columns
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

// autoJoinRelatedTables automatically joins related tables when filters or sort fields reference nested fields
func (f *Handler[T]) autoJoinRelatedTables(db *gorm.DB, filters []FieldFilter, sortFields []SortField) *gorm.DB {
	joinedTables := make(map[string]bool)

	// Check filters for nested fields
	for _, filter := range filters {
		// For GORM operations, allow nested fields even if they're not in getters map
		// GORM can handle nested relations through auto-joins
		if strings.Contains(filter.Field, ".") {
			parts := strings.Split(filter.Field, ".")
			if len(parts) >= 2 {
				// Convert snake_case/lowercase to PascalCase (e.g., "member_profile" -> "MemberProfile")
				tableName := f.toPascalCase(parts[0])
				if !joinedTables[tableName] {
					// GORM will auto-join based on the relationship
					db = db.Joins(tableName)
					joinedTables[tableName] = true
				}
			}
		}
	}

	// Check sort fields for nested fields
	for _, sortField := range sortFields {
		// For GORM operations, allow nested fields even if they're not in getters map
		// GORM can handle nested relations through auto-joins
		if strings.Contains(sortField.Field, ".") {
			parts := strings.Split(sortField.Field, ".")
			if len(parts) >= 2 {
				// Convert snake_case/lowercase to PascalCase
				tableName := f.toPascalCase(parts[0])
				if !joinedTables[tableName] {
					// GORM will auto-join based on the relationship
					db = db.Joins(tableName)
					joinedTables[tableName] = true
				}
			}
		}
	}

	return db
}
