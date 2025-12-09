package query_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
)

func seedUsers(t *testing.T, db *gorm.DB) []User {
	base := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)

	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 20, CreatedAt: base.Add(-48 * time.Hour)},
		{ID: uuid.New(), Name: "Bob", Age: 30, CreatedAt: base.Add(-24 * time.Hour)},
		{ID: uuid.New(), Name: "Charlie", Age: 40, CreatedAt: base.Add(-12 * time.Hour)},
		{ID: uuid.New(), Name: "David", Age: 50, CreatedAt: base.Add(-6 * time.Hour)},
		{ID: uuid.New(), Name: "Eve", Age: 60, CreatedAt: base},
	}

	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed seed: %v", err)
	}

	return users
}

func TestPaginationBasic(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	users := seedUsers(t, db)
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	filter := &User{} // No filter — match all

	res, err := p.NormalPagination(db, filter, 0, 2)
	assert.NoError(t, err)

	assert.Equal(t, 5, res.TotalSize)
	assert.Equal(t, 3, res.TotalPage)
	assert.Len(t, res.Data, 2)

	assert.Equal(t, users[0].Name, res.Data[0].Name)
	assert.Equal(t, users[1].Name, res.Data[1].Name)
}
func TestPaginationPage2(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	users := seedUsers(t, db)
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	res, err := p.NormalPagination(db.Order("created_at ASC"), &User{}, 1, 2)
	assert.NoError(t, err)

	assert.Len(t, res.Data, 2)
	assert.Equal(t, users[2].Name, res.Data[0].Name)
	assert.Equal(t, users[3].Name, res.Data[1].Name)
}

func TestPaginationLastPage(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	users := seedUsers(t, db)
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	res, err := p.NormalPagination(db.Order("created_at ASC"), &User{}, 2, 2)
	assert.NoError(t, err)

	assert.Len(t, res.Data, 1)
	assert.Equal(t, users[4].Name, res.Data[0].Name)
}

func TestPaginationWithFilter(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	seedUsers(t, db)
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	res, err := p.NormalPagination(db.Order("created_at ASC"), &User{Age: 30}, 0, 10)
	assert.NoError(t, err)

	assert.Equal(t, 1, res.TotalSize)
	assert.Len(t, res.Data, 1)
	assert.Equal(t, "Bob", res.Data[0].Name)
}

func TestPaginationPageSizeValidation(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	seedUsers(t, db)
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	res, err := p.NormalPagination(db, &User{}, 0, -1)
	assert.NoError(t, err)
	assert.Equal(t, 30, res.PageSize)
}

func TestPaginationPageIndexValidation(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	seedUsers(t, db)
	p := query.NewPagination[User](query.PaginationConfig{
		Verbose: true,
	})

	res, err := p.NormalPagination(db, &User{}, -10, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, res.PageIndex)
}
