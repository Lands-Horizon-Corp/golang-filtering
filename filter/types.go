package filter

import "time"

// FilterMode defines the type of comparison operation to perform
type FilterMode string

// Filter mode constants define available comparison operations
const (
	FilterModeEqual       FilterMode = "equal"       // Exact match
	FilterModeNotEqual    FilterMode = "notEqual"    // Not equal
	FilterModeContains    FilterMode = "contains"    // Contains substring
	FilterModeNotContains FilterMode = "notContains" // Does not contain substring
	FilterModeStartsWith  FilterMode = "startsWith"  // Starts with prefix
	FilterModeEndsWith    FilterMode = "endsWith"    // Ends with suffix
	FilterModeIsEmpty     FilterMode = "isEmpty"     // Is empty or null
	FilterModeIsNotEmpty  FilterMode = "isNotEmpty"  // Is not empty
	FilterModeGT          FilterMode = "gt"          // Greater than
	FilterModeGTE         FilterMode = "gte"         // Greater than or equal
	FilterModeLT          FilterMode = "lt"          // Less than
	FilterModeLTE         FilterMode = "lte"         // Less than or equal
	FilterModeRange       FilterMode = "range"       // Between two values
	FilterModeBefore      FilterMode = "before"      // Before (date/time)
	FilterModeAfter       FilterMode = "after"       // After (date/time)
)

// FilterDataType defines the data type being filtered
type FilterDataType string

// Filter data type constants define the type of data being filtered
const (
	FilterDataTypeNumber FilterDataType = "number" // Numeric values
	FilterDataTypeText   FilterDataType = "text"   // Text/string values
	FilterDataTypeBool   FilterDataType = "bool"   // Boolean values
	FilterDataTypeDate   FilterDataType = "date"   // Date values
	FilterDataTypeTime   FilterDataType = "time"   // Time values
)

// FilterLogic defines how multiple filters are combined
type FilterLogic string

// Filter logic constants define how to combine multiple filters
const (
	FilterLogicAnd FilterLogic = "and" // All filters must match
	FilterLogicOr  FilterLogic = "or"  // Any filter can match
)

// FilterSortOrder defines the sort direction
type FilterSortOrder string

// Sort order constants define ascending or descending order
const (
	FilterSortOrderAsc  FilterSortOrder = "asc"  // Ascending order
	FilterSortOrderDesc FilterSortOrder = "desc" // Descending order
)

// Filter represents a single filter condition
type Filter struct {
	Field          string         `json:"field"`          // Field name to filter on
	Value          any            `json:"value"`          // Value to compare against
	Mode           FilterMode     `json:"mode"`           // Comparison mode
	FilterDataType FilterDataType `json:"filterDataType"` // Data type of the field
}

// SortField represents a field to sort by
type SortField struct {
	Field string          `json:"field"` // Field name to sort by
	Order FilterSortOrder `json:"order"` // Sort direction
}

// FilterRoot represents the root filter configuration
type FilterRoot struct {
	Filters    []Filter    `json:"filters"`    // List of filter conditions
	SortFields []SortField `json:"sortFields"` // List of sort fields
	Logic      FilterLogic `json:"logic"`      // How to combine filters (AND/OR)
}

// FilterRange represents a range of values for filtering
type FilterRange struct {
	From any `json:"from"` // Start of range
	To   any `json:"to"`   // End of range
}

// PaginationResult contains filtered and paginated results
type PaginationResult[T any] struct {
	Data      []*T `json:"data"`      // Current page data
	TotalSize int  `json:"totalSize"` // Total matching records
	TotalPage int  `json:"totalPage"` // Total number of pages
	PageIndex int  `json:"pageIndex"` // Current page index (1-based)
	PageSize  int  `json:"pageSize"`  // Records per page
}

// FilterRangeNumber represents a numeric range
type FilterRangeNumber struct {
	From float64 // Start of numeric range
	To   float64 // End of numeric range
}

// FilterRangeDate represents a date range
type FilterRangeDate struct {
	From time.Time // Start date
	To   time.Time // End date
}
