package query_test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDBNoPaginationStr(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25},
		{ID: uuid.New(), Name: "Bob", Age: 30},
		{ID: uuid.New(), Name: "Charlie", Age: 35},
	}
	for _, u := range users {
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("failed to seed data: %v", err)
		}
	}
	return db
}

func TestNoPaginationStr_WithEncodedFilter(t *testing.T) {
	db := setupDBNoPaginationStr(t)
	p := &query.Pagination[User]{}

	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 30, Mode: query.ModeEqual, DataType: query.DataTypeNumber},
		},
	}
	encoded := encodeFilter(filter)

	result, err := p.NoPaginationStr(db, encoded)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Bob", result[0].Name)
}

func TestNoPaginationStructuredStr_WithEncodedFilter(t *testing.T) {
	db := setupDBNoPaginationStr(t)
	p := &query.Pagination[User]{}

	filter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 25, Mode: query.ModeEqual, DataType: query.DataTypeNumber},
		},
	}
	encoded := encodeFilter(query.StructuredFilter{})

	result, err := p.NoPaginationStructuredStr(db, encoded, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Alice", result[0].Name)
}

func TestNoPaginationArrayStr_WithEncodedFilter(t *testing.T) {
	db := setupDBNoPaginationStr(t)
	p := &query.Pagination[User]{}

	filters := []query.ArrFilterSQL{
		{Field: "age", Value: 35, Op: query.ModeEqual},
	}
	sorts := []query.ArrFilterSortSQL{
		{Field: "name", Order: query.SortOrderDesc},
	}

	result, err := p.NoPaginationArrayStr(db, "", filters, sorts)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Charlie", result[0].Name)
}

func TestNoPaginationNormalStr_WithFilter(t *testing.T) {
	db := setupDBNoPaginationStr(t)
	p := &query.Pagination[User]{}

	filter := &User{Age: 30}
	result, err := p.NoPaginationNormalStr(db, "", filter)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Bob", result[0].Name)
}
