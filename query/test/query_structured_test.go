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
	return db.Model(new(T)), nil
}

type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `json:"name"`
	Age       int            `json:"age"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func TestStructuredPagination(t *testing.T) {
	db, err := database(&User{})
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
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})
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

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

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

	result, err := p.StructuredPagination(db, filter, 0, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 0, result.PageIndex)
	assert.Equal(t, 2, result.PageSize)
	assert.Equal(t, 4, result.TotalSize)
	assert.Equal(t, 2, result.TotalPage)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, "Charlie", result.Data[0].Name)
	assert.Equal(t, "David", result.Data[1].Name)

	resultPage2, err := p.StructuredPagination(db, filter, 1, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 1, resultPage2.PageIndex)
	assert.Len(t, resultPage2.Data, 2)
	assert.Equal(t, "Eve", resultPage2.Data[0].Name)
	assert.Equal(t, "Frank", resultPage2.Data[1].Name)
}

func TestStructuredPaginationComplexAdvanced(t *testing.T) {
	db, err := database(User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	base := time.Date(2025, 12, 9, 0, 0, 0, 0, time.UTC)
	truncate := func(t time.Time) time.Time {
		return t.Truncate(time.Second)
	}

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

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})
	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 30, Mode: query.ModeGTE, DataType: query.DataTypeNumber},
			{Field: "created_at", Value: truncate(base.Add(-5 * time.Hour)), Mode: query.ModeGTE, DataType: query.DataTypeDate},
		},
		SortFields: []query.SortField{
			{Field: "age", Order: query.SortOrderDesc},
			{Field: "created_at", Order: query.SortOrderAsc},
			{Field: "name", Order: query.SortOrderAsc},
		},
		Logic: query.LogicAnd,
	}

	resultPage0, err := p.StructuredPagination(db, filter, 0, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 0, resultPage0.PageIndex)
	assert.Len(t, resultPage0.Data, 2)
	assert.Equal(t, "Grace", resultPage0.Data[0].Name)
	assert.Equal(t, "Frank", resultPage0.Data[1].Name)

	resultPage1, err := p.StructuredPagination(db, filter, 1, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 1, resultPage1.PageIndex)
	assert.Len(t, resultPage1.Data, 2)
	assert.Equal(t, "Eve", resultPage1.Data[0].Name)
	assert.Equal(t, "David", resultPage1.Data[1].Name)

	resultPage2, err := p.StructuredPagination(db, filter, 2, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}

	assert.Equal(t, 2, resultPage2.PageIndex)
	assert.Len(t, resultPage2.Data, 1)
	assert.Equal(t, "Charlie", resultPage2.Data[0].Name)
	assert.Equal(t, 5, resultPage2.TotalSize)
	assert.Equal(t, 3, resultPage2.TotalPage)
}

func TestStructuredPaginationWithRange(t *testing.T) {
	db, err := database(User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	base := time.Date(2025, 12, 9, 0, 0, 0, 0, time.UTC)
	truncate := func(t time.Time) time.Time { return t.Truncate(time.Second) }

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

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

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

	resultPage0, err := p.StructuredPagination(db, filter, 0, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}
	assert.Equal(t, 0, resultPage0.PageIndex)
	assert.Len(t, resultPage0.Data, 2)
	assert.Equal(t, "Charlie", resultPage0.Data[0].Name)
	assert.Equal(t, "David", resultPage0.Data[1].Name)

	resultPage1, err := p.StructuredPagination(db, filter, 1, 2)
	if err != nil {
		t.Fatalf("pagination failed: %v", err)
	}
	assert.Equal(t, 1, resultPage1.PageIndex)
	assert.Len(t, resultPage1.Data, 2)
	assert.Equal(t, "Eve", resultPage1.Data[0].Name)
	assert.Equal(t, "Frank", resultPage1.Data[1].Name)

	assert.Equal(t, 4, resultPage1.TotalSize)
	assert.Equal(t, 2, resultPage1.TotalPage)
}

func TestAllModes(t *testing.T) {
	db, err := database(User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	base := time.Date(2025, 12, 9, 0, 0, 0, 0, time.UTC)
	truncate := func(t time.Time) time.Time { return t.Truncate(time.Second) }

	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25, CreatedAt: truncate(base.Add(-5 * time.Hour))},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: truncate(base.Add(-4 * time.Hour))},
		{ID: uuid.New(), Name: "Charlie", Age: 35, CreatedAt: truncate(base.Add(-3 * time.Hour))},
		{ID: uuid.New(), Name: "David", Age: 40, CreatedAt: truncate(base.Add(-2 * time.Hour))},
		{ID: uuid.New(), Name: "Eve", Age: 45, CreatedAt: truncate(base.Add(-1 * time.Hour))},
		{ID: uuid.New(), Name: "", Age: 50, CreatedAt: truncate(base)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to insert sample data: %v", err)
	}

	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	tests := []struct {
		name          string
		fieldFilter   query.FieldFilter
		expectedNames []string
	}{
		{"ModeEqual", query.FieldFilter{Field: "name", Value: "Alice", Mode: query.ModeEqual, DataType: query.DataTypeText}, []string{"Alice"}},
		{"ModeNotEqual", query.FieldFilter{Field: "name", Value: "Alice", Mode: query.ModeNotEqual, DataType: query.DataTypeText}, []string{"Bob", "Charlie", "David", "Eve", ""}},
		{"ModeContains", query.FieldFilter{Field: "name", Value: "li", Mode: query.ModeContains, DataType: query.DataTypeText}, []string{"Alice", "Charlie"}},
		{"ModeNotContains", query.FieldFilter{Field: "name", Value: "li", Mode: query.ModeNotContains, DataType: query.DataTypeText}, []string{"Bob", "David", "Eve", ""}},
		{"ModeStartsWith", query.FieldFilter{Field: "name", Value: "A", Mode: query.ModeStartsWith, DataType: query.DataTypeText}, []string{"Alice"}},
		{"ModeEndsWith", query.FieldFilter{Field: "name", Value: "e", Mode: query.ModeEndsWith, DataType: query.DataTypeText}, []string{"Alice", "Charlie", "Eve"}},

		{"ModeGT", query.FieldFilter{Field: "age", Value: 35, Mode: query.ModeGT, DataType: query.DataTypeNumber}, []string{"David", "Eve", ""}},
		{"ModeGTE", query.FieldFilter{Field: "age", Value: 35, Mode: query.ModeGTE, DataType: query.DataTypeNumber}, []string{"Charlie", "David", "Eve", ""}},
		{"ModeLT", query.FieldFilter{Field: "age", Value: 35, Mode: query.ModeLT, DataType: query.DataTypeNumber}, []string{"Alice", "Bob"}},
		{"ModeLTE", query.FieldFilter{Field: "age", Value: 35, Mode: query.ModeLTE, DataType: query.DataTypeNumber}, []string{"Alice", "Bob", "Charlie"}},
		{"ModeRange", query.FieldFilter{Field: "age", Value: query.RangeNumber{From: 30, To: 40}, Mode: query.ModeRange, DataType: query.DataTypeNumber}, []string{"Bob", "Charlie", "David"}},
		{"ModeInside", query.FieldFilter{Field: "age", Value: query.RangeNumber{From: 30, To: 40}, Mode: query.ModeInside, DataType: query.DataTypeNumber}, []string{"Bob", "Charlie", "David"}},
		{"ModeOutside", query.FieldFilter{Field: "age", Value: query.RangeNumber{From: 30, To: 40}, Mode: query.ModeOutside, DataType: query.DataTypeNumber}, []string{"Alice", "Eve", ""}},

		{"ModeBefore", query.FieldFilter{Field: "created_at", Value: truncate(base.Add(-2 * time.Hour)), Mode: query.ModeBefore, DataType: query.DataTypeDate}, []string{"Alice", "Bob", "Charlie"}},
		{"ModeAfter", query.FieldFilter{Field: "created_at", Value: truncate(base.Add(-3 * time.Hour)), Mode: query.ModeAfter, DataType: query.DataTypeDate}, []string{"Charlie", "David", "Eve", ""}},

		{"ModeIsEmpty", query.FieldFilter{Field: "name", Mode: query.ModeIsEmpty, DataType: query.DataTypeText}, []string{""}},
		{"ModeIsNotEmpty", query.FieldFilter{Field: "name", Mode: query.ModeIsNotEmpty, DataType: query.DataTypeText}, []string{"Alice", "Bob", "Charlie", "David", "Eve"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := query.StructuredFilter{
				FieldFilters: []query.FieldFilter{tt.fieldFilter},
				Logic:        query.LogicAnd,
			}
			result, err := p.StructuredPagination(db, filter, 0, 10)
			if err != nil {
				t.Fatalf("pagination failed: %v", err)
			}

			assert.Equal(t, len(tt.expectedNames), result.TotalSize)
			names := make([]string, 0, len(result.Data))
			for _, u := range result.Data {
				names = append(names, u.Name)
			}
			assert.ElementsMatch(t, tt.expectedNames, names)
		})
	}
}
