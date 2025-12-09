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

func TestRegistryMinMaxVariants(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 20},
		{ID: uuid.New(), Name: "Bob", Age: 30},
		{ID: uuid.New(), Name: "Charlie", Age: 40},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	r := registry.NewRegistry(registry.RegistryParams[User, User, any]{
		Database: db,
		Resource: func(d *User) *User { return d },
	})

	ctx := context.Background()

	// Normal GetMax/GetMin
	maxVal, err := r.GetMax(ctx, "age", &User{})
	assert.NoError(t, err)
	assert.NotNil(t, maxVal)
	assert.EqualValues(t, 40, maxVal)

	minVal, err := r.GetMin(ctx, "age", &User{})
	assert.NoError(t, err)
	assert.NotNil(t, minVal)
	assert.EqualValues(t, 20, minVal)

	// Normal lock variants
	maxLock, err := r.GetMaxLock(ctx, "age", &User{})
	assert.NoError(t, err)
	assert.EqualValues(t, 40, maxLock)

	minLock, err := r.GetMinLock(ctx, "age", &User{})
	assert.NoError(t, err)
	assert.EqualValues(t, 20, minLock)

	// Structured filters: age >= 30 -> min 30, max 40
	structFilter := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{{Field: "age", Value: 30, Mode: query.ModeGTE, DataType: query.DataTypeNumber}},
		Logic:        query.LogicAnd,
	}

	sMax, err := r.StructuredGetMax(ctx, "age", structFilter)
	assert.NoError(t, err)
	assert.EqualValues(t, 40, sMax)

	sMin, err := r.StructuredGetMin(ctx, "age", structFilter)
	assert.NoError(t, err)
	assert.EqualValues(t, 30, sMin)

	// Structured lock variants
	sMaxLock, err := r.StructuredGetMaxLock(ctx, "age", structFilter)
	assert.NoError(t, err)
	assert.EqualValues(t, 40, sMaxLock)

	sMinLock, err := r.StructuredGetMinLock(ctx, "age", structFilter)
	assert.NoError(t, err)
	assert.EqualValues(t, 30, sMinLock)
}
