package filter

type FilterLogic string
type DataType string
type FilterMode string

const (
	FilterLogicAnd FilterLogic = "AND"
	FilterLogicOr  FilterLogic = "OR"
)

const (
	DataTypeNumber DataType = "number"
	DataTypeText   DataType = "text"
	DataTypeDate   DataType = "date"
	DataTypeBool   DataType = "boolean"
	DataTypeTime   DataType = "time"
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

type Filter struct {
	DataType DataType `json:"dataType"`
	Field    string   `json:"field"`
	Mode     string   `json:"mode"`
	Value    any      `json:"value"`
}

type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"`
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
