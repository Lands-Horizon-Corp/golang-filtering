package test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestFilterHandler_Hybrid_EmptyTable tests hybrid mode with empty table
func TestFilterHandler_Hybrid_EmptyTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	result, err := handler.Hybrid(db, 100, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalSize != 0 {
		t.Errorf("Expected total size 0, got %d", result.TotalSize)
	}
}

// TestFilterHandler_Hybrid_SmallDataset tests hybrid with small dataset (should use in-memory)
func TestFilterHandler_Hybrid_SmallDataset(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "role",
				Value:    "admin",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	// Threshold is 1000, we have 10 records, should use in-memory
	result, err := handler.Hybrid(db, 1000, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find 3 admins
	if result.TotalSize != 3 {
		t.Errorf("Expected 3 admins, got %d", result.TotalSize)
	}
}

// TestFilterHandler_Hybrid_LargeThreshold tests hybrid with threshold forcing database query
func TestFilterHandler_Hybrid_LargeThreshold(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				Value:    true,
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeBool,
			},
		},
	}

	// Threshold is 5, we have 10 records, should use database
	result, err := handler.Hybrid(db, 5, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find active users
	for _, user := range result.Data {
		if !user.IsActive {
			t.Errorf("Expected all users to be active, got %s (inactive)", user.Name)
		}
	}
}

// TestFilterHandler_Hybrid_ComplexFilter tests hybrid with complex filters
func TestFilterHandler_Hybrid_ComplexFilter(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				Value:    true,
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeBool,
			},
			{
				Field:    "age",
				Value:    filter.Range{From: 25, To: 35},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.SortOrderAsc},
		},
	}

	result, err := handler.Hybrid(db, 100, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find active users between 25-35
	for _, user := range result.Data {
		if !user.IsActive {
			t.Errorf("Expected active users, got %s (inactive)", user.Name)
		}
		if user.Age < 25 || user.Age > 35 {
			t.Errorf("Expected age 25-35, got %d for %s", user.Age, user.Name)
		}
	}

	// Check sorting
	for i := 1; i < len(result.Data); i++ {
		if result.Data[i-1].Age > result.Data[i].Age {
			t.Errorf("Expected ascending age order, got %d before %d", result.Data[i-1].Age, result.Data[i].Age)
		}
	}
}

// TestFilterHandler_Hybrid_OrLogic tests hybrid with OR logic
func TestFilterHandler_Hybrid_OrLogic(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicOr,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "role",
				Value:    "admin",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "age",
				Value:    25,
				Mode:     filter.ModeLT,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.Hybrid(db, 50, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find admins OR users younger than 25
	for _, user := range result.Data {
		if user.Role != "admin" && user.Age >= 25 {
			t.Errorf("Expected admin or age < 25, got %s (role=%s, age=%d)", user.Name, user.Role, user.Age)
		}
	}
}

// TestFilterHandler_Hybrid_Pagination tests hybrid pagination
func TestFilterHandler_Hybrid_Pagination(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	// Page 1, size 3
	result1, err := handler.Hybrid(db, 100, filterRoot, 1, 3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result1.Data) != 3 {
		t.Errorf("Expected 3 items on page 1, got %d", len(result1.Data))
	}

	// Page 2, size 3
	result2, err := handler.Hybrid(db, 100, filterRoot, 2, 3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result2.Data) != 3 {
		t.Errorf("Expected 3 items on page 2, got %d", len(result2.Data))
	}

	// Verify pages don't overlap
	for _, user1 := range result1.Data {
		for _, user2 := range result2.Data {
			if user1.ID == user2.ID {
				t.Errorf("Found duplicate user %d on both pages", user1.ID)
			}
		}
	}
}

// TestFilterHandler_Hybrid_ConsistencyCheck tests consistency between methods
func TestFilterHandler_Hybrid_ConsistencyCheck(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				Value:    true,
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeBool,
			},
		},
	}

	// Get all users for in-memory comparison
	var allUsers []*TestUser
	if err := db.Find(&allUsers).Error; err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	// Test DataQuery
	resultQuery, err := handler.DataQuery(allUsers, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DataQuery failed: %v", err)
	}

	// Test DataGorm
	resultGorm, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DataGorm failed: %v", err)
	}

	// Test Hybrid (with low threshold to use database)
	resultHybridDB, err := handler.Hybrid(db, 5, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Hybrid (database) failed: %v", err)
	}

	// Test Hybrid (with high threshold to use in-memory)
	resultHybridMem, err := handler.Hybrid(db, 1000, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Hybrid (in-memory) failed: %v", err)
	}

	// All methods should return the same total count
	if resultQuery.TotalSize != resultGorm.TotalSize {
		t.Errorf("DataQuery (%d) and DataGorm (%d) returned different totals", resultQuery.TotalSize, resultGorm.TotalSize)
	}
	if resultQuery.TotalSize != resultHybridDB.TotalSize {
		t.Errorf("DataQuery (%d) and Hybrid-DB (%d) returned different totals", resultQuery.TotalSize, resultHybridDB.TotalSize)
	}
	if resultQuery.TotalSize != resultHybridMem.TotalSize {
		t.Errorf("DataQuery (%d) and Hybrid-Mem (%d) returned different totals", resultQuery.TotalSize, resultHybridMem.TotalSize)
	}
}

// TestFilterHandler_Hybrid_MultipleFilters tests hybrid with multiple filters
func TestFilterHandler_Hybrid_MultipleFilters(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "John",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "age",
				Value:    20,
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeNumber,
			},
			{
				Field:    "is_active",
				Value:    true,
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeBool,
			},
		},
	}

	result, err := handler.Hybrid(db, 100, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify all conditions are met
	for _, user := range result.Data {
		if !containsIgnoreCase(user.Name, "John") {
			t.Errorf("Expected name to contain 'John', got %s", user.Name)
		}
		if user.Age <= 20 {
			t.Errorf("Expected age > 20, got %d", user.Age)
		}
		if !user.IsActive {
			t.Errorf("Expected active user, got inactive %s", user.Name)
		}
	}
}
