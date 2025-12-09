package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/registry"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegistryNormalPaginationSimple(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 20, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 40, CreatedAt: base.Add(-12 * time.Hour)},
	}

	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	r := registry.NewRegistry(registry.RegistryParams[User, User, any]{
		Database: db,
		Resource: func(d *User) *User { return d },
	})

	ctx := createEchoContext("pageIndex=0&pageSize=10")
	res, err := r.NormalPagination(context.Background(), ctx, &User{Age: 30})
	assert.NoError(t, err)
	assert.Equal(t, 1, res.TotalSize)
	assert.Len(t, res.Data, 1)
	assert.Equal(t, "Bob", res.Data[0].Name)
}
