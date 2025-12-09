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

func TestRegistryUpdateByID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	u := User{ID: uuid.New(), Name: "Before", Age: 20}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	r := registry.NewRegistry(registry.RegistryParams[User, User, any]{
		Database: db,
		Resource: func(d *User) *User { return d },
	})

	fields := &User{Name: "After", Age: 21}
	if err := r.UpdateByID(context.Background(), u.ID, fields); err != nil {
		t.Fatalf("UpdateByID failed: %v", err)
	}

	var got User
	if err := db.First(&got, "id = ?", u.ID).Error; err != nil {
		t.Fatalf("failed to fetch updated user: %v", err)
	}
	assert.Equal(t, "After", got.Name)
	assert.Equal(t, 21, got.Age)
	assert.Equal(t, got.Name, fields.Name)
}

func TestRegistryUpdateByIDWithTx(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	u := User{ID: uuid.New(), Name: "BeforeTx", Age: 30}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	r := registry.NewRegistry(registry.RegistryParams[User, User, any]{
		Database: db,
		Resource: func(d *User) *User { return d },
	})

	tx := db.Begin()
	fields := &User{Name: "AfterTx", Age: 31}
	if err := r.UpdateByIDWithTx(context.Background(), tx, u.ID, fields); err != nil {
		tx.Rollback()
		t.Fatalf("UpdateByIDWithTx failed: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}

	var got User
	if err := db.First(&got, "id = ?", u.ID).Error; err != nil {
		t.Fatalf("failed to fetch updated user after tx: %v", err)
	}
	assert.Equal(t, "AfterTx", got.Name)
	assert.Equal(t, 31, got.Age)
	assert.Equal(t, got.Name, fields.Name)
}
