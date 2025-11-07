package test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AccountTag represents preset filter conditions
type AccountTag struct {
	OrganizationID uint `gorm:"column:organization_id"`
	BranchID       uint `gorm:"column:branch_id"`
}

// Transaction model for testing preset struct conditions
type Transaction struct {
	ID             uint    `gorm:"primarykey" json:"id"`
	OrganizationID uint    `json:"organization_id"`
	BranchID       uint    `json:"branch_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Status         string  `json:"status"`
}

func setupTransactionDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&Transaction{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Insert test data for multiple organizations and branches
	transactions := []Transaction{
		// Org 1, Branch 1
		{ID: 1, OrganizationID: 1, BranchID: 1, Amount: 100.00, Currency: "USD", Status: "completed"},
		{ID: 2, OrganizationID: 1, BranchID: 1, Amount: 200.00, Currency: "USD", Status: "pending"},
		{ID: 3, OrganizationID: 1, BranchID: 1, Amount: 300.00, Currency: "EUR", Status: "completed"},

		// Org 1, Branch 2
		{ID: 4, OrganizationID: 1, BranchID: 2, Amount: 150.00, Currency: "USD", Status: "completed"},
		{ID: 5, OrganizationID: 1, BranchID: 2, Amount: 250.00, Currency: "EUR", Status: "pending"},

		// Org 2, Branch 1
		{ID: 6, OrganizationID: 2, BranchID: 1, Amount: 400.00, Currency: "USD", Status: "completed"},
		{ID: 7, OrganizationID: 2, BranchID: 1, Amount: 500.00, Currency: "GBP", Status: "completed"},

		// Org 2, Branch 2
		{ID: 8, OrganizationID: 2, BranchID: 2, Amount: 600.00, Currency: "USD", Status: "failed"},
	}

	db.Create(&transactions)
	return db
}

// TestApplyPresetConditionsHelper tests the helper function
func TestApplyPresetConditionsHelper(t *testing.T) {
	db := setupTransactionDB(t)
	handler := filter.NewFilter[Transaction](filter.GolangFilteringConfig{})

	// Create preset conditions struct
	tag := &AccountTag{
		OrganizationID: 1,
		BranchID:       1,
	}

	// Apply preset conditions using helper
	db = filter.ApplyPresetConditions(db, tag)

	// Apply additional filters
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "currency",
				Value:    "USD",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should return transactions: ID 1, 2 (Org 1, Branch 1, USD)
	if result.TotalSize != 2 {
		t.Errorf("Expected 2 transactions, got %d", result.TotalSize)
	}

	// Verify correct transactions
	for _, tx := range result.Data {
		if tx.OrganizationID != 1 || tx.BranchID != 1 || tx.Currency != "USD" {
			t.Errorf("Unexpected transaction: %+v", tx)
		}
	}
}

// TestDataGormWithPreset tests the convenience method
func TestDataGormWithPreset(t *testing.T) {
	db := setupTransactionDB(t)
	handler := filter.NewFilter[Transaction](filter.GolangFilteringConfig{})

	// Create preset conditions struct
	tag := &AccountTag{
		OrganizationID: 2,
		BranchID:       1,
	}

	// Additional filter for completed status
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "status",
				Value:    "completed",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	// Use convenience method
	result, err := handler.DataGormWithPreset(db, tag, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should return transactions: ID 6, 7 (Org 2, Branch 1, completed)
	if result.TotalSize != 2 {
		t.Errorf("Expected 2 transactions, got %d", result.TotalSize)
	}

	// Verify correct transactions
	for _, tx := range result.Data {
		if tx.OrganizationID != 2 || tx.BranchID != 1 || tx.Status != "completed" {
			t.Errorf("Unexpected transaction: %+v", tx)
		}
	}
}

// TestDataGormWithPresetNil tests nil preset conditions
func TestDataGormWithPresetNil(t *testing.T) {
	db := setupTransactionDB(t)
	handler := filter.NewFilter[Transaction](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "currency",
				Value:    "EUR",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	// Pass nil preset - should work like regular DataGorm
	result, err := handler.DataGormWithPreset(db, nil, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should return all EUR transactions across all orgs/branches
	if result.TotalSize != 2 {
		t.Errorf("Expected 2 EUR transactions, got %d", result.TotalSize)
	}
}

// TestMultiplePresetFields tests struct with multiple conditions
func TestMultiplePresetFields(t *testing.T) {
	db := setupTransactionDB(t)
	handler := filter.NewFilter[Transaction](filter.GolangFilteringConfig{})

	// Test with different org/branch combinations
	testCases := []struct {
		name     string
		tag      *AccountTag
		expected int
	}{
		{
			name:     "Org 1, Branch 1",
			tag:      &AccountTag{OrganizationID: 1, BranchID: 1},
			expected: 3,
		},
		{
			name:     "Org 1, Branch 2",
			tag:      &AccountTag{OrganizationID: 1, BranchID: 2},
			expected: 2,
		},
		{
			name:     "Org 2, Branch 1",
			tag:      &AccountTag{OrganizationID: 2, BranchID: 1},
			expected: 2,
		},
		{
			name:     "Org 2, Branch 2",
			tag:      &AccountTag{OrganizationID: 2, BranchID: 2},
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic:        filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{},
			}

			result, err := handler.DataGormWithPreset(db, tc.tag, filterRoot, 1, 10)
			if err != nil {
				t.Fatalf("Error filtering: %v", err)
			}

			if result.TotalSize != tc.expected {
				t.Errorf("Expected %d transactions for %s, got %d",
					tc.expected, tc.name, result.TotalSize)
			}
		})
	}
}

// TestPresetWithRangeFilter tests preset struct with range filters
func TestPresetWithRangeFilter(t *testing.T) {
	db := setupTransactionDB(t)
	handler := filter.NewFilter[Transaction](filter.GolangFilteringConfig{})

	tag := &AccountTag{
		OrganizationID: 1,
		BranchID:       1,
	}

	// Filter for amounts between 100 and 250
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "amount",
				Value: filter.Range{
					From: 100.0,
					To:   250.0,
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataGormWithPreset(db, tag, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should return ID 1 (100), ID 2 (200) from Org 1, Branch 1
	if result.TotalSize != 2 {
		t.Errorf("Expected 2 transactions, got %d", result.TotalSize)
	}

	// Verify amounts are in range
	for _, tx := range result.Data {
		if tx.Amount < 100 || tx.Amount > 250 {
			t.Errorf("Transaction amount %f out of range", tx.Amount)
		}
	}
}

// TestPresetWithSortingAndPagination tests preset with sorting and pagination
func TestPresetWithSortingAndPagination(t *testing.T) {
	db := setupTransactionDB(t)
	handler := filter.NewFilter[Transaction](filter.GolangFilteringConfig{})

	tag := &AccountTag{
		OrganizationID: 1,
		BranchID:       1,
	}

	filterRoot := filter.Root{
		Logic:        filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{},
		SortFields: []filter.SortField{
			{Field: "amount", Order: filter.SortOrderDesc},
		},
	}

	// Get first page with 2 items
	result, err := handler.DataGormWithPreset(db, tag, filterRoot, 1, 2)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	if result.TotalSize != 3 {
		t.Errorf("Expected total 3 transactions, got %d", result.TotalSize)
	}

	if len(result.Data) != 2 {
		t.Errorf("Expected 2 transactions on page 1, got %d", len(result.Data))
	}

	// Verify descending order
	if len(result.Data) == 2 && result.Data[0].Amount < result.Data[1].Amount {
		t.Error("Transactions not sorted in descending order")
	}

	// Highest amount should be 300
	if len(result.Data) > 0 && result.Data[0].Amount != 300.00 {
		t.Errorf("Expected highest amount 300, got %f", result.Data[0].Amount)
	}
}
