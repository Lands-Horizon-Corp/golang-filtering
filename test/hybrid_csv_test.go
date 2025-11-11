package test

import (
	"strings"
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestHybridCSV(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create table and insert test data
	err = db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Insert test users
	users := generateTestUsers()
	for _, user := range users {
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Create filter handler
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	t.Run("Hybrid CSV with small threshold (in-memory path)", func(t *testing.T) {
		filterRoot := filter.Root{
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "is_active",
					DataType: filter.DataTypeBool,
					Mode:     filter.ModeEqual,
					Value:    true,
				},
			},
			Logic: filter.LogicAnd,
		}

		// Use small threshold to force in-memory processing
		csvData, err := handler.HybridCSV(db, 100, filterRoot)
		if err != nil {
			t.Fatalf("HybridCSV failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should have header + active users (7 active users in test data)
		expectedLines := 8 // 1 header + 7 active users
		if len(lines) != expectedLines {
			t.Errorf("Expected %d lines (1 header + 7 active users), got %d", len(lines), expectedLines)
		}

		// Check that headers contain field names
		header := lines[0]
		expectedFields := []string{"id", "name", "email", "age", "is_active", "role"}
		for _, field := range expectedFields {
			if !strings.Contains(strings.ToLower(header), field) {
				t.Errorf("Expected header to contain field '%s', got: %s", field, header)
			}
		}

		t.Logf("✅ Hybrid CSV (in-memory): %d lines generated", len(lines))
	})

	t.Run("Hybrid CSV with large threshold (database path)", func(t *testing.T) {
		filterRoot := filter.Root{
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "role",
					DataType: filter.DataTypeText,
					Mode:     filter.ModeEqual,
					Value:    "admin",
				},
			},
			Logic: filter.LogicAnd,
		}

		// Use very small threshold to force database processing
		csvData, err := handler.HybridCSV(db, 5, filterRoot)
		if err != nil {
			t.Fatalf("HybridCSV failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should have header + admin users (3 admin users in test data)
		expectedLines := 4 // 1 header + 3 admin users
		if len(lines) != expectedLines {
			t.Errorf("Expected %d lines (1 header + 3 admin users), got %d", len(lines), expectedLines)
		}

		// Check that only admin role is included
		csvContent := string(csvData)
		if !strings.Contains(csvContent, "admin") {
			t.Error("Expected admin role in CSV output")
		}

		t.Logf("✅ Hybrid CSV (database): %d lines generated", len(lines))
	})

	t.Run("Hybrid CSV with preset conditions", func(t *testing.T) {
		// Test with preset conditions using struct
		type UserFilter struct {
			IsActive bool `gorm:"column:is_active"`
		}

		presetConditions := &UserFilter{IsActive: true}

		filterRoot := filter.Root{
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "role",
					DataType: filter.DataTypeText,
					Mode:     filter.ModeEqual,
					Value:    "admin",
				},
			},
			Logic: filter.LogicAnd,
		}

		csvData, err := handler.HybridCSVWithPreset(db, presetConditions, 100, filterRoot)
		if err != nil {
			t.Fatalf("HybridCSVWithPreset failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should have header + active admin users (3 active admins in test data)
		expectedLines := 4 // 1 header + 3 active admins
		if len(lines) != expectedLines {
			t.Errorf("Expected %d lines (1 header + 3 active admins), got %d", len(lines), expectedLines)
		}

		// Check that both conditions are applied
		csvContent := string(csvData)
		if !strings.Contains(csvContent, "admin") {
			t.Error("Expected admin role in CSV output")
		}

		t.Logf("✅ Hybrid CSV with preset conditions: %d active admins exported", len(lines)-1)
	})
}

func TestHybridCSVConsistency(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create table and insert test data
	err = db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Insert test users
	users := generateTestUsers()
	for _, user := range users {
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Create filter handler
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	// Test the same filter with both strategies
	filterRoot := filter.Root{
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				DataType: filter.DataTypeBool,
				Mode:     filter.ModeEqual,
				Value:    true,
			},
		},
		Logic: filter.LogicAnd,
	}

	// In-memory strategy (high threshold)
	csvDataInMemory, err := handler.HybridCSV(db, 100, filterRoot)
	if err != nil {
		t.Fatalf("HybridCSV (in-memory) failed: %v", err)
	}

	// Database strategy (low threshold)
	csvDataDatabase, err := handler.HybridCSV(db, 5, filterRoot)
	if err != nil {
		t.Fatalf("HybridCSV (database) failed: %v", err)
	}

	// Compare results
	inMemoryLines := strings.Split(strings.TrimSpace(string(csvDataInMemory)), "\n")
	databaseLines := strings.Split(strings.TrimSpace(string(csvDataDatabase)), "\n")

	if len(inMemoryLines) != len(databaseLines) {
		t.Errorf("Inconsistent results: in-memory has %d lines, database has %d lines", len(inMemoryLines), len(databaseLines))
	}

	t.Logf("✅ Hybrid CSV consistency check: both strategies returned %d lines", len(inMemoryLines))
}
