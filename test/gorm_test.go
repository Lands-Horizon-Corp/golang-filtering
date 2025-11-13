package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a test database with sample data
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Insert test data
	users := generateTestUsers()
	for _, user := range users {
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
	}

	return db
}

// TestFilterHandler_DataGorm_EmptyTable tests filtering empty table
func TestFilterHandler_DataGorm_EmptyTable(t *testing.T) {
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

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalSize != 0 {
		t.Errorf("Expected total size 0, got %d", result.TotalSize)
	}
	if len(result.Data) != 0 {
		t.Errorf("Expected empty data, got %d items", len(result.Data))
	}
}

// TestFilterHandler_DataGorm_NoFilters tests with no filters applied
func TestFilterHandler_DataGorm_NoFilters(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.TotalSize != 10 {
		t.Errorf("Expected total size 10, got %d", result.TotalSize)
	}
}

// TestFilterHandler_DataGorm_ModeEqual tests equal filter mode
func TestFilterHandler_DataGorm_ModeEqual(t *testing.T) {
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

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find 3 admins
	if result.TotalSize != 3 {
		t.Errorf("Expected 3 admins, got %d", result.TotalSize)
	}

	for _, user := range result.Data {
		if user.Role != "admin" {
			t.Errorf("Expected all users to be admins, got role %s", user.Role)
		}
	}
}

// TestFilterHandler_DataGorm_ModeContains tests contains filter mode
func TestFilterHandler_DataGorm_ModeContains(t *testing.T) {
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
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find 3 users with "John" in name
	if result.TotalSize != 3 {
		t.Errorf("Expected 3 users, got %d", result.TotalSize)
	}
}

// TestFilterHandler_DataGorm_ModeGTE tests greater than or equal filter
func TestFilterHandler_DataGorm_ModeGTE(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find users 30 or older
	for _, user := range result.Data {
		if user.Age < 30 {
			t.Errorf("Expected age >= 30, got %d for %s", user.Age, user.Name)
		}
	}
}

// TestFilterHandler_DataGorm_ModeRange tests range filter
func TestFilterHandler_DataGorm_ModeRange(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    filter.Range{From: 28, To: 35},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find users between 28 and 35
	for _, user := range result.Data {
		if user.Age < 28 || user.Age > 35 {
			t.Errorf("Expected age between 28-35, got %d for %s", user.Age, user.Name)
		}
	}
}

// TestFilterHandler_DataGorm_BoolFilter tests boolean filter
func TestFilterHandler_DataGorm_BoolFilter(t *testing.T) {
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

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find only active users
	for _, user := range result.Data {
		if !user.IsActive {
			t.Errorf("Expected all users to be active, got %s (inactive)", user.Name)
		}
	}
}

// TestFilterHandler_DataGorm_LogicAnd tests AND logic
func TestFilterHandler_DataGorm_LogicAnd(t *testing.T) {
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
				Value:    30,
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find active users who are 30 or older
	for _, user := range result.Data {
		if !user.IsActive || user.Age < 30 {
			t.Errorf("Expected active users aged >= 30, got %s (active=%v, age=%d)", user.Name, user.IsActive, user.Age)
		}
	}
}

// TestFilterHandler_DataGorm_LogicOr tests OR logic
func TestFilterHandler_DataGorm_LogicOr(t *testing.T) {
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
				Field:    "role",
				Value:    "moderator",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find admins OR moderators
	for _, user := range result.Data {
		if user.Role != "admin" && user.Role != "moderator" {
			t.Errorf("Expected admin or moderator, got role %s for %s", user.Role, user.Name)
		}
	}
}

// TestFilterHandler_DataGorm_Pagination tests pagination
func TestFilterHandler_DataGorm_Pagination(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	// Page 1 (second page in 0-based indexing), size 3
	result, err := handler.DataGorm(db, filterRoot, 1, 3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.PageIndex != 1 {
		t.Errorf("Expected page index 1 (0-based), got %d", result.PageIndex)
	}
	if result.PageSize != 3 {
		t.Errorf("Expected page size 3, got %d", result.PageSize)
	}
	if len(result.Data) != 3 {
		t.Errorf("Expected 3 items on page 1, got %d", len(result.Data))
	}
	if result.TotalSize != 10 {
		t.Errorf("Expected total size 10, got %d", result.TotalSize)
	}
}

// TestFilterHandler_DataGorm_Sorting tests sorting
func TestFilterHandler_DataGorm_Sorting(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.SortOrderDesc},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be sorted by age descending
	for i := 1; i < len(result.Data); i++ {
		if result.Data[i-1].Age < result.Data[i].Age {
			t.Errorf("Expected descending age order, got %d before %d", result.Data[i-1].Age, result.Data[i].Age)
		}
	}
}

// TestFilterHandler_DataGorm_DateFilter tests date filtering
func TestFilterHandler_DataGorm_DateFilter(t *testing.T) {
	db := setupTestDB(t)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	dateThreshold := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "created_at",
				Value:    dateThreshold,
				Mode:     filter.ModeAfter,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should find users created after March 1, 2024
	for _, user := range result.Data {
		if user.CreatedAt.Before(dateThreshold) {
			t.Errorf("Expected created_at after %v, got %v for %s", dateThreshold, user.CreatedAt, user.Name)
		}
	}
}
