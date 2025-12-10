package query_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
)

func setupDBNoPagination(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	db.Create(&User{ID: uuid.New(), Name: "Alice", Age: 30})
	db.Create(&User{ID: uuid.New(), Name: "Bob", Age: 25})
	db.Create(&User{ID: uuid.New(), Name: "Charlie", Age: 30})
	return db
}

func setupEchoContext() echo.Context {
	e := echo.New()
	// Add dummy pageIndex and pageSize to satisfy parseQuery
	req := httptest.NewRequest(http.MethodGet, "/?pageIndex=1&pageSize=10", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	return ctx
}

func TestNoPaginationMethods(t *testing.T) {
	db := setupDBNoPagination(t)
	ctx := setupEchoContext()
	p := &query.Pagination[User]{}

	t.Run("NoPagination Basic", func(t *testing.T) {
		users, err := p.NoPagination(db, ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 1)
	})

	t.Run("NoPaginationStructured", func(t *testing.T) {
		filter := query.StructuredFilter{
			FieldFilters: []query.FieldFilter{
				{Field: "age", Value: 30, Mode: query.ModeEqual, DataType: query.DataTypeNumber},
			},
		}
		users, err := p.NoPaginationStructured(db, context.Background(), ctx, filter)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 1)
		for _, u := range users {
			assert.Equal(t, 30, u.Age)
		}
	})

	t.Run("NoPaginationArray", func(t *testing.T) {
		filters := []query.ArrFilterSQL{
			{Field: "age", Value: 30, Op: query.ModeEqual},
		}
		users, err := p.NoPaginationArray(db, context.Background(), ctx, filters, nil)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 1)
		for _, u := range users {
			assert.Equal(t, 30, u.Age)
		}
	})

	t.Run("NoPaginationNormal", func(t *testing.T) {
		filter := &User{Age: 30}
		users, err := p.NoPaginationNormal(db, context.Background(), ctx, filter)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 1)
		for _, u := range users {
			assert.Equal(t, 30, u.Age)
		}
	})
}
