package query

import "time"

type Mode string

const (
	ModeEqual       Mode = "equal"
	ModeNotEqual    Mode = "notEqual"
	ModeContains    Mode = "contains"
	ModeNotContains Mode = "notContains"
	ModeStartsWith  Mode = "startsWith"
	ModeEndsWith    Mode = "endsWith"

	ModeInside  Mode = "inside"
	ModeOutside Mode = "outside"

	ModeGT     Mode = "gt"
	ModeGTE    Mode = "gte"
	ModeLT     Mode = "lt"
	ModeLTE    Mode = "lte"
	ModeRange  Mode = "range"
	ModeBefore Mode = "before"
	ModeAfter  Mode = "after"

	ModeIsEmpty    Mode = "isEmpty"
	ModeIsNotEmpty Mode = "isNotEmpty"
)

type DataType string

const (
	DataTypeNumber DataType = "number"
	DataTypeText   DataType = "text"
	DataTypeBool   DataType = "bool"
	DataTypeDate   DataType = "date"
	DataTypeTime   DataType = "time"
)

type Logic string

const (
	LogicAnd Logic = "and"
	LogicOr  Logic = "or"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type FieldFilter struct {
	Field    string   `json:"field"`
	Value    any      `json:"value"`
	Mode     Mode     `json:"mode"`
	DataType DataType `json:"dataType"`
}

type SortField struct {
	Field string    `json:"field"`
	Order SortOrder `json:"order"`
}

type StructuredFilter struct {
	FieldFilters []FieldFilter `json:"filters"`
	SortFields   []SortField   `json:"sortFields"`
	Logic        Logic         `json:"logic"`
	Preload      []string      `json:"preload"`
}

type Range struct {
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

type RangeNumber struct {
	From float64
	To   float64
}

type RangeDate struct {
	From time.Time
	To   time.Time
}

type ArrFilterSQL struct {
	Field string
	Op    Mode
	Value any
}

type ArrFilterSortSQL struct {
	Field string
	Order SortOrder
}
