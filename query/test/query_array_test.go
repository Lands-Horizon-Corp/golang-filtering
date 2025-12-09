package query_test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestArrPaginationBasic(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	users := seedUsers(t, db)
	p := query.NewPagination[User](true)

	// No filters — match all
	res, err := p.ArrPagination(db, nil, nil, 0, 2)
	assert.NoError(t, err)

	assert.Equal(t, 5, res.TotalSize)
	assert.Equal(t, 3, res.TotalPage)
	assert.Len(t, res.Data, 2)
	assert.Equal(t, users[4].Name, res.Data[0].Name)
	assert.Equal(t, users[3].Name, res.Data[1].Name)
}

func TestArrPaginationWithFilter(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
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
	p := query.NewPagination[User](false)

	filters := []query.ArrFilterSQL{{Field: "age", Op: query.ModeEqual, Value: 30}}

	res, err := p.ArrPagination(db, filters, nil, 0, 10)
	assert.NoError(t, err)

	assert.Equal(t, 1, res.TotalSize)
	assert.Len(t, res.Data, 1)
	assert.Equal(t, "Bob", res.Data[0].Name)
}
