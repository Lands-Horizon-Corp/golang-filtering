package query_test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
)

func TestDetectDataType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected query.DataType
	}{
		{"int", 42, query.DataTypeNumber},
		{"int8", int8(8), query.DataTypeNumber},
		{"int32", int32(32), query.DataTypeNumber},
		{"int64", int64(64), query.DataTypeNumber},
		{"uint", uint(10), query.DataTypeNumber},
		{"float32", float32(3.14), query.DataTypeNumber},
		{"float64", float64(3.14), query.DataTypeNumber},
		{"bool true", true, query.DataTypeBool},
		{"bool false", false, query.DataTypeBool},
		{"string", "hello", query.DataTypeText},
		{"time as date", time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC), query.DataTypeDate},
		{"time as time", time.Date(2025, 12, 10, 15, 30, 45, 0, time.UTC), query.DataTypeTime},
		{"nil value", nil, query.DataTypeText},
		{"pointer to int", func() *int { x := 5; return &x }(), query.DataTypeNumber},
		{"pointer to string", func() *string { s := "test"; return &s }(), query.DataTypeText},
		{"pointer to nil", (*int)(nil), query.DataTypeText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := query.DetectDataType(tt.input)
			if got != tt.expected {
				t.Errorf("DetectDataType(%v) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
