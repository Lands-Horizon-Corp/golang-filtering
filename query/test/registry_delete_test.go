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

func TestRegistryDeleteVariants(t *testing.T) {
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
	if err := r.Delete(ctx, users[1].ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	var got User
	if err := db.First(&got, "id = ?", users[1].ID).Error; err == nil {
		t.Fatalf("expected record to be deleted, but found: %v", got)
	}
	txUser := User{ID: uuid.New(), Name: "TxDelete", Age: 40}
	if err := db.Create(&txUser).Error; err != nil {
		t.Fatalf("failed to seed txUser: %v", err)
	}
	tx := db.Begin()
	if err := r.DeleteWithTx(ctx, tx, txUser.ID); err != nil {
		tx.Rollback()
		t.Fatalf("DeleteWithTx failed: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("failed to commit tx: %v", err)
	}
	got = User{}
	if err := db.First(&got, "id = ?", txUser.ID).Error; err == nil {
		t.Fatalf("expected txUser to be deleted, but found: %v", got)
	}
	b1 := User{ID: uuid.New(), Name: "Bulk1", Age: 50}
	b2 := User{ID: uuid.New(), Name: "Bulk2", Age: 51}
	if err := db.Create(&[]User{b1, b2}).Error; err != nil {
		t.Fatalf("failed to seed bulk users: %v", err)
	}
	ids := []any{b1.ID, b2.ID}
	if err := r.BulkDelete(ctx, ids); err != nil {
		t.Fatalf("BulkDelete failed: %v", err)
	}
	var count int64
	if err := db.Model(&User{}).Where("id IN (?)", ids).Count(&count).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	assert.Equal(t, int64(0), count)
	c1 := User{ID: uuid.New(), Name: "BulkTx1", Age: 60}
	c2 := User{ID: uuid.New(), Name: "BulkTx2", Age: 61}
	if err := db.Create(&[]User{c1, c2}).Error; err != nil {
		t.Fatalf("failed to seed bulk tx users: %v", err)
	}
	ids2 := []any{c1.ID, c2.ID}
	tx2 := db.Begin()
	if err := r.BulkDeleteWithTx(ctx, tx2, ids2); err != nil {
		tx2.Rollback()
		t.Fatalf("BulkDeleteWithTx failed: %v", err)
	}
	if err := tx2.Commit().Error; err != nil {
		t.Fatalf("failed to commit tx2: %v", err)
	}
	if err := db.Model(&User{}).Where("id IN (?)", ids2).Count(&count).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	assert.Equal(t, int64(0), count)
}

func TestRegistryRawFindIncludeDeletedVariants(t *testing.T) {
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

	if err := r.Delete(ctx, users[1].ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	results, err := r.RawFindIncludeDeleted(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(results))

	filter := db.Where("age > ?", 28)
	resultsFiltered, err := r.RawFindIncludeDeleted(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resultsFiltered))

	resultsLock, err := r.RawFindLockIncludeDeleted(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(resultsLock))

	resultsLockFiltered, err := r.RawFindLockIncludeDeleted(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resultsLockFiltered))

	rawResults, err := r.RawFindIncludeDeletedRaw(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(rawResults))
	assert.IsType(t, &User{}, rawResults[0])

	rawResultsFiltered, err := r.RawFindIncludeDeletedRaw(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(rawResultsFiltered))

	rawResultsLock, err := r.RawFindLockIncludeDeletedRaw(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(rawResultsLock))

	rawResultsLockFiltered, err := r.RawFindLockIncludeDeletedRaw(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(rawResultsLockFiltered))
}
