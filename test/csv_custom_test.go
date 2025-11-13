package test

import (
	"strings"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

func TestDataQueryNoPageCSVCustom(t *testing.T) {
	// Generate test data
	users := generateTestUsers()

	// Create handler (the getters won't be used since we're using custom callback)
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	t.Run("Custom CSV with custom field mapping", func(t *testing.T) {
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

		// Custom callback that creates different field names and formats
		csvData, err := handler.DataQueryNoPageCSVCustom(users, filterRoot, func(user *TestUser) map[string]any {
			return map[string]any{
				"User ID":    user.ID,
				"Full Name":  user.Name,
				"Email Addr": user.Email,
				"User Age":   user.Age,
				"Active":     user.IsActive,
				"User Role":  strings.ToUpper(user.Role),          // Transform data
				"Created":    user.CreatedAt.Format("2006-01-02"), // Format dates
			}
		})

		if err != nil {
			t.Fatalf("DataQueryNoPageCSVCustom failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should have header + filtered active users (7 active users)
		expectedLines := 8 // 1 header + 7 active users
		if len(lines) != expectedLines {
			t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
		}

		// Check headers are correctly ordered alphabetically
		headers := strings.Split(lines[0], ",")
		expectedHeaders := []string{"Active", "Created", "Email Addr", "Full Name", "User Age", "User ID", "User Role"}

		if len(headers) != len(expectedHeaders) {
			t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(headers))
		}

		for i, expected := range expectedHeaders {
			if i >= len(headers) || headers[i] != expected {
				t.Errorf("Expected header %d to be '%s', got '%s'", i, expected, headers[i])
			}
		}

		// Check that data contains transformed values
		if !strings.Contains(csvString, "ADMIN") && !strings.Contains(csvString, "USER") {
			t.Errorf("Expected uppercase roles in CSV output")
		}

		// Check date formatting
		if !strings.Contains(csvString, "2024-") {
			t.Errorf("Expected formatted dates in CSV output")
		}

		t.Logf("✅ Custom CSV with field mapping: %d active users exported with custom headers", len(lines)-1)
		t.Logf("Headers: %s", lines[0])
	})

	t.Run("Custom CSV with nested field access", func(t *testing.T) {
		// Create test data with nested structure simulation
		type Department struct {
			Name string
			Code string
		}

		type ExtendedUser struct {
			TestUser
			Department *Department
		}

		extendedUsers := make([]*ExtendedUser, len(users))
		for i, user := range users {
			dept := &Department{
				Name: "Engineering",
				Code: "ENG",
			}
			if user.Role == "admin" {
				dept.Name = "Administration"
				dept.Code = "ADM"
			}
			extendedUsers[i] = &ExtendedUser{
				TestUser:   *user,
				Department: dept,
			}
		}

		// Create handler for extended users
		extendedHandler := filter.NewFilter[ExtendedUser](filter.GolangFilteringConfig{})

		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
		}

		// Custom callback accessing nested fields
		csvData, err := extendedHandler.DataQueryNoPageCSVCustom(extendedUsers, filterRoot, func(user *ExtendedUser) map[string]any {
			return map[string]any{
				"Employee":  user.Name,
				"Dept Name": user.Department.Name,
				"Dept Code": user.Department.Code,
				"Email":     user.Email,
				"Status":    map[bool]string{true: "Active", false: "Inactive"}[user.IsActive],
			}
		})

		if err != nil {
			t.Fatalf("DataQueryNoPageCSVCustom with nested fields failed: %v", err)
		}

		csvString := string(csvData)
		lines := strings.Split(strings.TrimSpace(csvString), "\n")

		// Should have header + all users (10 users)
		expectedLines := 11 // 1 header + 10 users
		if len(lines) != expectedLines {
			t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
		}

		// Check that nested field access works
		if !strings.Contains(csvString, "Engineering") && !strings.Contains(csvString, "Administration") {
			t.Errorf("Expected department names in CSV output")
		}

		if !strings.Contains(csvString, "ENG") && !strings.Contains(csvString, "ADM") {
			t.Errorf("Expected department codes in CSV output")
		}

		// Check status transformation
		if !strings.Contains(csvString, "Active") && !strings.Contains(csvString, "Inactive") {
			t.Errorf("Expected status transformation in CSV output")
		}

		t.Logf("✅ Custom CSV with nested fields: %d users exported with department info", len(lines)-1)
	})

	t.Run("Custom CSV with empty result", func(t *testing.T) {
		// Filter that returns no results
		filterRoot := filter.Root{
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "age",
					DataType: filter.DataTypeNumber,
					Mode:     filter.ModeEqual,
					Value:    999, // No user has age 999
				},
			},
			Logic: filter.LogicAnd,
		}

		csvData, err := handler.DataQueryNoPageCSVCustom(users, filterRoot, func(user *TestUser) map[string]any {
			return map[string]any{
				"Name": user.Name,
				"Age":  user.Age,
			}
		})

		if err != nil {
			t.Fatalf("DataQueryNoPageCSVCustom with empty result failed: %v", err)
		}

		// Should return empty CSV when no results
		csvString := string(csvData)
		if csvString != "" {
			t.Errorf("Expected empty CSV for no results, got: %s", csvString)
		}

		t.Logf("✅ Custom CSV with empty result: correctly returns empty CSV")
	})

	t.Run("Custom CSV deterministic ordering", func(t *testing.T) {
		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
		}

		// Generate CSV multiple times with same callback
		customCallback := func(user *TestUser) map[string]any {
			return map[string]any{
				"name":       user.Name,
				"email":      user.Email,
				"age":        user.Age,
				"created_at": user.CreatedAt.Format(time.RFC3339),
				"role":       user.Role,
			}
		}

		var csvOutputs []string
		for i := 0; i < 3; i++ {
			csvData, err := handler.DataQueryNoPageCSVCustom(users, filterRoot, customCallback)
			if err != nil {
				t.Fatalf("DataQueryNoPageCSVCustom failed on iteration %d: %v", i, err)
			}
			csvOutputs = append(csvOutputs, string(csvData))
		}

		// All outputs should be identical (deterministic ordering)
		for i := 1; i < len(csvOutputs); i++ {
			if csvOutputs[i] != csvOutputs[0] {
				t.Errorf("CSV output %d differs from first output", i)
				t.Logf("First output headers: %s", strings.Split(csvOutputs[0], "\n")[0])
				t.Logf("Output %d headers: %s", i, strings.Split(csvOutputs[i], "\n")[0])
			}
		}

		// Verify alphabetical header ordering
		lines := strings.Split(csvOutputs[0], "\n")
		headers := strings.Split(lines[0], ",")
		expectedHeaders := []string{"age", "created_at", "email", "name", "role"}

		for i, expected := range expectedHeaders {
			if i >= len(headers) || headers[i] != expected {
				t.Errorf("Expected header %d to be '%s', got '%s'", i, expected, headers[i])
			}
		}

		t.Logf("✅ Custom CSV deterministic ordering: consistent across multiple runs")
	})
}
