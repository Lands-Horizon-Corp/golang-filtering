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

func createEchoContext(queryStr string) echo.Context {
	req := httptest.NewRequest(http.MethodGet, "/?"+queryStr, nil)
	rec := httptest.NewRecorder()
	e := echo.New()
	return e.NewContext(req, rec)
}

func encodeFilter(filter query.StructuredFilter) string {
	data, _ := json.Marshal(filter)
	return base64.StdEncoding.EncodeToString(data)
}

func encodeSort(sorts []query.SortField) string {
	data, _ := json.Marshal(sorts)
	return base64.StdEncoding.EncodeToString(data)
}

func TestPaginationBasicNoFilters(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	seedUsers(t, db)
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Query: pageIndex=0, pageSize=2
	ctx := createEchoContext("pageIndex=0&pageSize=2")
	result, err := p.Pagination(db, ctx.Request().Context(), ctx)
	assert.NoError(t, err)

	assert.Equal(t, 5, result.TotalSize)
	assert.Equal(t, 3, result.TotalPage)
	assert.Len(t, result.Data, 2)
}

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
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

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

	assert.Equal(t, 4, result.TotalSize)
	assert.Len(t, result.Data, 4)
}

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
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

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
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

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
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

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

// TestPaginationStructuredNoRouteFilter tests PaginationStructured without route filter
func TestPaginationStructuredNoRouteFilter(t *testing.T) {
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

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Query params only: age >= 30
	queryFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     query.ModeGTE,
				DataType: query.DataTypeNumber,
			},
		},
	}

	filterEncoded := encodeFilter(queryFilter)
	ctx := createEchoContext("filter=" + filterEncoded + "&pageIndex=0&pageSize=10")

	result, err := p.PaginationStructured(db, ctx.Request().Context(), ctx, query.StructuredFilter{})
	assert.NoError(t, err)

	assert.Equal(t, 4, result.TotalSize) // Bob, Charlie, David, Eve
	assert.Len(t, result.Data, 4)
}

// TestPaginationStructuredWithRouteFilter tests PaginationStructured merging route filter with query filters
func TestPaginationStructuredWithRouteFilter(t *testing.T) {
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

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Query params: age >= 30
	queryFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     query.ModeGTE,
				DataType: query.DataTypeNumber,
			},
		},
	}

	// Route filter: name = "Bob"
	routeFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "name",
				Value:    "Bob",
				Mode:     query.ModeEqual,
				DataType: query.DataTypeText,
			},
		},
	}

	filterEncoded := encodeFilter(queryFilter)
	ctx := createEchoContext("filter=" + filterEncoded + "&pageIndex=0&pageSize=10")

	result, err := p.PaginationStructured(db, ctx.Request().Context(), ctx, routeFilter)
	assert.NoError(t, err)

	// Both conditions must be true (AND logic): age >= 30 AND name = "Bob" -> only Bob matches
	assert.Equal(t, 1, result.TotalSize)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "Bob", result.Data[0].Name)
	assert.Equal(t, 30, result.Data[0].Age)
}

// TestPaginationStructuredSortFieldOverride tests that route filter sort takes precedence
func TestPaginationStructuredSortFieldOverride(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 28, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 30, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 35, CreatedAt: base.Add(-6 * time.Hour)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Query has no sort (defaults to created_at DESC)
	queryFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    25,
				Mode:     query.ModeGTE,
				DataType: query.DataTypeNumber,
			},
		},
	}

	// Route filter specifies sort by age ASC
	routeFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{}, // No additional filters
		SortFields: []query.SortField{
			{
				Field: "age",
				Order: "ASC",
			},
		},
	}

	filterEncoded := encodeFilter(queryFilter)
	ctx := createEchoContext("filter=" + filterEncoded + "&pageIndex=0&pageSize=10")

	result, err := p.PaginationStructured(db, ctx.Request().Context(), ctx, routeFilter)
	assert.NoError(t, err)

	// Results should be sorted by age ASC: 25, 28, 30, 35
	assert.Len(t, result.Data, 4)
	assert.Equal(t, 25, result.Data[0].Age) // Alice
	assert.Equal(t, 28, result.Data[1].Age) // Bob
	assert.Equal(t, 30, result.Data[2].Age) // Charlie
	assert.Equal(t, 35, result.Data[3].Age) // David
}

// TestPaginationStructuredPreloadMerging tests that preloads are merged
func TestPaginationStructuredPreloadMerging(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	seedUsers(t, db)

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Route filter without additional filters, just with empty preloads
	routeFilter := query.StructuredFilter{
		Preload: []string{}, // Empty preloads
	}

	ctx := createEchoContext("pageIndex=0&pageSize=10")

	result, err := p.PaginationStructured(db, ctx.Request().Context(), ctx, routeFilter)
	assert.NoError(t, err)

	// Just verify the query executed without error and returned all users
	assert.Equal(t, 5, result.TotalSize)
	assert.Len(t, result.Data, 5)
}

// TestPaginationStructuredComplexMerge tests complex scenario with multiple filters and sorts
func TestPaginationStructuredComplexMerge(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 28, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 30, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 35, CreatedAt: base.Add(-6 * time.Hour)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Query params: age >= 28
	queryFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    28,
				Mode:     query.ModeGTE,
				DataType: query.DataTypeNumber,
			},
		},
	}

	// Route filter: name = "Charlie", with sort by name ASC
	routeFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "name",
				Value:    "Charlie",
				Mode:     query.ModeEqual,
				DataType: query.DataTypeText,
			},
		},
		SortFields: []query.SortField{
			{
				Field: "name",
				Order: "ASC",
			},
		},
	}

	filterEncoded := encodeFilter(queryFilter)
	ctx := createEchoContext("filter=" + filterEncoded + "&pageIndex=0&pageSize=10")

	result, err := p.PaginationStructured(db, ctx.Request().Context(), ctx, routeFilter)
	assert.NoError(t, err)

	// Conditions: (age >= 28) AND (name = 'Charlie')
	// Only Charlie matches: age 30 >= 28 and name = "Charlie"
	assert.Equal(t, 1, result.TotalSize)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "Charlie", result.Data[0].Name)
	assert.Equal(t, 30, result.Data[0].Age)
}

// ------------------------------------------
// PAGINATION ARRAY TESTS
// ------------------------------------------

// TestPaginationArrayNoFilters tests PaginationArray without array filters
func TestPaginationArrayNoFilters(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	seedUsers(t, db)

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Query: pageIndex=0, pageSize=2, no filters
	ctx := createEchoContext("pageIndex=0&pageSize=2")
	result, err := p.PaginationArray(db, ctx.Request().Context(), ctx, nil, nil)
	assert.NoError(t, err)

	assert.Equal(t, 5, result.TotalSize)
	assert.Equal(t, 3, result.TotalPage)
	assert.Len(t, result.Data, 2)
}

// TestPaginationArrayWithSingleFilter tests PaginationArray with one array filter
func TestPaginationArrayWithSingleFilter(t *testing.T) {
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

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Filter: name = "Bob" (use string for text filters)
	filters := []query.ArrFilterSQL{{Field: "name", Op: query.ModeEqual, Value: "Bob"}}

	ctx := createEchoContext("pageIndex=0&pageSize=10")
	result, err := p.PaginationArray(db, ctx.Request().Context(), ctx, filters, nil)
	assert.NoError(t, err)

	assert.Len(t, result.Data, 1)
	assert.Equal(t, "Bob", result.Data[0].Name)
	assert.Equal(t, 30, result.Data[0].Age)
}

// TestPaginationArrayWithMultipleFilters tests PaginationArray with multiple array filters (AND logic)
func TestPaginationArrayWithMultipleFilters(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 20, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 30, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 50, CreatedAt: base.Add(-6 * time.Hour)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Filters: name starts with "B" AND age = 30 (AND logic)
	filters := []query.ArrFilterSQL{
		{Field: "name", Op: query.ModeStartsWith, Value: "B"},
		{Field: "age", Op: query.ModeEqual, Value: 30},
	}

	ctx := createEchoContext("pageIndex=0&pageSize=10")
	result, err := p.PaginationArray(db, ctx.Request().Context(), ctx, filters, nil)
	assert.NoError(t, err)

	// Only Bob matches: name starts with "B" AND age = 30
	assert.Equal(t, 1, result.TotalSize)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "Bob", result.Data[0].Name)
}

// TestPaginationArrayWithSort tests PaginationArray with sort fields
func TestPaginationArrayWithSort(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 28, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 30, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 35, CreatedAt: base.Add(-6 * time.Hour)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// Sort by name ASC (overrides default created_at DESC)
	sorts := []query.ArrFilterSortSQL{
		{Field: "name", Order: "ASC"},
	}

	ctx := createEchoContext("pageIndex=0&pageSize=10")
	result, err := p.PaginationArray(db, ctx.Request().Context(), ctx, nil, sorts)
	assert.NoError(t, err)

	assert.Len(t, result.Data, 4)
	// Sorted by name ASC: Alice, Bob, Charlie, David
	assert.Equal(t, "Alice", result.Data[0].Name)
	assert.Equal(t, "Bob", result.Data[1].Name)
	assert.Equal(t, "Charlie", result.Data[2].Name)
	assert.Equal(t, "David", result.Data[3].Name)
}

func TestPaginationArrayWithFilterAndSort(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 28, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 30, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 35, CreatedAt: base.Add(-6 * time.Hour)},
		{ID: uuid.New(), Name: "Emma", Age: 40, CreatedAt: base},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	filters := []query.ArrFilterSQL{
		{Field: "name", Op: query.ModeStartsWith, Value: "C"},
	}

	sorts := []query.ArrFilterSortSQL{
		{Field: "name", Order: "ASC"},
	}

	ctx := createEchoContext("pageIndex=0&pageSize=10")
	result, err := p.PaginationArray(db, ctx.Request().Context(), ctx, filters, sorts)
	assert.NoError(t, err)

	assert.Len(t, result.Data, 1)
	assert.Equal(t, "Charlie", result.Data[0].Name)
}

func TestPaginationArrayWithQueryFilters(t *testing.T) {
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
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	queryFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    35,
				Mode:     query.ModeGTE,
				DataType: query.DataTypeNumber,
			},
		},
	}

	arrayFilters := []query.ArrFilterSQL{
		{Field: "name", Op: query.ModeStartsWith, Value: "D"},
	}

	filterEncoded := encodeFilter(queryFilter)
	ctx := createEchoContext("filter=" + filterEncoded + "&pageIndex=0&pageSize=10")
	result, err := p.PaginationArray(db, ctx.Request().Context(), ctx, arrayFilters, nil)
	assert.NoError(t, err)

	assert.Equal(t, 1, result.TotalSize)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "David", result.Data[0].Name)
	assert.Equal(t, 50, result.Data[0].Age)
}

func TestPaginationArrayPagination(t *testing.T) {
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
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	filters := []query.ArrFilterSQL{
		{Field: "age", Op: query.ModeGTE, Value: 20},
	}

	ctx1 := createEchoContext("pageIndex=0&pageSize=2")
	result1, err := p.PaginationArray(db, ctx1.Request().Context(), ctx1, filters, nil)
	assert.NoError(t, err)
	assert.Equal(t, 4, result1.TotalSize)
	assert.Equal(t, 2, result1.TotalPage)
	assert.Len(t, result1.Data, 2)

	ctx2 := createEchoContext("pageIndex=1&pageSize=2")
	result2, err := p.PaginationArray(db, ctx2.Request().Context(), ctx2, filters, nil)
	assert.NoError(t, err)
	assert.Equal(t, 4, result2.TotalSize)
	assert.Equal(t, 2, result2.TotalPage)
	assert.Len(t, result2.Data, 2)

	assert.NotEqual(t, result1.Data[0].ID, result2.Data[0].ID)
}

func TestPaginationNormalNoFilter(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	seedUsers(t, db)

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	ctx := createEchoContext("pageIndex=0&pageSize=2")
	result, err := p.PaginationNormal(db, ctx.Request().Context(), ctx, nil)
	assert.NoError(t, err)

	assert.Equal(t, 5, result.TotalSize)
	assert.Equal(t, 3, result.TotalPage)
	assert.Len(t, result.Data, 2)
}

func TestPaginationNormalWithModelFilter(t *testing.T) {
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

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	filter := &User{Age: 30}

	ctx := createEchoContext("pageIndex=0&pageSize=10")
	result, err := p.PaginationNormal(db, ctx.Request().Context(), ctx, filter)
	assert.NoError(t, err)

	assert.Len(t, result.Data, 1)
	assert.Equal(t, "Bob", result.Data[0].Name)
	assert.Equal(t, 30, result.Data[0].Age)
}

func TestPaginationNormalWithQueryFilter(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 20, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 30, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 50, CreatedAt: base.Add(-6 * time.Hour)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	queryFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     query.ModeGTE,
				DataType: query.DataTypeNumber,
			},
		},
	}

	modelFilter := &User{Name: "Bob"}

	filterEncoded := encodeFilter(queryFilter)
	ctx := createEchoContext("filter=" + filterEncoded + "&pageIndex=0&pageSize=10")
	result, err := p.PaginationNormal(db, ctx.Request().Context(), ctx, modelFilter)
	assert.NoError(t, err)

	assert.Equal(t, 1, result.TotalSize)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "Bob", result.Data[0].Name)
	assert.Equal(t, 30, result.Data[0].Age)
}

func TestPaginationNormalWithQuerySort(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 28, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 30, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 35, CreatedAt: base.Add(-6 * time.Hour)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	sorts := []query.SortField{
		{Field: "age", Order: "ASC"},
	}
	sortEncoded := encodeSort(sorts)

	ctx := createEchoContext("sort=" + sortEncoded + "&pageIndex=0&pageSize=10")
	result, err := p.PaginationNormal(db, ctx.Request().Context(), ctx, nil)
	assert.NoError(t, err)

	assert.Len(t, result.Data, 4)
	assert.Equal(t, 25, result.Data[0].Age)
	assert.Equal(t, 28, result.Data[1].Age)
	assert.Equal(t, 30, result.Data[2].Age)
	assert.Equal(t, 35, result.Data[3].Age)
}

func TestPaginationNormalPagination(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 30, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 30, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 30, CreatedAt: base.Add(-6 * time.Hour)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})
	filter := &User{Age: 30}

	ctx1 := createEchoContext("pageIndex=0&pageSize=2")
	result1, err := p.PaginationNormal(db, ctx1.Request().Context(), ctx1, filter)
	assert.NoError(t, err)
	assert.Equal(t, 4, result1.TotalSize)
	assert.Equal(t, 2, result1.TotalPage)
	assert.Len(t, result1.Data, 2)

	ctx2 := createEchoContext("pageIndex=1&pageSize=2")
	result2, err := p.PaginationNormal(db, ctx2.Request().Context(), ctx2, filter)
	assert.NoError(t, err)
	assert.Equal(t, 4, result2.TotalSize)
	assert.Equal(t, 2, result2.TotalPage)
	assert.Len(t, result2.Data, 2)

	assert.NotEqual(t, result1.Data[0].ID, result2.Data[0].ID)
}

func TestPaginationNormalComplexMerge(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 30, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 35, CreatedAt: base.Add(-6 * time.Hour)},
		{ID: uuid.New(), Name: "Eve", Age: 40, CreatedAt: base},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	// URL query filter: age >= 30
	queryFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     query.ModeGTE,
				DataType: query.DataTypeNumber,
			},
		},
	}

	modelFilter := &User{Name: "Bob"}

	sorts := []query.SortField{
		{Field: "age", Order: "DESC"},
	}

	filterEncoded := encodeFilter(queryFilter)
	sortEncoded := encodeSort(sorts)
	ctx := createEchoContext("filter=" + filterEncoded + "&sort=" + sortEncoded + "&pageIndex=0&pageSize=10")
	result, err := p.PaginationNormal(db, ctx.Request().Context(), ctx, modelFilter)
	assert.NoError(t, err)

	assert.Equal(t, 1, result.TotalSize)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "Bob", result.Data[0].Name)
	assert.Equal(t, 30, result.Data[0].Age)
}

func TestPaginationNormalNoModelFilterWithQueryFilter(t *testing.T) {
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
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})
	queryFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    35,
				Mode:     query.ModeGTE,
				DataType: query.DataTypeNumber,
			},
		},
	}

	filterEncoded := encodeFilter(queryFilter)
	ctx := createEchoContext("filter=" + filterEncoded + "&pageIndex=0&pageSize=10")
	result, err := p.PaginationNormal(db, ctx.Request().Context(), ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, result.TotalSize)
	assert.Len(t, result.Data, 2)
}
