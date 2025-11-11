package test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// TestDataGormNoPage tests the DataGormNoPage function that returns results without pagination
func TestDataGormNoPage(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Test 1: Get all users without any filters
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	results, err := handler.DataGormNoPage(db, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataGormNoPage without filters: %v", err)
	}

	if len(results) != 6 {
		t.Errorf("Expected 6 users without filters, got %d", len(results))
	}

	t.Logf("✅ DataGormNoPage without filters: %d users returned", len(results))

	// Test 2: Filter by active users only
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "active",
				Value:    true,
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeBool,
			},
		},
	}

	results, err = handler.DataGormNoPage(db, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataGormNoPage with active filter: %v", err)
	}

	activeCount := 0
	for _, user := range results {
		if user.Active {
			activeCount++
		}
	}

	if activeCount != len(results) {
		t.Errorf("Expected all results to be active users, but found %d active out of %d total",
			activeCount, len(results))
	}

	t.Logf("✅ DataGormNoPage with active filter: %d active users returned", len(results))

	// Test 3: Filter with ORDER BY
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    30,
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
			{Field: "age", Order: filter.SortOrderDesc},
		},
	}

	results, err = handler.DataGormNoPage(db, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataGormNoPage with age filter and sorting: %v", err)
	}

	t.Logf("✅ DataGormNoPage with age >= 30 and ORDER BY name ASC, age DESC:")
	for i, user := range results {
		if user.Age < 30 {
			t.Errorf("Expected all users to have age >= 30, but user %s has age %d", user.Name, user.Age)
		}
		t.Logf("  %d: %s (age %d)", i+1, user.Name, user.Age)
	}

	// Verify sorting
	for i := 1; i < len(results); i++ {
		prev := results[i-1]
		curr := results[i]

		// Check name ordering (should be ascending)
		if prev.Name > curr.Name {
			t.Errorf("Names not in ascending order: '%s' should come before '%s'", prev.Name, curr.Name)
		} else if prev.Name == curr.Name && prev.Age < curr.Age {
			// Same name, check age ordering (should be descending)
			t.Errorf("For name '%s': ages not in descending order: %d should come before %d",
				prev.Name, prev.Age, curr.Age)
		}
	}

	// Test 4: With nested field filtering and preloads
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "department.name",
				Value:    "Engineering",
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
		Preload: []string{"Department"},
	}

	results, err = handler.DataGormNoPage(db, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataGormNoPage with nested field filter: %v", err)
	}

	t.Logf("✅ DataGormNoPage with department filter (Engineering):")
	for i, user := range results {
		if user.Department == nil {
			t.Errorf("Department not preloaded for user %s", user.Name)
			continue
		}
		if user.Department.Name != "Engineering" {
			t.Errorf("Expected user %s to be in Engineering department, got %s",
				user.Name, user.Department.Name)
		}
		t.Logf("  %d: %s (%s department)", i+1, user.Name, user.Department.Name)
	}

	t.Logf("✅ DataGormNoPage tests completed successfully")
}

// TestDataGormNoPageWithPreset tests the DataGormNoPageWithPreset convenience function
func TestDataGormNoPageWithPreset(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Create a preset condition struct
	type UserPresetConditions struct {
		DepartmentID uint `gorm:"column:department_id"`
	}

	presetConditions := &UserPresetConditions{
		DepartmentID: 1, // Engineering department
	}

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "active",
				Value:    true,
				Mode:     filter.ModeEqual,
				DataType: filter.DataTypeBool,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
		Preload: []string{"Department"},
	}

	results, err := handler.DataGormNoPageWithPreset(db, presetConditions, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataGormNoPageWithPreset: %v", err)
	}

	t.Logf("✅ DataGormNoPageWithPreset (department_id=1, active=true):")
	for i, user := range results {
		if user.DepartmentID != 1 {
			t.Errorf("Expected user %s to have department_id=1, got %d", user.Name, user.DepartmentID)
		}
		if !user.Active {
			t.Errorf("Expected user %s to be active", user.Name)
		}
		if user.Department != nil {
			t.Logf("  %d: %s (active: %t, department: %s)",
				i+1, user.Name, user.Active, user.Department.Name)
		} else {
			t.Logf("  %d: %s (active: %t, department_id: %d)",
				i+1, user.Name, user.Active, user.DepartmentID)
		}
	}

	// Test with nil preset conditions (should work like regular DataGormNoPage)
	results2, err := handler.DataGormNoPageWithPreset(db, nil, filter.Root{Logic: filter.LogicAnd})
	if err != nil {
		t.Fatalf("Failed to execute DataGormNoPageWithPreset with nil conditions: %v", err)
	}

	if len(results2) != 6 {
		t.Errorf("Expected 6 users with nil preset conditions, got %d", len(results2))
	}

	t.Logf("✅ DataGormNoPageWithPreset with nil conditions: %d users returned", len(results2))
	t.Logf("✅ DataGormNoPageWithPreset tests completed successfully")
}

// TestDataGormNoPageComparison compares results between DataGorm and DataGormNoPage
func TestDataGormNoPageComparison(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "age",
				Value:    25,
				Mode:     filter.ModeGT,
				DataType: filter.DataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	// Get results with pagination (large page size to get all results)
	paginatedResult, err := handler.DataGorm(db, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to execute DataGorm: %v", err)
	}

	// Get results without pagination
	noPageResults, err := handler.DataGormNoPage(db, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataGormNoPage: %v", err)
	}

	// Compare results
	if len(paginatedResult.Data) != len(noPageResults) {
		t.Errorf("Result count mismatch: DataGorm returned %d, DataGormNoPage returned %d",
			len(paginatedResult.Data), len(noPageResults))
	}

	// Compare each record
	for i := 0; i < len(paginatedResult.Data) && i < len(noPageResults); i++ {
		paginatedUser := paginatedResult.Data[i]
		noPageUser := noPageResults[i]

		if paginatedUser.ID != noPageUser.ID {
			t.Errorf("User ID mismatch at position %d: DataGorm ID=%d, DataGormNoPage ID=%d",
				i, paginatedUser.ID, noPageUser.ID)
		}

		if paginatedUser.Name != noPageUser.Name {
			t.Errorf("User name mismatch at position %d: DataGorm='%s', DataGormNoPage='%s'",
				i, paginatedUser.Name, noPageUser.Name)
		}
	}

	t.Logf("✅ DataGorm vs DataGormNoPage comparison:")
	t.Logf("  DataGorm (paginated): %d records (TotalSize: %d, TotalPage: %d)",
		len(paginatedResult.Data), paginatedResult.TotalSize, paginatedResult.TotalPage)
	t.Logf("  DataGormNoPage: %d records", len(noPageResults))
	t.Logf("  Results match: %t", len(paginatedResult.Data) == len(noPageResults))
}
