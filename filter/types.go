package filter

import "time"

type FilterMode string

const (
	FilterModeEqual       FilterMode = "equal"
	FilterModeNotEqual    FilterMode = "notEqual"
	FilterModeContains    FilterMode = "contains"
	FilterModeNotContains FilterMode = "notContains"
	FilterModeStartsWith  FilterMode = "startsWith"
	FilterModeEndsWith    FilterMode = "endsWith"
	FilterModeIsEmpty     FilterMode = "isEmpty"
	FilterModeIsNotEmpty  FilterMode = "isNotEmpty"
	FilterModeGT          FilterMode = "gt"
	FilterModeGTE         FilterMode = "gte"
	FilterModeLT          FilterMode = "lt"
	FilterModeLTE         FilterMode = "lte"
	FilterModeRange       FilterMode = "range"
	FilterModeBefore      FilterMode = "before"
	FilterModeAfter       FilterMode = "after"
)

type FilterDataType string

const (
	FilterDataTypeNumber FilterDataType = "number"
	FilterDataTypeText   FilterDataType = "text"
	FilterDataTypeBool   FilterDataType = "bool"
	FilterDataTypeDate   FilterDataType = "date"
	FilterDataTypeTime   FilterDataType = "time"
)

type FilterLogic string

const (
	FilterLogicAnd FilterLogic = "and"
	FilterLogicOr  FilterLogic = "or"
)

type FilterSortOrder string

const (
	FilterSortOrderAsc  FilterSortOrder = "asc"
	FilterSortOrderDesc FilterSortOrder = "desc"
)

type Filter struct {
	Field          string         `json:"field"`
	Value          any            `json:"value"`
	Mode           FilterMode     `json:"mode"`
	FilterDataType FilterDataType `json:"filterDataType"`
}

type SortField struct {
	Field string          `json:"field"`
	Order FilterSortOrder `json:"order"`
}

type FilterRoot struct {
	Filters    []Filter    `json:"filters"`
	SortFields []SortField `json:"sortFields"`
	Logic      FilterLogic `json:"logic"`
}

type FilterRange struct {
	From any `json:"from"`
	To   any `json:"to"`
}

type PaginationResult[T any] struct {
	Data      []*T `json:"data"`
	TotalSize int  `json:"totalSize"`
	TotalPage int  `json:"totalPage"`
	PageIndex int  `json:"pageIndex"`
	PageSize  int  `json:"pageSize"`
}

type FilterRangeNumber struct {
	From float64
	To   float64
}

type FilterRangeDate struct {
	From time.Time
	To   time.Time
}
