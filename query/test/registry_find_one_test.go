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

func TestRegistryFindOneVariants(t *testing.T) {
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
	res, err := r.FindOne(ctx, &User{Age: 30})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Bob", res.Name)
	resLock, err := r.FindOneWithLock(ctx, &User{Age: 30})
	assert.NoError(t, err)
	assert.NotNil(t, resLock)
	assert.Equal(t, "Bob", resLock.Name)
	filters := []query.ArrFilterSQL{{Field: "age", Op: query.ModeEqual, Value: 35}}
	resArr, err := r.ArrFindOne(ctx, filters, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resArr)
	assert.Equal(t, "Charlie", resArr.Name)
	resArrLock, err := r.ArrFindOneWithLock(ctx, db, filters, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resArrLock)
	assert.Equal(t, "Charlie", resArrLock.Name)
	structFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{{Field: "age", Value: 30, Mode: query.ModeEqual, DataType: query.DataTypeNumber}},
		Logic:        query.LogicAnd,
	}
	resStruct, err := r.StructuredFindOne(ctx, structFilter)
	assert.NoError(t, err)
	assert.NotNil(t, resStruct)
	assert.Equal(t, "Bob", resStruct.Name)
	resStructLock, err := r.StructuredFindOneWithLock(ctx, db, structFilter)
	assert.NoError(t, err)
	assert.NotNil(t, resStructLock)
	assert.Equal(t, "Bob", resStructLock.Name)
}
