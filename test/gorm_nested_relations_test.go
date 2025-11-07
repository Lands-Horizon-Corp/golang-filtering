package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCurrency represents a currency entity (separate table)
type TestCurrency struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"type:varchar(255);not null" json:"name"`
	CurrencyCode string `gorm:"type:varchar(10);not null" json:"currency_code"`
	Symbol       string `gorm:"type:varchar(10)" json:"symbol"`
}

// TestBillAndCoin represents a bill or coin entity with foreign key to TestCurrency
type TestBillAndCoin struct {
	ID         uint          `gorm:"primaryKey" json:"id"`
	CreatedAt  time.Time     `gorm:"not null" json:"created_at"`
	CurrencyID uint          `gorm:"not null" json:"currency_id"`
	Currency   *TestCurrency `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
	Name       string        `gorm:"type:varchar(255)" json:"name"`
	Value      float64       `gorm:"type:decimal;not null" json:"value"`
}

func setupNestedRelationsDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schemas
	if err := db.AutoMigrate(&TestCurrency{}, &TestBillAndCoin{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create currencies
	currencies := []TestCurrency{
		{ID: 1, Name: "US Dollar", CurrencyCode: "USD", Symbol: "$"},
		{ID: 2, Name: "Euro", CurrencyCode: "EUR", Symbol: "€"},
		{ID: 3, Name: "Philippine Peso", CurrencyCode: "PHP", Symbol: "₱"},
		{ID: 4, Name: "Japanese Yen", CurrencyCode: "JPY", Symbol: "¥"},
	}
	for _, currency := range currencies {
		if err := db.Create(&currency).Error; err != nil {
			t.Fatalf("Failed to create currency: %v", err)
		}
	}

	// Create bills and coins
	billsAndCoins := []TestBillAndCoin{
		{ID: 1, CurrencyID: 1, Name: "One Dollar Bill", Value: 1.00, CreatedAt: time.Now()},
		{ID: 2, CurrencyID: 1, Name: "Five Dollar Bill", Value: 5.00, CreatedAt: time.Now()},
		{ID: 3, CurrencyID: 2, Name: "Euro Coin", Value: 1.00, CreatedAt: time.Now()},
		{ID: 4, CurrencyID: 3, Name: "Peso Coin", Value: 1.00, CreatedAt: time.Now()},
		{ID: 5, CurrencyID: 3, Name: "Twenty Peso Bill", Value: 20.00, CreatedAt: time.Now()},
		{ID: 6, CurrencyID: 4, Name: "Yen Coin", Value: 100.00, CreatedAt: time.Now()},
	}
	for _, item := range billsAndCoins {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("Failed to create bill/coin: %v", err)
		}
	}

	return db
}

// TestGormNestedRelationFilterEqual tests filtering on related table field (exact match)
func TestGormNestedRelationFilterEqual(t *testing.T) {
	db := setupNestedRelationsDB(t)
	handler := filter.NewFilter[TestBillAndCoin](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "currency.name",
				Value:    "US Dollar",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"Currency"}, // Preload Currency for the result
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by currency.name: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 items with US Dollar, got %d", result.TotalSize)
	}

	// Verify all results have Currency preloaded and name is "US Dollar"
	for _, item := range result.Data {
		if item.Currency == nil {
			t.Error("Currency was not preloaded")
			continue
		}
		if item.Currency.Name != "US Dollar" {
			t.Errorf("Expected currency name 'US Dollar', got '%s'", item.Currency.Name)
		}
	}
}

// TestGormNestedRelationFilterContains tests filtering with LIKE on related table
func TestGormNestedRelationFilterContains(t *testing.T) {
	db := setupNestedRelationsDB(t)
	handler := filter.NewFilter[TestBillAndCoin](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "currency.name",
				Value:    "peso",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"Currency"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by currency.name contains: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 items with 'peso' in currency name, got %d", result.TotalSize)
	}

	for _, item := range result.Data {
		if item.Currency == nil {
			t.Error("Currency was not preloaded")
			continue
		}
		t.Logf("Found: %s - %s (%s)", item.Name, item.Currency.Name, item.Currency.CurrencyCode)
	}
}

// TestGormNestedRelationFilterCurrencyCode tests filtering by currency code
func TestGormNestedRelationFilterCurrencyCode(t *testing.T) {
	db := setupNestedRelationsDB(t)
	handler := filter.NewFilter[TestBillAndCoin](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "currency.currency_code",
				Value:    "PHP",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"Currency"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter by currency.currency_code: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 PHP items, got %d", result.TotalSize)
	}

	for _, item := range result.Data {
		if item.Currency == nil {
			t.Error("Currency was not preloaded")
			continue
		}
		if item.Currency.CurrencyCode != "PHP" {
			t.Errorf("Expected PHP, got %s", item.Currency.CurrencyCode)
		}
	}
}

// TestGormNestedRelationMultipleFilters tests combining main table and related table filters
func TestGormNestedRelationMultipleFilters(t *testing.T) {
	db := setupNestedRelationsDB(t)
	handler := filter.NewFilter[TestBillAndCoin](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "currency.name",
				Value:    "Philippine Peso",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "value",
				Value:    10.0,
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeNumber,
			},
		},
		Preload: []string{"Currency"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter with multiple conditions: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 PHP item with value > 10, got %d", result.TotalSize)
	}

	if len(result.Data) > 0 {
		item := result.Data[0]
		if item.Name != "Twenty Peso Bill" {
			t.Errorf("Expected 'Twenty Peso Bill', got '%s'", item.Name)
		}
		if item.Value != 20.00 {
			t.Errorf("Expected value 20.00, got %.2f", item.Value)
		}
	}
}

// TestGormNestedRelationOrLogic tests OR logic with nested relations
func TestGormNestedRelationOrLogic(t *testing.T) {
	db := setupNestedRelationsDB(t)
	handler := filter.NewFilter[TestBillAndCoin](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicOr,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "currency.currency_code",
				Value:    "USD",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "currency.currency_code",
				Value:    "EUR",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		Preload: []string{"Currency"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter with OR logic: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected 3 items (USD or EUR), got %d", result.TotalSize)
	}

	for _, item := range result.Data {
		if item.Currency == nil {
			t.Error("Currency was not preloaded")
			continue
		}
		if item.Currency.CurrencyCode != "USD" && item.Currency.CurrencyCode != "EUR" {
			t.Errorf("Expected USD or EUR, got %s", item.Currency.CurrencyCode)
		}
	}
}

// TestGormNestedRelationSorting tests sorting by related table field
func TestGormNestedRelationSorting(t *testing.T) {
	db := setupNestedRelationsDB(t)
	handler := filter.NewFilter[TestBillAndCoin](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "currency.name", Order: filter.SortOrderAsc},
		},
		Preload: []string{"Currency"},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to sort by currency.name: %v", err)
	}

	if result.TotalSize != 6 {
		t.Errorf("Expected 6 total items, got %d", result.TotalSize)
	}

	// Verify sorted by currency name
	expectedOrder := []string{"Euro", "Japanese Yen", "Philippine Peso", "Philippine Peso", "US Dollar", "US Dollar"}
	for i, item := range result.Data {
		if item.Currency == nil {
			t.Error("Currency was not preloaded")
			continue
		}
		if item.Currency.Name != expectedOrder[i] {
			t.Errorf("Expected currency name '%s' at position %d, got '%s'", expectedOrder[i], i, item.Currency.Name)
		}
		t.Logf("Position %d: %s - %s", i, item.Name, item.Currency.Name)
	}
}

// TestGormNestedRelationWithoutPreload tests that join works even without Preload
func TestGormNestedRelationWithoutPreload(t *testing.T) {
	db := setupNestedRelationsDB(t)
	handler := filter.NewFilter[TestBillAndCoin](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "currency.name",
				Value:    "Japanese Yen",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		// Note: No Preload specified
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to filter without preload: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 Japanese Yen item, got %d", result.TotalSize)
	}

	// Currency should be nil because we didn't preload it
	if len(result.Data) > 0 {
		item := result.Data[0]
		if item.Currency != nil {
			t.Log("Note: Currency was loaded even without explicit Preload (GORM optimization)")
		} else {
			t.Log("Currency is nil as expected (no Preload specified)")
		}
		t.Logf("Found item: %s (value: %.2f)", item.Name, item.Value)
	}
}
