package test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// BillAndCoin represents a multi-tenant financial record
type BillAndCoin struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	OrganizationID uint    `json:"organization_id"`
	BranchID       uint    `json:"branch_id"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Status         string  `json:"status"`
}

// TestDataGorm_PresetConditions tests DataGorm with existing WHERE clauses
func TestDataGorm_PresetConditions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&BillAndCoin{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create test data for multiple organizations and branches
	records := []BillAndCoin{
		// Organization 1, Branch 1
		{OrganizationID: 1, BranchID: 1, Amount: 100.0, Currency: "USD", Status: "active"},
		{OrganizationID: 1, BranchID: 1, Amount: 200.0, Currency: "USD", Status: "pending"},
		{OrganizationID: 1, BranchID: 1, Amount: 300.0, Currency: "EUR", Status: "active"},
		// Organization 1, Branch 2
		{OrganizationID: 1, BranchID: 2, Amount: 150.0, Currency: "USD", Status: "active"},
		{OrganizationID: 1, BranchID: 2, Amount: 250.0, Currency: "USD", Status: "inactive"},
		// Organization 2, Branch 1
		{OrganizationID: 2, BranchID: 1, Amount: 400.0, Currency: "USD", Status: "active"},
		{OrganizationID: 2, BranchID: 1, Amount: 500.0, Currency: "GBP", Status: "active"},
		// Organization 2, Branch 2
		{OrganizationID: 2, BranchID: 2, Amount: 600.0, Currency: "USD", Status: "pending"},
	}

	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create test records: %v", err)
	}

	handler := filter.NewFilter[BillAndCoin]()

	// Test 1: Preset conditions only (no filterRoot filters)
	t.Run("PresetConditionsOnly", func(t *testing.T) {
		// User wants all records for org=1, branch=1
		presetDB := db.Where("organization_id = ? AND branch_id = ?", 1, 1)

		filterRoot := filter.Root{
			Logic:        filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{}, // No additional filters
		}

		pageIndex := 1
		pageSize := 10

		result, err := handler.DataGorm(presetDB, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("Error filtering: %v", err)
		}

		// Should get 3 records for org=1, branch=1
		if result.TotalSize != 3 {
			t.Errorf("Expected 3 records for org=1 branch=1, got %d", result.TotalSize)
		}

		// Verify all returned records match the preset conditions
		for _, record := range result.Data {
			if record.OrganizationID != 1 || record.BranchID != 1 {
				t.Errorf("Record doesn't match preset conditions: org=%d, branch=%d",
					record.OrganizationID, record.BranchID)
			}
		}
	})

	// Test 2: Preset conditions + filterRoot filters
	t.Run("PresetConditionsWithFilters", func(t *testing.T) {
		// User wants org=1, branch=1, with status=active and amount>150
		presetDB := db.Where("organization_id = ? AND branch_id = ?", 1, 1)

		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "status",
					Value:    "active",
					Mode:     filter.ModeEqual,
					DataType: filter.DataTypeText,
				},
				{
					Field:    "amount",
					Value:    150,
					Mode:     filter.ModeGT,
					DataType: filter.DataTypeNumber,
				},
			},
		}

		pageIndex := 1
		pageSize := 10

		result, err := handler.DataGorm(presetDB, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("Error filtering: %v", err)
		}

		// Should get 1 record: org=1, branch=1, status=active, amount=300
		if result.TotalSize != 1 {
			t.Errorf("Expected 1 record matching all conditions, got %d", result.TotalSize)
		}

		if len(result.Data) > 0 {
			record := result.Data[0]
			if record.OrganizationID != 1 || record.BranchID != 1 {
				t.Errorf("Record doesn't match preset conditions")
			}
			if record.Status != "active" {
				t.Errorf("Expected status=active, got %s", record.Status)
			}
			if record.Amount <= 150 {
				t.Errorf("Expected amount>150, got %f", record.Amount)
			}
		}
	})

	// Test 3: Multiple tenants with currency filter
	t.Run("MultiTenantWithCurrencyFilter", func(t *testing.T) {
		// User wants org=2 only, filter by USD currency
		presetDB := db.Where("organization_id = ?", 2)

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

		pageIndex := 1
		pageSize := 10

		result, err := handler.DataGorm(presetDB, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("Error filtering: %v", err)
		}

		// Should get 2 records: org=2 with USD currency
		if result.TotalSize != 2 {
			t.Errorf("Expected 2 USD records for org=2, got %d", result.TotalSize)
		}

		for _, record := range result.Data {
			if record.OrganizationID != 2 {
				t.Errorf("Expected org=2, got org=%d", record.OrganizationID)
			}
			if record.Currency != "USD" {
				t.Errorf("Expected USD, got %s", record.Currency)
			}
		}
	})

	// Test 4: Preset + Sorting
	t.Run("PresetWithSorting", func(t *testing.T) {
		presetDB := db.Where("organization_id = ? AND branch_id = ?", 1, 1)

		filterRoot := filter.Root{
			Logic:        filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{},
			SortFields: []filter.SortField{
				{Field: "amount", Order: filter.SortOrderDesc},
			},
		}

		pageIndex := 1
		pageSize := 10

		result, err := handler.DataGorm(presetDB, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("Error filtering: %v", err)
		}

		if result.TotalSize != 3 {
			t.Errorf("Expected 3 records, got %d", result.TotalSize)
		}

		// Verify descending order by amount
		if len(result.Data) >= 2 {
			for i := 0; i < len(result.Data)-1; i++ {
				if result.Data[i].Amount < result.Data[i+1].Amount {
					t.Errorf("Results not sorted by amount DESC: %f < %f",
						result.Data[i].Amount, result.Data[i+1].Amount)
				}
			}
		}
	})

	// Test 5: No preset conditions (backward compatibility)
	t.Run("NoPresetConditions", func(t *testing.T) {
		// Pass db without any WHERE clauses - should work as before
		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "status",
					Value:    "active",
					Mode:     filter.ModeEqual,
					DataType: filter.DataTypeText,
				},
			},
		}

		pageIndex := 1
		pageSize := 10

		result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("Error filtering: %v", err)
		}

		// Should get all active records across all orgs/branches
		if result.TotalSize != 5 {
			t.Errorf("Expected 5 active records total, got %d", result.TotalSize)
		}
	})

	// Test 6: OR logic with preset conditions
	t.Run("PresetWithORLogic", func(t *testing.T) {
		presetDB := db.Where("organization_id = ?", 1)

		filterRoot := filter.Root{
			Logic: filter.LogicOr,
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "currency",
					Value:    "EUR",
					Mode:     filter.ModeEqual,
					DataType: filter.DataTypeText,
				},
				{
					Field:    "status",
					Value:    "pending",
					Mode:     filter.ModeEqual,
					DataType: filter.DataTypeText,
				},
			},
		}

		pageIndex := 1
		pageSize := 10

		result, err := handler.DataGorm(presetDB, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("Error filtering: %v", err)
		}

		// Should get records from org=1 that are either EUR or pending
		// EUR: 1 record, pending: 1 record = 2 total
		if result.TotalSize != 2 {
			t.Errorf("Expected 2 records (EUR or pending) for org=1, got %d", result.TotalSize)
		}
	})
}

// TestDataGorm_ComplexPresetConditions tests more complex preset scenarios
func TestDataGorm_ComplexPresetConditions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&BillAndCoin{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	records := []BillAndCoin{
		{OrganizationID: 1, BranchID: 1, Amount: 100.0, Currency: "USD", Status: "active"},
		{OrganizationID: 1, BranchID: 2, Amount: 200.0, Currency: "USD", Status: "active"},
		{OrganizationID: 1, BranchID: 3, Amount: 300.0, Currency: "USD", Status: "inactive"},
		{OrganizationID: 2, BranchID: 1, Amount: 400.0, Currency: "EUR", Status: "active"},
	}

	if err := db.Create(&records).Error; err != nil {
		t.Fatalf("Failed to create test records: %v", err)
	}

	handler := filter.NewFilter[BillAndCoin]()

	t.Run("ComplexPresetWithIN", func(t *testing.T) {
		// Preset: organization_id=1 AND branch_id IN (1,2)
		presetDB := db.Where("organization_id = ? AND branch_id IN ?", 1, []uint{1, 2})

		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "status",
					Value:    "active",
					Mode:     filter.ModeEqual,
					DataType: filter.DataTypeText,
				},
			},
		}

		pageIndex := 1
		pageSize := 10

		result, err := handler.DataGorm(presetDB, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("Error filtering: %v", err)
		}

		// Should get 2 records: org=1, branch IN (1,2), status=active
		if result.TotalSize != 2 {
			t.Errorf("Expected 2 records, got %d", result.TotalSize)
		}
	})

	t.Run("PresetWithRange", func(t *testing.T) {
		// Preset: amount >= 200
		presetDB := db.Where("amount >= ?", 200)

		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "organization_id",
					Value:    1,
					Mode:     filter.ModeEqual,
					DataType: filter.DataTypeNumber,
				},
			},
		}

		pageIndex := 1
		pageSize := 10

		result, err := handler.DataGorm(presetDB, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("Error filtering: %v", err)
		}

		// Should get 2 records: org=1 with amount>=200
		if result.TotalSize != 2 {
			t.Errorf("Expected 2 records, got %d", result.TotalSize)
		}

		for _, record := range result.Data {
			if record.Amount < 200 {
				t.Errorf("Expected amount>=200, got %f", record.Amount)
			}
		}
	})
}
