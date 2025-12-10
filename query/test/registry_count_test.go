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

func TestRegistryCounts(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&User{})
	assert.NoError(t, err)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 22},
		{ID: uuid.New(), Name: "Bob", Age: 30},
		{ID: uuid.New(), Name: "Charlie", Age: 25},
	}
	for _, u := range users {
		err := db.Create(&u).Error
		assert.NoError(t, err)
	}

	r := registry.NewRegistry(registry.RegistryParams[User, User, any]{
		Database: db,
		Resource: func(d *User) *User { return d },
	})

	ctx := context.Background()
	filter := &User{Age: 25}
	count, err := r.Count(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	filters := []query.ArrFilterSQL{
		{Field: "age", Op: query.ModeGT, Value: 22},
	}
	arrCount, err := r.ArrCount(ctx, filters)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), arrCount)

	filterRoot := query.StructuredFilter{
		FieldFilters: []query.FieldFilter{
			{
				Field:    "age",
				Mode:     query.ModeGTE,
				Value:    25,
				DataType: query.DataTypeNumber,
			},
		},
		Logic: query.LogicAnd,
	}
	structCount, err := r.StructuredCount(ctx, db, filterRoot)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), structCount)

	rawDB := db.Model(&User{}).Where("age <= ?", 25)
	rawCount, err := r.RawCount(ctx, rawDB)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), rawCount)

	rawAll, err := r.RawCount(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), rawAll)
}
