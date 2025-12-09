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

func TestRegistryCreateVariants(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	r := registry.NewRegistry(registry.RegistryParams[User, User, any]{
		Database: db,
		Resource: func(d *User) *User { return d },
	})

	u1 := User{ID: uuid.New(), Name: "CreateOne", Age: 21}
	err = r.Create(context.Background(), &u1)
	assert.NoError(t, err)
	var got User
	if err := db.First(&got, "id = ?", u1.ID).Error; err != nil {
		t.Fatalf("record not found after Create: %v", err)
	}
	assert.Equal(t, "CreateOne", got.Name)

	u2 := User{ID: uuid.New(), Name: "CreateTx", Age: 22}
	tx := db.Begin()
	if err := r.CreateWithTx(context.Background(), tx, &u2); err != nil {
		tx.Rollback()
		t.Fatalf("CreateWithTx failed: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}
	got = User{}
	if err := db.First(&got, "id = ?", u2.ID).Error; err != nil {
		t.Fatalf("record not found after CreateWithTx: %v", err)
	}
	assert.Equal(t, "CreateTx", got.Name)

	u3 := &User{ID: uuid.New(), Name: "Many1", Age: 23}
	u4 := &User{ID: uuid.New(), Name: "Many2", Age: 24}
	if err := r.CreateMany(context.Background(), []*User{u3, u4}); err != nil {
		t.Fatalf("CreateMany failed: %v", err)
	}
	var count int64
	if err := db.Model(&User{}).Where("name IN (?)", []string{"Many1", "Many2"}).Count(&count).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	assert.Equal(t, int64(2), count)

	u5 := &User{ID: uuid.New(), Name: "ManyTx1", Age: 25}
	u6 := &User{ID: uuid.New(), Name: "ManyTx2", Age: 26}
	tx2 := db.Begin()
	if err := r.CreateManyWithTx(context.Background(), tx2, []*User{u5, u6}); err != nil {
		tx2.Rollback()
		t.Fatalf("CreateManyWithTx failed: %v", err)
	}
	if err := tx2.Commit().Error; err != nil {
		t.Fatalf("failed to commit tx2: %v", err)
	}
	if err := db.Model(&User{}).Where("name IN (?)", []string{"ManyTx1", "ManyTx2"}).Count(&count).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	assert.Equal(t, int64(2), count)
}
