package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// TestDataQueryNoPage tests the DataQueryNoPage function that returns results without pagination
func TestDataQueryNoPage(t *testing.T) {
	maxDepth := 3
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Create test data
	users := []*TestUser{
		{ID: 1, Name: "Alice Johnson", Age: 30, IsActive: true, CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Bob Smith", Age: 25, IsActive: true, CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Charlie Brown", Age: 35, IsActive: false, CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Diana Prince", Age: 28, IsActive: true, CreatedAt: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Name: "Eve Adams", Age: 32, IsActive: true, CreatedAt: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 6, Name: "Frank Wilson", Age: 27, IsActive: false, CreatedAt: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)},
	}

	// Test 1: Get all users without any filters
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	results, err := handler.DataQueryNoPage(users, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataQueryNoPage without filters: %v", err)
	}

	if len(results) != 6 {
		t.Errorf("Expected 6 users without filters, got %d", len(results))
	}

	t.Logf("✅ DataQueryNoPage without filters: %d users returned", len(results))

	// Test 2: Filter by active users only
	filterRoot = filter.Root{
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

	results, err = handler.DataQueryNoPage(users, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataQueryNoPage with active filter: %v", err)
	}

	activeCount := 0
	for _, user := range results {
		if user.IsActive {
			activeCount++
		}
	}

	if activeCount != len(results) {
		t.Errorf("Expected all results to be active users, but found %d active out of %d total",
			activeCount, len(results))
	}

	t.Logf("✅ DataQueryNoPage with active filter: %d active users returned", len(results))

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

	results, err = handler.DataQueryNoPage(users, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataQueryNoPage with age filter and sorting: %v", err)
	}

	t.Logf("✅ DataQueryNoPage with age >= 30 and ORDER BY name ASC, age DESC:")
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

	// Test 4: Text filtering with contains
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "son",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	results, err = handler.DataQueryNoPage(users, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataQueryNoPage with name contains filter: %v", err)
	}

	t.Logf("✅ DataQueryNoPage with name contains 'son':")
	for i, user := range results {
		t.Logf("  %d: %s", i+1, user.Name)
	}

	// Test 5: Empty data slice
	emptyResults, err := handler.DataQueryNoPage([]*TestUser{}, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataQueryNoPage with empty data: %v", err)
	}

	if len(emptyResults) != 0 {
		t.Errorf("Expected empty results with empty data, got %d", len(emptyResults))
	}

	t.Logf("✅ DataQueryNoPage with empty data: %d users returned", len(emptyResults))
	t.Logf("✅ DataQueryNoPage tests completed successfully")
}

// TestDataQueryNoPageComparison compares results between DataQuery and DataQueryNoPage
func TestDataQueryNoPageComparison(t *testing.T) {
	maxDepth := 3
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Create test data
	users := []*TestUser{
		{ID: 1, Name: "Alice Johnson", Age: 30, IsActive: true, CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Bob Smith", Age: 25, IsActive: true, CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Charlie Brown", Age: 35, IsActive: false, CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Diana Prince", Age: 28, IsActive: true, CreatedAt: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Name: "Eve Adams", Age: 32, IsActive: true, CreatedAt: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 6, Name: "Frank Wilson", Age: 27, IsActive: false, CreatedAt: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)},
	}

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
	paginatedResult, err := handler.DataQuery(users, filterRoot, 1, 100)
	if err != nil {
		t.Fatalf("Failed to execute DataQuery: %v", err)
	}

	// Get results without pagination
	noPageResults, err := handler.DataQueryNoPage(users, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataQueryNoPage: %v", err)
	}

	// Compare results
	if len(paginatedResult.Data) != len(noPageResults) {
		t.Errorf("Result count mismatch: DataQuery returned %d, DataQueryNoPage returned %d",
			len(paginatedResult.Data), len(noPageResults))
	}

	// Compare each record
	for i := 0; i < len(paginatedResult.Data) && i < len(noPageResults); i++ {
		paginatedUser := paginatedResult.Data[i]
		noPageUser := noPageResults[i]

		if paginatedUser.ID != noPageUser.ID {
			t.Errorf("User ID mismatch at position %d: DataQuery ID=%d, DataQueryNoPage ID=%d",
				i, paginatedUser.ID, noPageUser.ID)
		}

		if paginatedUser.Name != noPageUser.Name {
			t.Errorf("User name mismatch at position %d: DataQuery='%s', DataQueryNoPage='%s'",
				i, paginatedUser.Name, noPageUser.Name)
		}
	}

	t.Logf("✅ DataQuery vs DataQueryNoPage comparison:")
	t.Logf("  DataQuery (paginated): %d records (TotalSize: %d, TotalPage: %d)",
		len(paginatedResult.Data), paginatedResult.TotalSize, paginatedResult.TotalPage)
	t.Logf("  DataQueryNoPage: %d records", len(noPageResults))
	t.Logf("  Results match: %t", len(paginatedResult.Data) == len(noPageResults))
}

// TestDataQueryNoPagePerformance tests performance with different data sizes
func TestDataQueryNoPagePerformance(t *testing.T) {
	maxDepth := 3
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Create larger test data for performance testing
	largeDataSet := make([]*TestUser, 1000)
	for i := 0; i < 1000; i++ {
		largeDataSet[i] = &TestUser{
			ID:        uint(i + 1),
			Name:      fmt.Sprintf("User %d", i+1),
			Age:       20 + (i % 40), // Ages from 20 to 59
			IsActive:  i%2 == 0,      // Alternating active status
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Hour),
		}
	}

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
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.SortOrderAsc},
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	start := time.Now()
	results, err := handler.DataQueryNoPage(largeDataSet, filterRoot)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to execute DataQueryNoPage with large dataset: %v", err)
	}

	t.Logf("✅ DataQueryNoPage performance test:")
	t.Logf("  Dataset size: %d records", len(largeDataSet))
	t.Logf("  Filtered results: %d records", len(results))
	t.Logf("  Processing time: %v", duration)
	t.Logf("  Records per second: %.0f", float64(len(largeDataSet))/duration.Seconds())

	// Verify filtering worked correctly
	for _, user := range results {
		if !user.IsActive {
			t.Errorf("Expected only active users, found inactive user: %s", user.Name)
		}
		if user.Age < 30 {
			t.Errorf("Expected users with age >= 30, found user %s with age %d", user.Name, user.Age)
		}
	}

	// Verify sorting
	for i := 1; i < len(results); i++ {
		prev := results[i-1]
		curr := results[i]

		if prev.Age > curr.Age {
			t.Errorf("Ages not in ascending order: %d should come before %d", prev.Age, curr.Age)
		} else if prev.Age == curr.Age && prev.Name > curr.Name {
			t.Errorf("For age %d: names not in ascending order: '%s' should come before '%s'",
				prev.Age, prev.Name, curr.Name)
		}
	}

	t.Logf("✅ DataQueryNoPage performance test completed successfully")
}
