package filter

import "time"

type FilterLogic string
type FilterDataType string
type FilterMode string
type FilterSortOrder string

const (
	FilterSortOrderAsc  FilterSortOrder = "ASC"
	FilterSortOrderDesc FilterSortOrder = "DESC"
)

const (
	FilterLogicAnd FilterLogic = "AND"
	FilterLogicOr  FilterLogic = "OR"
)

const (
	FilterDataTypeNumber FilterDataType = "number"
	FilterDataTypeText   FilterDataType = "text"
	FilterDataTypeDate   FilterDataType = "date"
	FilterDataTypeBool   FilterDataType = "boolean"
	FilterDataTypeTime   FilterDataType = "time"
)
const (
	FilterModeEqual       FilterMode = "equal"
	FilterModeNotEqual    FilterMode = "nequal"
	FilterModeContains    FilterMode = "contains"
	FilterModeNotContains FilterMode = "ncontains"
	FilterModeStartsWith  FilterMode = "startswith"
	FilterModeEndsWith    FilterMode = "endswith"
	FilterModeIsEmpty     FilterMode = "isempty"
	FilterModeIsNotEmpty  FilterMode = "isnotempty"
	FilterModeGT          FilterMode = "gt"
	FilterModeGTE         FilterMode = "gte"
	FilterModeLT          FilterMode = "lt"
	FilterModeLTE         FilterMode = "lte"
	FilterModeRange       FilterMode = "range"
	FilterModeBefore      FilterMode = "before"
	FilterModeAfter       FilterMode = "after"
)

type FilterRange struct {
	From any `json:"from"`
	To   any `json:"to"`
}
type Filter struct {
	FilterDataType FilterDataType `json:"FilterDataType"`
	Field          string         `json:"field"`
	Mode           FilterMode     `json:"mode"`
	Value          any            `json:"value"`
}

type SortField struct {
	Order FilterSortOrder `json:"order"`
	Field string          `json:"field"`
}
type FilterRoot struct {
	Filters    []Filter    `json:"filters"`
	SortFields []SortField `json:"sortFields"`
	Logic      FilterLogic `json:"logic"`
}

type PaginationResult[T any] struct {
	Data      []*T `json:"data"`
	PageIndex int  `json:"pageIndex"`
	TotalPage int  `json:"totalPage"`
	PageSize  int  `json:"pageSize"`
	TotalSize int  `json:"totalSize"`
}

type FilterRangeNumber struct {
	From float64 `json:"from"`
	To   float64 `json:"to"`
}

type FilterRangeTime struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type FilterRangeDate struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}
