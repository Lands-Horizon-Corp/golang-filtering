package test

import (
	"strings"
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormNoPaginationCSV(t *testing.T) {
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

	t.Run("CSV export with no filters", func(t *testing.T) {
		filterRoot := filter.Root{
			FieldFilters: []filter.FieldFilter{},
			Logic:        filter.LogicAnd,
		}

		csvData, err := handler.GormNoPaginationCSV(db, filterRoot)
		if err != nil {
			t.Fatalf("GormNoPaginationCSV failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should have header + 10 data rows
		if len(lines) != 11 {
			t.Errorf("Expected 11 lines (1 header + 10 data), got %d", len(lines))
		}

		// Check that headers contain field names
		header := lines[0]
		expectedFields := []string{"id", "name", "email", "age", "is_active", "role"}
		for _, field := range expectedFields {
			if !strings.Contains(strings.ToLower(header), field) {
				t.Errorf("Expected header to contain field '%s', got: %s", field, header)
			}
		}

		// Check first data row contains some user data
		if !strings.Contains(csvString, "John Doe") && !strings.Contains(csvString, "Jane Smith") {
			t.Errorf("Expected user data in CSV output, got: %s", csvString)
		}

		t.Logf("✅ GORM CSV export with no filters: %d lines generated", len(lines))
	})

	t.Run("CSV export with active filter", func(t *testing.T) {
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

		csvData, err := handler.GormNoPaginationCSV(db, filterRoot)
		if err != nil {
			t.Fatalf("GormNoPaginationCSV failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should have header + active users (7 active users in test data)
		expectedLines := 8 // 1 header + 7 active users
		if len(lines) != expectedLines {
			t.Errorf("Expected %d lines (1 header + 7 active users), got %d", len(lines), expectedLines)
		}

		t.Logf("✅ GORM CSV export with active filter: %d active users exported", len(lines)-1)
	})

	t.Run("CSV export with preset conditions", func(t *testing.T) {
		// Test with preset conditions using struct
		type UserFilter struct {
			Role string `gorm:"column:role"`
		}

		presetConditions := &UserFilter{Role: "admin"}

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

		csvData, err := handler.GormNoPaginationCSVWithPreset(db, presetConditions, filterRoot)
		if err != nil {
			t.Fatalf("GormNoPaginationCSVWithPreset failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should have header + active admin users (3 active admins in test data: John Doe, Charlie Wilson, Grace Lee)
		expectedLines := 4 // 1 header + 3 active admins
		if len(lines) != expectedLines {
			t.Errorf("Expected %d lines (1 header + 3 active admins), got %d", len(lines), expectedLines)
		}

		// Check that only admin users are included
		csvContent := string(csvData)
		if !strings.Contains(csvContent, "admin") {
			t.Error("Expected admin role in CSV output")
		}

		t.Logf("✅ GORM CSV export with preset conditions: %d active admins exported", len(lines)-1)
	})
}

func TestGormNoPaginationCSVEmptyResult(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create table but don't insert any data
	err = db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create filter handler
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	csvData, err := handler.GormNoPaginationCSV(db, filterRoot)
	if err != nil {
		t.Fatalf("GormNoPaginationCSV failed: %v", err)
	}

	csvString := string(csvData)
	lines := strings.Split(strings.TrimSpace(csvString), "\n")

	// Should have only header
	if len(lines) != 1 {
		t.Errorf("Expected 1 line (header only), got %d", len(lines))
	}

	// Check that header contains field names
	header := lines[0]
	if header == "" {
		t.Errorf("Expected non-empty header, got: %s", header)
	}

	t.Logf("✅ GORM CSV export with empty result: header only generated")
}
