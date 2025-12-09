package query_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// Helper to create echo context from query string
func createEchoContext(queryStr string) echo.Context {
	req := httptest.NewRequest(http.MethodGet, "/?"+queryStr, nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	return e.NewContext(req, rec)
}

// Helper to encode filter to base64
func encodeFilter(filter query.StructuredFilter) string {
	data, _ := json.Marshal(filter)
	return base64.StdEncoding.EncodeToString(data)
}

// Helper to encode sort fields to base64
func encodeSort(sorts []query.SortField) string {
	data, _ := json.Marshal(sorts)
	return base64.StdEncoding.EncodeToString(data)
}

// ------------------------------------------
// TEST 1: BASIC PAGINATION WITHOUT FILTERS
// ------------------------------------------
func TestPaginationBasicNoFilters(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	seedUsers(t, db)
	p := query.NewPagination[User](false)

	// Query: pageIndex=0, pageSize=2
	ctx := createEchoContext("pageIndex=0&pageSize=2")
	result, err := p.Pagination(db, ctx.Request().Context(), ctx)
	assert.NoError(t, err)

	assert.Equal(t, 5, result.TotalSize)
	assert.Equal(t, 3, result.TotalPage)
	assert.Len(t, result.Data, 2)
}

// ------------------------------------------
// TEST 2: PAGINATION WITH STRUCTURED FILTER
// ------------------------------------------
func TestPaginationWithStructuredFilter(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 20, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 40, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 50, CreatedAt: base.Add(-6 * time.Hour)},
		{ID: uuid.New(), Name: "Eve", Age: 60, CreatedAt: base},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}
	p := query.NewPagination[User](false)

	// Filter: age >= 30
	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 30, Mode: query.ModeGTE, DataType: query.DataTypeNumber},
		},
		Logic: query.LogicAnd,
	}
	filterEncoded := encodeFilter(filter)

	ctx := createEchoContext("filter=" + filterEncoded + "&pageIndex=0&pageSize=10")
	result, err := p.Pagination(db, ctx.Request().Context(), ctx)
	assert.NoError(t, err)

	assert.Equal(t, 4, result.TotalSize) // Bob, Charlie, David, Eve
	assert.Len(t, result.Data, 4)
}

// ------------------------------------------
// TEST 3: PAGINATION WITH SORT FIELDS
// ------------------------------------------
func TestPaginationWithSort(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 20, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 40, CreatedAt: base.Add(-12 * time.Hour)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}
	p := query.NewPagination[User](false)

	// Sort by age ascending
	sorts := []query.SortField{
		{Field: "age", Order: query.SortOrderAsc},
	}
	sortEncoded := encodeSort(sorts)

	ctx := createEchoContext("sort=" + sortEncoded + "&pageIndex=0&pageSize=10")
	result, err := p.Pagination(db, ctx.Request().Context(), ctx)
	assert.NoError(t, err)

	assert.Equal(t, 3, result.TotalSize)
	assert.Len(t, result.Data, 3)
	assert.Equal(t, "Alice", result.Data[0].Name)   // Age 20
	assert.Equal(t, "Bob", result.Data[1].Name)     // Age 30
	assert.Equal(t, "Charlie", result.Data[2].Name) // Age 40
}

// ------------------------------------------
// TEST 4: PAGINATION PAGE 2 WITH QUERY PARAMS
// ------------------------------------------
func TestPaginationQueryPage2(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	seedList := []User{
		{ID: uuid.New(), Name: "Alice", Age: 20, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 40, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 50, CreatedAt: base.Add(-6 * time.Hour)},
		{ID: uuid.New(), Name: "Eve", Age: 60, CreatedAt: base},
	}
	if err := db.Create(&seedList).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}
	p := query.NewPagination[User](false)

	// Page 1 (second page), size 2
	ctx := createEchoContext("pageIndex=1&pageSize=2")
	result, err := p.Pagination(db, ctx.Request().Context(), ctx)
	assert.NoError(t, err)

	assert.Equal(t, 1, result.PageIndex)
	assert.Equal(t, 2, result.PageSize)
	assert.Equal(t, 5, result.TotalSize)
	assert.Equal(t, 3, result.TotalPage)
	assert.Len(t, result.Data, 2)
}

// ------------------------------------------
// TEST 5: COMPLEX FILTER + SORT + PAGINATION
// ------------------------------------------
func TestPaginationComplex(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	truncate := func(t time.Time) time.Time {
		return t.Truncate(time.Second)
	}
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 20, CreatedAt: truncate(base.Add(-48 * time.Hour))},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: truncate(base.Add(-24 * time.Hour))},
		{ID: uuid.New(), Name: "Charlie", Age: 40, CreatedAt: truncate(base.Add(-12 * time.Hour))},
		{ID: uuid.New(), Name: "David", Age: 50, CreatedAt: truncate(base.Add(-6 * time.Hour))},
		{ID: uuid.New(), Name: "Eve", Age: 60, CreatedAt: truncate(base)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}
	p := query.NewPagination[User](false)

	// Filter: age >= 30, Sort: age ASC
	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 30, Mode: query.ModeGTE, DataType: query.DataTypeNumber},
		},
		SortFields: []query.SortField{
			{Field: "age", Order: query.SortOrderAsc},
		},
		Logic: query.LogicAnd,
	}
	filterEncoded := encodeFilter(filter)

	ctx := createEchoContext("filter=" + filterEncoded + "&pageIndex=0&pageSize=2")
	result, err := p.Pagination(db, ctx.Request().Context(), ctx)
	assert.NoError(t, err)

	assert.Equal(t, 4, result.TotalSize) // Bob, Charlie, David, Eve
	assert.Equal(t, 2, result.TotalPage)
	assert.Len(t, result.Data, 2)
	// Default sort is created_at DESC, so newest first
	assert.Equal(t, "Eve", result.Data[0].Name)   // Newest
	assert.Equal(t, "David", result.Data[1].Name) // Second newest
}
