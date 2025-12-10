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

func TestRegistryRawFindOneVariants(t *testing.T) {
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

	// Create a query with model and order
	queryAll := db.Model(&User{}).Order("age ASC")
	res, err := r.RawFindOne(ctx, queryAll)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Alice", res.Name)

	queryCharlie := db.Model(&User{}).Where("name = ?", "Charlie")
	resFiltered, err := r.RawFindOne(ctx, queryCharlie)
	assert.NoError(t, err)
	assert.NotNil(t, resFiltered)
	assert.Equal(t, "Charlie", resFiltered.Name)

	queryLock := db.Model(&User{}).Order("age ASC")
	resLock, err := r.RawFindOneWithLock(ctx, queryLock)
	assert.NoError(t, err)
	assert.NotNil(t, resLock)
	assert.Equal(t, "Alice", resLock.Name)

	queryBobLock := db.Model(&User{}).Where("name = ?", "Bob")
	resLockFiltered, err := r.RawFindOneWithLock(ctx, queryBobLock)
	assert.NoError(t, err)
	assert.NotNil(t, resLockFiltered)
	assert.Equal(t, "Bob", resLockFiltered.Name)
}
