package query_test

import (
	"context"
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/registry"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegistryGetByIDVariants(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&User{})
	assert.NoError(t, err)

	u := User{
		ID:   uuid.New(),
		Name: "Alice",
		Age:  22,
	}
	err = db.Create(&u).Error
	assert.NoError(t, err)

	r := registry.NewRegistry(registry.RegistryParams[User, User, any]{
		Database: db,
		Resource: func(d *User) *User { return d },
	})

	ctx := context.Background()

	res, err := r.GetByID(ctx, u.ID)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Alice", res.Name)

	nonID := uuid.New()
	res2, err := r.GetByID(ctx, nonID)
	assert.Error(t, err)
	assert.Nil(t, res2)

	res3, err := r.GetByIDIncludingDeleted(ctx, u.ID)
	assert.NoError(t, err)
	assert.NotNil(t, res3)
	assert.Equal(t, "Alice", res3.Name)

	nonDelID := uuid.New()
	res4, err := r.GetByIDIncludingDeleted(ctx, nonDelID)
	assert.Error(t, err)
	assert.Nil(t, res4)
}
