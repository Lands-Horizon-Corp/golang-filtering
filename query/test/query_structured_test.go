package query_test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func database[T any](model T) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(model); err != nil {
		return nil, err
	}
	return db, nil
}

// Example model
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

func TestStructuredPagination(t *testing.T) {
	db, err := database[any](&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Charlie", Age: 35, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "David", Age: 40, CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Eve", Age: 45, CreatedAt: time.Now()},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to insert sample data: %v", err)
	}
	p := query.NewPagination[User](true)
	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     query.ModeGTE,
				DataType: query.DataTypeNumber,
			},
		},
		SortFields: []query.SortField{
			{Field: "age", Order: query.SortOrderAsc},
		},
		Logic: query.LogicAnd,
	}
	result, err := p.StructuredPagination(db, filter, 0, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}
	assert.Equal(t, 0, result.PageIndex)
	assert.Equal(t, 2, result.PageSize)
	assert.Equal(t, 4, result.TotalSize)
	assert.Equal(t, 2, result.TotalPage)
	assert.Len(t, result.Data, 2)

	assert.Equal(t, "Bob", result.Data[0].Name)
	assert.Equal(t, 30, result.Data[0].Age)

	assert.Equal(t, "Charlie", result.Data[1].Name)
	assert.Equal(t, 35, result.Data[1].Age)

	users = []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: time.Now().Add(-5 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: time.Now().Add(-4 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 35, CreatedAt: time.Now().Add(-3 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 40, CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: uuid.New(), Name: "Eve", Age: 45, CreatedAt: time.Now().Add(-1 * time.Hour)},
		{ID: uuid.New(), Name: "Frank", Age: 50, CreatedAt: time.Now()},
	}
}

func TestStructuredPaginationComplex(t *testing.T) {
	db, err := database(User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	truncate := func(t time.Time) time.Time {
		return t.Truncate(time.Second)
	}

	// Deterministic timestamps
	base := truncate(time.Date(2025, 12, 9, 0, 0, 0, 0, time.UTC))

	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: truncate(base.Add(-5 * time.Hour))},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: truncate(base.Add(-4 * time.Hour))},
		{ID: uuid.New(), Name: "Charlie", Age: 35, CreatedAt: truncate(base.Add(-3 * time.Hour))},
		{ID: uuid.New(), Name: "David", Age: 40, CreatedAt: truncate(base.Add(-2 * time.Hour))},
		{ID: uuid.New(), Name: "Eve", Age: 45, CreatedAt: truncate(base.Add(-1 * time.Hour))},
		{ID: uuid.New(), Name: "Frank", Age: 50, CreatedAt: truncate(base)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to insert sample data: %v", err)
	}

	p := query.NewPagination[User](true)

	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 30, Mode: query.ModeGTE, DataType: query.DataTypeNumber},
			{Field: "created_at", Value: base.Add(-3 * time.Hour), Mode: query.ModeGTE, DataType: query.DataTypeDate},
		},
		SortFields: []query.SortField{
			{Field: "age", Order: query.SortOrderAsc},
			{Field: "created_at", Order: query.SortOrderAsc},
			{Field: "id", Order: query.SortOrderAsc},
		},
		Logic: query.LogicAnd,
	}

	// Page 0, size 2
	result, err := p.StructuredPagination(db, filter, 0, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 0, result.PageIndex)
	assert.Equal(t, 2, result.PageSize)
	assert.Equal(t, 4, result.TotalSize) // Charlie, David, Eve, Frank
	assert.Equal(t, 2, result.TotalPage)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "Charlie", result.Data[0].Name)
	assert.Equal(t, "David", result.Data[1].Name)

	// Page 1
	resultPage2, err := p.StructuredPagination(db, filter, 1, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 1, resultPage2.PageIndex)
	assert.Len(t, resultPage2.Data, 2) // Eve, Frank
	assert.Equal(t, "Eve", resultPage2.Data[0].Name)
	assert.Equal(t, "Frank", resultPage2.Data[1].Name)
}

func TestStructuredPaginationComplexAdvanced(t *testing.T) {
	db, err := database(User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Deterministic timestamps
	base := time.Date(2025, 12, 9, 0, 0, 0, 0, time.UTC)
	truncate := func(t time.Time) time.Time {
		return t.Truncate(time.Second)
	}

	// Create 7 users with varied ages and timestamps
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: truncate(base.Add(-7 * time.Hour))},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: truncate(base.Add(-6 * time.Hour))},
		{ID: uuid.New(), Name: "Charlie", Age: 35, CreatedAt: truncate(base.Add(-5 * time.Hour))},
		{ID: uuid.New(), Name: "David", Age: 40, CreatedAt: truncate(base.Add(-4 * time.Hour))},
		{ID: uuid.New(), Name: "Eve", Age: 45, CreatedAt: truncate(base.Add(-3 * time.Hour))},
		{ID: uuid.New(), Name: "Frank", Age: 50, CreatedAt: truncate(base.Add(-2 * time.Hour))},
		{ID: uuid.New(), Name: "Grace", Age: 55, CreatedAt: truncate(base.Add(-1 * time.Hour))},
	}

	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to insert sample data: %v", err)
	}

	p := query.NewPagination[User](true)

	// Complex filter: age >= 30 AND created_at >= base-5h
	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 30, Mode: query.ModeGTE, DataType: query.DataTypeNumber},
			{Field: "created_at", Value: truncate(base.Add(-5 * time.Hour)), Mode: query.ModeGTE, DataType: query.DataTypeDate},
		},
		SortFields: []query.SortField{
			{Field: "age", Order: query.SortOrderDesc},       // sort by age descending
			{Field: "created_at", Order: query.SortOrderAsc}, // then by created_at ascending
			{Field: "name", Order: query.SortOrderAsc},       // then by name ascending
		},
		Logic: query.LogicAnd,
	}

	// Page size 2 → total matching: Charlie, David, Eve, Frank, Grace = 5
	// This ensures last page will have odd number (1 item)
	// Page 0
	resultPage0, err := p.StructuredPagination(db, filter, 0, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 0, resultPage0.PageIndex)
	assert.Len(t, resultPage0.Data, 2)
	assert.Equal(t, "Grace", resultPage0.Data[0].Name) // highest age
	assert.Equal(t, "Frank", resultPage0.Data[1].Name) // next highest

	// Page 1
	resultPage1, err := p.StructuredPagination(db, filter, 1, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 1, resultPage1.PageIndex)
	assert.Len(t, resultPage1.Data, 2)
	assert.Equal(t, "Eve", resultPage1.Data[0].Name)
	assert.Equal(t, "David", resultPage1.Data[1].Name)

	// Page 2 (last page, odd item)
	resultPage2, err := p.StructuredPagination(db, filter, 2, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 2, resultPage2.PageIndex)
	assert.Len(t, resultPage2.Data, 1)
	assert.Equal(t, "Charlie", resultPage2.Data[0].Name)
	assert.Equal(t, 5, resultPage2.TotalSize)
	assert.Equal(t, 3, resultPage2.TotalPage) // 5 items / page size 2 → 3 pages
}

func TestStructuredPaginationWithRange(t *testing.T) {
	db, err := database(User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Deterministic timestamps
	base := time.Date(2025, 12, 9, 0, 0, 0, 0, time.UTC)
	truncate := func(t time.Time) time.Time { return t.Truncate(time.Second) }

	// Create 7 users with varied ages
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: truncate(base.Add(-7 * time.Hour))},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: truncate(base.Add(-6 * time.Hour))},
		{ID: uuid.New(), Name: "Charlie", Age: 35, CreatedAt: truncate(base.Add(-5 * time.Hour))},
		{ID: uuid.New(), Name: "David", Age: 40, CreatedAt: truncate(base.Add(-4 * time.Hour))},
		{ID: uuid.New(), Name: "Eve", Age: 45, CreatedAt: truncate(base.Add(-3 * time.Hour))},
		{ID: uuid.New(), Name: "Frank", Age: 50, CreatedAt: truncate(base.Add(-2 * time.Hour))},
		{ID: uuid.New(), Name: "Grace", Age: 55, CreatedAt: truncate(base.Add(-1 * time.Hour))},
	}

	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to insert sample data: %v", err)
	}

	p := query.NewPagination[User](true)

	// Range filter: age between 35 and 50 inclusive
	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Value:    query.RangeNumber{From: 35, To: 50},
				Mode:     query.ModeRange,
				DataType: query.DataTypeNumber,
			},
		},
		SortFields: []query.SortField{
			{Field: "age", Order: query.SortOrderAsc},
			{Field: "created_at", Order: query.SortOrderDesc},
		},
		Logic: query.LogicAnd,
	}

	// Page size 2 → 4 matching users: Charlie(35), David(40), Eve(45), Frank(50)
	resultPage0, err := p.StructuredPagination(db, filter, 0, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}
	assert.Equal(t, 0, resultPage0.PageIndex)
	assert.Len(t, resultPage0.Data, 2)
	assert.Equal(t, "Charlie", resultPage0.Data[0].Name)
	assert.Equal(t, "David", resultPage0.Data[1].Name)

	// Page 1
	resultPage1, err := p.StructuredPagination(db, filter, 1, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}
	assert.Equal(t, 1, resultPage1.PageIndex)
	assert.Len(t, resultPage1.Data, 2)
	assert.Equal(t, "Eve", resultPage1.Data[0].Name)
	assert.Equal(t, "Frank", resultPage1.Data[1].Name)

	// Total size & pages
	assert.Equal(t, 4, resultPage1.TotalSize)
	assert.Equal(t, 2, resultPage1.TotalPage)
}
