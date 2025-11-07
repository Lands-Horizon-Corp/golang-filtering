package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestAmbiguousColumnWithORLogic tests the scenario where both direct and nested fields
// are used with OR logic, which previously caused "ambiguous column" errors
func TestAmbiguousColumnWithORLogic(t *testing.T) {
	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&TestCurrency{}, &TestBillAndCoin{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create test data
	currencies := []TestCurrency{
		{ID: 1, Name: "Philippine Peso", CurrencyCode: "PHP", Symbol: "₱"},
		{ID: 2, Name: "US Dollar", CurrencyCode: "USD", Symbol: "$"},
	}
	db.Create(&currencies)

	bills := []TestBillAndCoin{
		{ID: 1, Name: "PHP Coin", Value: 1.0, CurrencyID: 1, CreatedAt: time.Now()},
		{ID: 2, Name: "Dollar Bill", Value: 1.0, CurrencyID: 2, CreatedAt: time.Now()},
		{ID: 3, Name: "PHP Bill", Value: 20.0, CurrencyID: 1, CreatedAt: time.Now()},
	}
	db.Create(&bills)

	// Test the problematic scenario: OR logic with both direct and nested fields
	handler := filter.NewFilter[TestBillAndCoin](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicOr,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name", // Direct field
				Value:    "PHP",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "currency.currency_code", // Nested field
				Value:    "PHP",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering with OR logic (direct + nested): %v", err)
	}

	// Should return records where name='PHP' OR currency_code='PHP'
	// Expected: PHP Coin (matches both), PHP Bill (matches both) = 2 records
	if result.TotalSize != 2 {
		t.Errorf("Expected 2 records, got %d", result.TotalSize)
	}

	// Verify the results
	foundNames := make(map[string]bool)
	for _, bill := range result.Data {
		foundNames[bill.Name] = true
	}

	expectedNames := []string{"PHP Coin", "PHP Bill"}
	for _, name := range expectedNames {
		if !foundNames[name] {
			t.Errorf("Expected to find '%s' in results", name)
		}
	}
}

// TestAmbiguousColumnWithANDLogic tests AND logic with mixed fields
func TestAmbiguousColumnWithANDLogic(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&TestCurrency{}, &TestBillAndCoin{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	currencies := []TestCurrency{
		{ID: 1, Name: "Philippine Peso", CurrencyCode: "PHP", Symbol: "₱"},
		{ID: 2, Name: "US Dollar", CurrencyCode: "USD", Symbol: "$"},
	}
	db.Create(&currencies)

	bills := []TestBillAndCoin{
		{ID: 1, Name: "Small PHP Coin", Value: 1.0, CurrencyID: 1, CreatedAt: time.Now()},
		{ID: 2, Name: "Large PHP Coin", Value: 10.0, CurrencyID: 1, CreatedAt: time.Now()},
		{ID: 3, Name: "Small Dollar Coin", Value: 1.0, CurrencyID: 2, CreatedAt: time.Now()},
	}
	db.Create(&bills)

	handler := filter.NewFilter[TestBillAndCoin](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "Small",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "currency.currency_code",
				Value:    "PHP",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering with AND logic (direct + nested): %v", err)
	}

	// Should return only "Small PHP Coin"
	if result.TotalSize != 1 {
		t.Errorf("Expected 1 record, got %d", result.TotalSize)
	}

	if len(result.Data) > 0 && result.Data[0].Name != "Small PHP Coin" {
		t.Errorf("Expected 'Small PHP Coin', got '%s'", result.Data[0].Name)
	}
}

// TestAmbiguousColumnSorting tests sorting with mixed fields
func TestAmbiguousColumnSorting(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&TestCurrency{}, &TestBillAndCoin{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	currencies := []TestCurrency{
		{ID: 1, Name: "Philippine Peso", CurrencyCode: "PHP", Symbol: "₱"},
		{ID: 2, Name: "US Dollar", CurrencyCode: "USD", Symbol: "$"},
	}
	db.Create(&currencies)

	bills := []TestBillAndCoin{
		{ID: 1, Name: "Coin A", Value: 5.0, CurrencyID: 1, CreatedAt: time.Now()},
		{ID: 2, Name: "Coin B", Value: 1.0, CurrencyID: 2, CreatedAt: time.Now()},
		{ID: 3, Name: "Coin C", Value: 10.0, CurrencyID: 1, CreatedAt: time.Now()},
	}
	db.Create(&bills)

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
		SortFields: []filter.SortField{
			{Field: "value", Order: filter.SortOrderDesc}, // Sort by direct field
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error sorting with nested filter: %v", err)
	}

	// Should return PHP coins sorted by value descending
	if result.TotalSize != 2 {
		t.Errorf("Expected 2 records, got %d", result.TotalSize)
	}

	// Verify sort order: Coin C (10.0), Coin A (5.0)
	if len(result.Data) >= 2 {
		if result.Data[0].Name != "Coin C" {
			t.Errorf("Expected first record to be 'Coin C', got '%s'", result.Data[0].Name)
		}
		if result.Data[1].Name != "Coin A" {
			t.Errorf("Expected second record to be 'Coin A', got '%s'", result.Data[1].Name)
		}
	}
}
