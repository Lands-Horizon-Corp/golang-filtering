package query_test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/stretchr/testify/assert"
)

func TestStrParseFilters(t *testing.T) {
	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 30, Mode: query.ModeEqual, DataType: query.DataTypeNumber},
		},
	}
	encoded := encodeFilter(filter)
	result, err := query.StrParseFilters(encoded)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.FieldFilters))
	assert.Equal(t, "age", result.FieldFilters[0].Field)
	assert.Equal(t, 30.0, result.FieldFilters[0].Value)
}

func TestStrParseSort(t *testing.T) {
	sort := []query.SortField{
		{Field: "name", Order: query.SortOrderAsc},
	}
	encoded := encodeSort(sort)
	result, err := query.StrParseSort(encoded)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "name", result[0].Field)
	assert.Equal(t, query.SortOrderAsc, result[0].Order)
}

func TestStrParsePageIndexAndSize(t *testing.T) {
	index, err := query.StrParsePageIndex("2")
	assert.NoError(t, err)
	assert.Equal(t, 2, index)

	size, err := query.StrParsePageSize("50")
	assert.NoError(t, err)
	assert.Equal(t, 50, size)

	index, err = query.StrParsePageIndex("")
	assert.NoError(t, err)
	assert.Equal(t, 0, index)

	size, err = query.StrParsePageSize("")
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
}

func TestStrParseQuery(t *testing.T) {
	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 35, Mode: query.ModeEqual, DataType: query.DataTypeNumber},
		},
	}
	sort := []query.SortField{
		{Field: "name", Order: query.SortOrderDesc},
	}

	filterEncoded := encodeFilter(filter)
	sortEncoded := encodeSort(sort)
	pageIndex := "1"
	pageSize := "10"

	queryStr := filterEncoded + "|" + sortEncoded + "|" + pageIndex + "|" + pageSize

	resultFilter, idx, size, err := query.StrParseQuery(queryStr)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resultFilter.FieldFilters))
	assert.Equal(t, "age", resultFilter.FieldFilters[0].Field)
	assert.Equal(t, 35.0, resultFilter.FieldFilters[0].Value)
	assert.Equal(t, 1, idx)
	assert.Equal(t, 10, size)
	assert.Equal(t, 1, len(resultFilter.SortFields))
	assert.Equal(t, "name", resultFilter.SortFields[0].Field)
	assert.Equal(t, query.SortOrderDesc, resultFilter.SortFields[0].Order)
}
