package filter

import "time"

// Mode defines the type of comparison operation to perform
type Mode string

// mode constants define available comparison operations
const (
	ModeEqual       Mode = "equal"       // Exact match
	ModeNotEqual    Mode = "notEqual"    // Not equal
	ModeContains    Mode = "contains"    // Contains substring
	ModeNotContains Mode = "notContains" // Does not contain substring
	ModeStartsWith  Mode = "startsWith"  // Starts with prefix
	ModeEndsWith    Mode = "endsWith"    // Ends with suffix
	ModeIsEmpty     Mode = "isEmpty"     // Is empty or null
	ModeIsNotEmpty  Mode = "isNotEmpty"  // Is not empty
	ModeGT          Mode = "gt"          // Greater than
	ModeGTE         Mode = "gte"         // Greater than or equal
	ModeLT          Mode = "lt"          // Less than
	ModeLTE         Mode = "lte"         // Less than or equal
	ModeRange       Mode = "range"       // Between two values
	ModeBefore      Mode = "before"      // Before (date/time)
	ModeAfter       Mode = "after"       // After (date/time)
)

// DataType defines the data type being filtered
type DataType string

// data type constants define the type of data being filtered
const (
	DataTypeNumber DataType = "number" // Numeric values
	DataTypeText   DataType = "text"   // Text/string values
	DataTypeBool   DataType = "bool"   // Boolean values
	DataTypeDate   DataType = "date"   // Date values
	DataTypeTime   DataType = "time"   // Time values
)

// Logic defines how multiple filters are combined
type Logic string

// logic constants define how to combine multiple filters
const (
	LogicAnd Logic = "and" // All filters must match
	LogicOr  Logic = "or"  // Any filter can match
)

// SortOrder defines the sort direction
type SortOrder string

// Sort order constants define ascending or descending order
const (
	SortOrderAsc  SortOrder = "asc"  // Ascending order
	SortOrderDesc SortOrder = "desc" // Descending order
)

// represents a single filter condition
type FieldFilter struct {
	Field    string   `json:"field"`    // Field name to filter on
	Value    any      `json:"value"`    // Value to compare against
	Mode     Mode     `json:"mode"`     // Comparison mode
	DataType DataType `json:"dataType"` // Data type of the field
}

// SortField represents a field to sort by
type SortField struct {
	Field string    `json:"field"` // Field name to sort by
	Order SortOrder `json:"order"` // Sort direction
}

// Root represents the root filter configuration
type Root struct {
	FieldFilters []FieldFilter `json:"filters"`    // List of filter conditions
	SortFields   []SortField   `json:"sortFields"` // List of sort fields
	Logic        Logic         `json:"logic"`      // How to combine filters (AND/OR)
	Preload      []string      `json:"preload"`    // List of related entities to preload (only applicable for GORM)
}

// Range represents a range of values for filtering
type Range struct {
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

// RangeNumber represents a numeric range
type RangeNumber struct {
	From float64 // Start of numeric range
	To   float64 // End of numeric range
}

// RangeDate represents a date range
type RangeDate struct {
	From time.Time // Start date
	To   time.Time // End date
}
