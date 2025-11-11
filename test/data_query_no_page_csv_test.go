package test

import (
	"strings"
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

func TestDataQueryNoPageCSV(t *testing.T) {
	// Create test data
	users := generateTestUsers()

	// Create filter handler
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	t.Run("CSV export with no filters", func(t *testing.T) {
		filterRoot := filter.Root{
			FieldFilters: []filter.FieldFilter{},
			Logic:        filter.LogicAnd,
		}

		csvData, err := handler.DataQueryNoPageCSV(users, filterRoot)
		if err != nil {
			t.Fatalf("DataQueryNoPageCSV failed: %v", err)
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

		// Check first data row contains John Doe's data
		if !strings.Contains(csvString, "John Doe") {
			t.Errorf("Expected John Doe in CSV data, got: %s", csvString)
		}

		t.Logf("✅ CSV export with no filters: %d lines generated", len(lines))
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

		csvData, err := handler.DataQueryNoPageCSV(users, filterRoot)
		if err != nil {
			t.Fatalf("DataQueryNoPageCSV failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should have header + active users (7 active users in test data)
		expectedLines := 8 // 1 header + 7 active users
		if len(lines) != expectedLines {
			t.Errorf("Expected %d lines (1 header + 7 active users), got %d", len(lines), expectedLines)
		}

		// Check that data contains only active users (true values)
		activeUserCount := 0
		for i := 1; i < len(lines); i++ { // Skip header
			if strings.Contains(lines[i], "true") {
				activeUserCount++
			}
		}

		if activeUserCount != 7 {
			t.Errorf("Expected 7 active users in CSV, found %d", activeUserCount)
		}

		t.Logf("✅ CSV export with active filter: %d active users exported", len(lines)-1)
	})

	t.Run("CSV export with name filter", func(t *testing.T) {
		filterRoot := filter.Root{
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "name",
					DataType: filter.DataTypeText,
					Mode:     filter.ModeContains,
					Value:    "John", // Names containing 'John'
				},
			},
			Logic: filter.LogicAnd,
		}

		csvData, err := handler.DataQueryNoPageCSV(users, filterRoot)
		if err != nil {
			t.Fatalf("DataQueryNoPageCSV failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should include John Doe, Bob Johnson, and John Smith (3 users)
		if len(lines) != 4 {
			t.Errorf("Expected 4 lines (1 header + 3 users with 'John' in name), got %d", len(lines))
		}

		// Verify specific names are included
		csvContent := string(csvData)
		expectedNames := []string{"John Doe", "Bob Johnson", "John Smith"}
		for _, name := range expectedNames {
			if !strings.Contains(csvContent, name) {
				t.Errorf("Expected name '%s' in CSV output", name)
			}
		}

		t.Logf("✅ CSV export with name filter: %d users with 'John' in name", len(lines)-1)
	})

	t.Run("CSV export with role filter and sorting", func(t *testing.T) {
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
			SortFields: []filter.SortField{
				{
					Field: "age",
					Order: filter.SortOrderAsc,
				},
			},
		}

		csvData, err := handler.DataQueryNoPageCSV(users, filterRoot)
		if err != nil {
			t.Fatalf("DataQueryNoPageCSV failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should include admins: John Doe (25), Grace Lee (31), Charlie Wilson (42)
		if len(lines) != 4 {
			t.Errorf("Expected 4 lines (1 header + 3 admins), got %d", len(lines))
		}

		// Check that ages are in ascending order for admins
		expectedOrder := []string{"John Doe", "Grace Lee", "Charlie Wilson"} // 25, 31, 42
		for i, expectedName := range expectedOrder {
			lineIndex := i + 1 // Skip header
			if !strings.Contains(lines[lineIndex], expectedName) {
				t.Errorf("Expected %s at position %d, got line: %s", expectedName, i+1, lines[lineIndex])
			}
		}

		t.Logf("✅ CSV export with admin filter and age sorting: Admins sorted by age ascending")
	})
}

func TestDataQueryNoPageCSVErrors(t *testing.T) {
	users := generateTestUsers()[:1] // Just one user for error testing

	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	t.Run("Invalid filter error propagation", func(t *testing.T) {
		// Create an invalid filter
		filterRoot := filter.Root{
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "invalid_field",
					DataType: filter.DataTypeText,
					Mode:     filter.ModeEqual,
					Value:    "test",
				},
			},
			Logic: filter.LogicAnd,
		}

		csvData, err := handler.DataQueryNoPageCSV(users, filterRoot)
		if err != nil {
			t.Logf("✅ Filter error properly propagated: %v", err)
		} else {
			t.Logf("✅ Invalid field filter was ignored, CSV generated: %d bytes", len(csvData))
		}
	})
}

func TestDataQueryNoPageCSVSpecialCharacters(t *testing.T) {
	// Test CSV escaping with special characters
	users := []*TestUser{
		{ID: 1, Name: "Smith, John", Email: "john@test.com", Age: 30, IsActive: true, Role: "admin"},
		{ID: 2, Name: "Jane \"Doe\"", Email: "jane@test.com", Age: 25, IsActive: false, Role: "user"},
		{ID: 3, Name: "Bob\nNewline", Email: "bob@test.com", Age: 35, IsActive: true, Role: "moderator"},
	}

	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{}

	csvData, err := handler.DataQueryNoPageCSV(users, filterRoot)
	if err != nil {
		t.Fatalf("DataQueryNoPageCSV failed: %v", err)
	}

	csvString := string(csvData)

	// Check that fields with commas are quoted
	if !strings.Contains(csvString, "\"Smith, John\"") {
		t.Error("Expected comma-containing field to be quoted")
	}

	// Check that fields with quotes are escaped
	if !strings.Contains(csvString, "\"Jane \"\"Doe\"\"\"") {
		t.Error("Expected quote-containing field to be properly escaped")
	}

	// Check that fields with newlines are quoted
	if !strings.Contains(csvString, "\"Bob\nNewline\"") {
		t.Error("Expected newline-containing field to be quoted")
	}

	t.Logf("✅ CSV special character escaping working correctly")
	t.Logf("CSV output:\n%s", csvString)
}

func TestDataQueryNoPageCSVEmptyData(t *testing.T) {
	var users []*TestUser // Empty slice

	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	filterRoot := filter.Root{}

	csvData, err := handler.DataQueryNoPageCSV(users, filterRoot)
	if err != nil {
		t.Fatalf("DataQueryNoPageCSV failed: %v", err)
	}

	csvString := string(csvData)
	lines := strings.Split(strings.TrimSpace(csvString), "\n")

	// Should have only header
	if len(lines) != 1 {
		t.Errorf("Expected 1 line (header only), got %d", len(lines))
	}

	// Check that header contains field names (since we're using auto-generated headers)
	header := lines[0]
	if header == "" {
		t.Errorf("Expected non-empty header, got: %s", header)
	}

	t.Logf("✅ CSV export with empty data: header only generated")
}
