package query_test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNormalCount(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	// seed
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25},
		{ID: uuid.New(), Name: "Bob", Age: 30},
		{ID: uuid.New(), Name: "Charlie", Age: 30},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed users: %v", err)
	}

	p := query.Pagination[User]{}
	count, err := p.NormalCount(db, User{Age: 30})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestArrCount(t *testing.T) {
	db, err := database(&User{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	// seed
	users := []User{
		{ID: uuid.New(), Name: "Alice", Age: 25},
		{ID: uuid.New(), Name: "Bob", Age: 30},
		{ID: uuid.New(), Name: "Charlie", Age: 30},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("failed to seed users: %v", err)
	}

	p := query.Pagination[User]{}
	filters := []query.ArrFilterSQL{
		{Field: "age", Op: query.ModeEqual, Value: 30},
	}
	count, err := p.ArrCount(db, filters)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
