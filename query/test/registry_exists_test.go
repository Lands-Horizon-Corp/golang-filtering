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

func TestRegistryExistsVariants(t *testing.T) {
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
		{ID: uuid.New(), Name: "Charlie", Age: 30},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	r := registry.NewRegistry(registry.RegistryParams[User, User, any]{
		Database: db,
		Resource: func(d *User) *User { return d },
	})

	ctx := context.Background()

	ok, err := r.Exists(ctx, &User{Age: 30})
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = r.ExistsByID(ctx, users[1].ID)
	assert.NoError(t, err)
	assert.True(t, ok)

	filters := []query.ArrFilterSQL{{Field: "age", Op: query.ModeEqual, Value: 30}}

	ok, err = r.ExistsIncludingDeleted(ctx, filters)
	assert.NoError(t, err)
	assert.True(t, ok)
}
