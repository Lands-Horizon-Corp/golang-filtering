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

func TestRegistryFindVariants(t *testing.T) {
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
	})

	ctx := context.Background()

	res, err := r.Find(ctx, &User{Age: 30})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 1)
	assert.Equal(t, "Bob", res[0].Name)

	arrFilters := []query.ArrFilterSQL{{Field: "age", Op: query.ModeEqual, Value: 35}}
	resArr, err := r.ArrFind(ctx, arrFilters, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resArr)
	assert.Len(t, resArr, 1)
	assert.Equal(t, "Charlie", resArr[0].Name)

	structFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{{Field: "age", Value: 30, Mode: query.ModeEqual, DataType: query.DataTypeNumber}},
		Logic:        query.LogicAnd,
	}
	resStruct, err := r.StructuredFind(ctx, structFilter)
	assert.NoError(t, err)
	assert.NotNil(t, resStruct)
	assert.Len(t, resStruct, 1)
	assert.Equal(t, "Bob", resStruct[0].Name)
}
