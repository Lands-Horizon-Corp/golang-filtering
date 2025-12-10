package query_test

import (
	"context"
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/Lands-Horizon-Corp/golang-filtering/registry"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegistryTabularVariants(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25},
		{ID: uuid.New(), Name: "Bob", Age: 30},
		{ID: uuid.New(), Name: "Charlie", Age: 35},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	r := registry.NewRegistry(registry.RegistryParams[User, User, any]{
		Database: db,
		Resource: func(d *User) *User { return d },
		Tabular: func(data *User) map[string]any {
			return map[string]any{
				"id":   data.ID,
				"name": data.Name,
				"age":  data.Age,
			}
		},
	})
	ctx := context.Background()
	normalFilter := User{Age: 30}
	normalCSV, err := r.Tabular(ctx, normalFilter)
	assert.NoError(t, err)
	assert.NotEmpty(t, normalCSV)
	csvStr := string(normalCSV)
	assert.Contains(t, csvStr, "Bob")
	assert.NotContains(t, csvStr, "Alice")
	assert.NotContains(t, csvStr, "Charlie")

	arrFilters := []query.ArrFilterSQL{{Field: "age", Op: query.ModeEqual, Value: 35}}
	arrCSV, err := r.ArrTabular(ctx, arrFilters, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, arrCSV)
	arrCSVStr := string(arrCSV)
	assert.Contains(t, arrCSVStr, "Charlie")
	assert.NotContains(t, arrCSVStr, "Alice")
	assert.NotContains(t, arrCSVStr, "Bob")

	structFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{Field: "age", Value: 30, Mode: query.ModeEqual, DataType: query.DataTypeNumber},
		},
		Logic: query.LogicAnd,
	}
	structCSV, err := r.StructuredTabular(ctx, structFilter)
	assert.NoError(t, err)
	assert.NotEmpty(t, structCSV)
	structCSVStr := string(structCSV)
	assert.Contains(t, structCSVStr, "Bob")
	assert.NotContains(t, structCSVStr, "Alice")
	assert.NotContains(t, structCSVStr, "Charlie")
}
