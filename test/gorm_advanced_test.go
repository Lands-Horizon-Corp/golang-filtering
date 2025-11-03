package test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestGormTextModes tests all text filtering modes in GORM
func TestGormTextModes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	users := generateTestUsers()
	if err := db.Create(users).Error; err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	handler := filter.NewFilter[TestUser]()

	tests := []struct {
		name    string
		mode    filter.Mode
		value   string
		checkFn func(t *testing.T, result *filter.PaginationResult[TestUser])
	}{
		{
			name:  "NotContains",
			mode:  filter.ModeNotContains,
			value: "xyz",
			checkFn: func(t *testing.T, result *filter.PaginationResult[TestUser]) {
				if result.TotalSize == 0 {
					t.Error("Expected some users without 'xyz' in name")
				}
			},
		},
		{
			name:  "StartsWith",
			mode:  filter.ModeStartsWith,
			value: "John",
			checkFn: func(t *testing.T, result *filter.PaginationResult[TestUser]) {
				for _, user := range result.Data {
					if len(user.Name) < 4 || user.Name[:4] != "John" {
						t.Errorf("Expected name to start with 'John', got %s", user.Name)
					}
				}
			},
		},
		{
			name:  "EndsWith",
			mode:  filter.ModeEndsWith,
			value: "Smith",
			checkFn: func(t *testing.T, result *filter.PaginationResult[TestUser]) {
				for _, user := range result.Data {
					nameLen := len(user.Name)
					if nameLen < 5 || user.Name[nameLen-5:] != "Smith" {
						t.Errorf("Expected name to end with 'Smith', got %s", user.Name)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    "name",
						Value:    tt.value,
						Mode:     tt.mode,
						DataType: filter.DataTypeText,
					},
				},
			}

			result, err := handler.DataGorm(db, filterRoot, 1, 100)
			if err != nil {
				t.Fatalf("DataGorm %s failed: %v", tt.name, err)
			}

			tt.checkFn(t, result)
		})
	}
}

// TestGormEmptyChecks tests IsEmpty and IsNotEmpty modes
func TestGormEmptyChecks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	type Document struct {
		ID          uint   `gorm:"primarykey" json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := db.AutoMigrate(&Document{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	docs := []Document{
		{ID: 1, Title: "Doc1", Description: "Has description"},
		{ID: 2, Title: "Doc2", Description: ""},
		{ID: 3, Title: "Doc3", Description: "Also has description"},
	}

	if err := db.Create(&docs).Error; err != nil {
		t.Fatalf("Failed to create documents: %v", err)
	}

	handler := filter.NewFilter[Document]()

	// Test IsEmpty
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "description",
				Value:    "",
				Mode:     filter.ModeIsEmpty,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DataGorm IsEmpty failed: %v", err)
	}

	if result.TotalSize != 1 {
		t.Errorf("Expected 1 document with empty description, got %d", result.TotalSize)
	}

	// Test IsNotEmpty
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "description",
				Value:    "",
				Mode:     filter.ModeIsNotEmpty,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err = handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DataGorm IsNotEmpty failed: %v", err)
	}

	if result.TotalSize != 2 {
		t.Errorf("Expected 2 documents with non-empty description, got %d", result.TotalSize)
	}
}

// TestGormDateTimeFiltering tests date and datetime filtering in database
func TestGormDateTimeFiltering(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	users := generateTestUsers()
	if err := db.Create(users).Error; err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	handler := filter.NewFilter[TestUser]()

	// Test date range
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "created_at",
				Value: filter.Range{
					From: "2024-01-01",
					To:   "2024-03-31",
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DataGorm date range failed: %v", err)
	}

	t.Logf("Date range filter returned %d users", result.TotalSize)
}

// TestGormNumberRanges tests number range filtering
func TestGormNumberRanges(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	users := generateTestUsers()
	if err := db.Create(users).Error; err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	handler := filter.NewFilter[TestUser]()

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field: "age",
				Value: filter.Range{
					From: 25,
					To:   35,
				},
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeNumber,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DataGorm number range failed: %v", err)
	}

	for _, user := range result.Data {
		if user.Age < 25 || user.Age > 35 {
			t.Errorf("Expected age between 25 and 35, got %d for user %s", user.Age, user.Name)
		}
	}
}

// TestGormComplexConditions tests complex SQL conditions
func TestGormComplexConditions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	users := generateTestUsers()
	if err := db.Create(users).Error; err != nil {
		t.Fatalf("Failed to create users: %v", err)
	}

	handler := filter.NewFilter[TestUser]()

	filterRoot := filter.Root{
		Logic: filter.LogicOr,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     filter.ModeLT,
				DataType: filter.DataTypeNumber,
			},
			{
				Field:    "role",
				Value:    "admin",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
	}

	result, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("DataGorm complex conditions failed: %v", err)
	}

	// Verify OR logic
	for _, user := range result.Data {
		matches := user.Age < 30 || user.Role == "admin"
		if !matches {
			t.Errorf("User %s (age %d, role %s) should not match OR condition", user.Name, user.Age, user.Role)
		}
	}
}
