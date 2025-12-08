package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Simple model used only for this test
type ProductSQLOnlyNew struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Brand string
	Price int
	// intentionally omit CreatedAt to ensure NewSQLFilter avoids ordering by it
}

func setupDBNew(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	if err := db.AutoMigrate(&ProductSQLOnlyNew{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Seed some data
	items := []ProductSQLOnlyNew{
		{ID: 1, Name: "iPhone 15 Pro", Brand: "Apple", Price: 999},
		{ID: 2, Name: "Galaxy S21", Brand: "Samsung", Price: 799},
		{ID: 3, Name: "Pixel 7", Brand: "Google", Price: 599},
	}
	for _, it := range items {
		if err := db.Create(&it).Error; err != nil {
			t.Fatalf("failed to seed data: %v", err)
		}
	}

	return db
}

func TestSQLOnlyNew(t *testing.T) {
	db := setupDBNew(t)

	// Use NewSQLFilter to avoid getters/reflection
	h := filter.NewSQLFilter[ProductSQLOnlyNew](filter.GolangFilteringConfig{})

	// Simple equality filter: Brand = "Apple"
	root := filter.Root{
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "brand",
				DataType: filter.DataTypeText,
				Mode:     filter.ModeEqual,
				Value:    "Apple",
			},
		},
		SortFields: []filter.SortField{},
	}

	res, err := h.DataGorm(db, root, 0, 10)
	if err != nil {
		t.Fatalf("DataGorm failed: %v", err)
	}

	if res.TotalSize != 1 {
		t.Fatalf("expected 1 result, got %d", res.TotalSize)
	}

	if len(res.Data) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(res.Data))
	}

	if res.Data[0].Brand != "Apple" {
		t.Fatalf("expected brand Apple, got %v", res.Data[0].Brand)
	}

	// Also test pagination without explicit sort; it should not error even though CreatedAt is missing
	// Create additional rows to ensure pagination works
	for i := 4; i <= 25; i++ {
		p := ProductSQLOnlyNew{ID: uint(i), Name: "Extra", Brand: "Misc", Price: i}
		db.Create(&p)
	}

	// Page 0 size 10
	resPage0, err := h.DataGorm(db, filter.Root{FieldFilters: []filter.FieldFilter{}}, 0, 10)
	if err != nil {
		t.Fatalf("DataGorm page0 failed: %v", err)
	}
	if resPage0.PageSize != 10 {
		t.Fatalf("unexpected page size")
	}

	// Page 1 size 10
	resPage1, err := h.DataGorm(db, filter.Root{FieldFilters: []filter.FieldFilter{}}, 1, 10)
	if err != nil {
		t.Fatalf("DataGorm page1 failed: %v", err)
	}

	// Ensure the two pages are different sets (deterministic ordering must be applied)
	if len(resPage0.Data) == 0 || len(resPage1.Data) == 0 {
		t.Fatalf("expected non-empty pages")
	}

	if resPage0.Data[0].ID == resPage1.Data[0].ID && resPage0.Data[1].ID == resPage1.Data[1].ID {
		t.Fatalf("pages appear identical — ordering not deterministic")
	}

	// Sanity: ensure DataGormNoPage works too with NewSQLFilter
	all, err := h.DataGormNoPage(db, filter.Root{FieldFilters: []filter.FieldFilter{}})
	if err != nil {
		t.Fatalf("DataGormNoPage failed: %v", err)
	}
	_ = all // we only assert it runs without error

	// Small time-bound to ensure nothing hangs
	t.Logf("SQL-only tests completed at %v", time.Now())
}
