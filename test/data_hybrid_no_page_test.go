package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDataHybridNoPage tests the DataHybridNoPage function that returns results without pagination
func TestDataHybridNoPage(t *testing.T) {
	db := setupOrderByDB(t)
	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Test 1: Small dataset - should use DataQueryNoPage (in-memory)
	smallThreshold := 1000 // Our test DB has only 6 users, so this will trigger in-memory processing
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	results, err := handler.DataHybridNoPage(db, smallThreshold, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataHybridNoPage with small threshold: %v", err)
	}

	if len(results) != 6 {
		t.Errorf("Expected 6 users with small threshold, got %d", len(results))
	}

	// Verify sorting
	for i := 1; i < len(results); i++ {
		if results[i-1].Name > results[i].Name {
			t.Errorf("Names not in ascending order: '%s' should come before '%s'",
				results[i-1].Name, results[i].Name)
		}
	}

	t.Logf("✅ DataHybridNoPage with small threshold (in-memory): %d users returned", len(results))

	// Test 2: Large dataset threshold - should use DataGormNoPage (database)
	smallThreshold = 1 // Force database processing by setting a very low threshold
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
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.SortOrderDesc},
		},
	}

	results, err = handler.DataHybridNoPage(db, smallThreshold, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataHybridNoPage with small threshold (database mode): %v", err)
	}

	// Verify all users are active
	for _, user := range results {
		if !user.Active {
			t.Errorf("Expected all users to be active, found inactive user: %s", user.Name)
		}
	}

	// Verify age sorting (descending)
	for i := 1; i < len(results); i++ {
		if results[i-1].Age < results[i].Age {
			t.Errorf("Ages not in descending order: %d should come before %d",
				results[i-1].Age, results[i].Age)
		}
	}

	t.Logf("✅ DataHybridNoPage with small threshold (database): %d active users returned", len(results))

	// Test 3: With preloads
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

	results, err = handler.DataHybridNoPage(db, 1000, filterRoot) // Use in-memory mode
	if err != nil {
		t.Fatalf("Failed to execute DataHybridNoPage with preloads: %v", err)
	}

	t.Logf("✅ DataHybridNoPage with preloads (Engineering department):")
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

	// Test 4: With preset conditions
	presetDB := db.Where("age > ?", 25)
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
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.SortOrderAsc},
		},
	}

	results, err = handler.DataHybridNoPage(presetDB, 1000, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataHybridNoPage with preset conditions: %v", err)
	}

	t.Logf("✅ DataHybridNoPage with preset conditions (age > 25, active = true):")
	for i, user := range results {
		if user.Age <= 25 {
			t.Errorf("Expected user %s to have age > 25, got %d", user.Name, user.Age)
		}
		if !user.Active {
			t.Errorf("Expected user %s to be active", user.Name)
		}
		t.Logf("  %d: %s (age %d, active: %t)", i+1, user.Name, user.Age, user.Active)
	}

	t.Logf("✅ DataHybridNoPage tests completed successfully")
}

// TestDataHybridNoPageComparison compares results between different approaches
func TestDataHybridNoPageComparison(t *testing.T) {
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
				Value:    27,
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	// Get results using DataHybridNoPage (in-memory mode)
	hybridInMemoryResults, err := handler.DataHybridNoPage(db, 1000, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataHybridNoPage (in-memory): %v", err)
	}

	// Get results using DataHybridNoPage (database mode)
	hybridDbResults, err := handler.DataHybridNoPage(db, 1, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataHybridNoPage (database): %v", err)
	}

	// Get results using DataGormNoPage directly
	gormResults, err := handler.DataGormNoPage(db, filterRoot)
	if err != nil {
		t.Fatalf("Failed to execute DataGormNoPage: %v", err)
	}

	// Compare result counts
	if len(hybridInMemoryResults) != len(hybridDbResults) {
		t.Errorf("Hybrid in-memory (%d) and database (%d) results count mismatch",
			len(hybridInMemoryResults), len(hybridDbResults))
	}

	if len(hybridDbResults) != len(gormResults) {
		t.Errorf("Hybrid database (%d) and GORM (%d) results count mismatch",
			len(hybridDbResults), len(gormResults))
	}

	// Compare individual records
	for i := 0; i < len(hybridInMemoryResults) && i < len(gormResults); i++ {
		hybridUser := hybridInMemoryResults[i]
		gormUser := gormResults[i]

		if hybridUser.ID != gormUser.ID {
			t.Errorf("User ID mismatch at position %d: Hybrid ID=%d, GORM ID=%d",
				i, hybridUser.ID, gormUser.ID)
		}

		if hybridUser.Name != gormUser.Name {
			t.Errorf("User name mismatch at position %d: Hybrid='%s', GORM='%s'",
				i, hybridUser.Name, gormUser.Name)
		}
	}

	t.Logf("✅ DataHybridNoPage comparison results:")
	t.Logf("  Hybrid (in-memory): %d records", len(hybridInMemoryResults))
	t.Logf("  Hybrid (database): %d records", len(hybridDbResults))
	t.Logf("  GORM direct: %d records", len(gormResults))
	t.Logf("  All results match: %t",
		len(hybridInMemoryResults) == len(hybridDbResults) &&
			len(hybridDbResults) == len(gormResults))
}

// TestDataHybridNoPageThresholdBehavior tests threshold-based decision making
func TestDataHybridNoPageThresholdBehavior(t *testing.T) {
	// Create a separate database with more data to test thresholds
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&OrderByTestDept{}, &OrderByTestUser{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Insert departments
	departments := []OrderByTestDept{
		{ID: 1, Name: "Engineering", Code: "ENG"},
		{ID: 2, Name: "Marketing", Code: "MKT"},
		{ID: 3, Name: "Sales", Code: "SAL"},
	}
	for _, dept := range departments {
		if err := db.Create(&dept).Error; err != nil {
			t.Fatalf("Failed to create department: %v", err)
		}
	}

	// Insert a larger number of users for threshold testing
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 1; i <= 50; i++ {
		user := OrderByTestUser{
			ID:           uint(i),
			Name:         fmt.Sprintf("User %02d", i),
			Age:          20 + (i % 30), // Ages from 20 to 49
			Salary:       50000.00 + float64(i*1000),
			Active:       i%2 == 0, // Alternating active status
			DepartmentID: uint((i % 3) + 1),
			CreatedAt:    baseTime.Add(time.Duration(i) * time.Hour),
			UpdatedAt:    baseTime.Add(time.Duration(i) * time.Hour),
		}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	maxDepth := 3
	handler := filter.NewFilter[OrderByTestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

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
	}

	// Test with threshold above dataset size (should use in-memory)
	highThreshold := 100
	start := time.Now()
	resultsInMemory, err := handler.DataHybridNoPage(db, highThreshold, filterRoot)
	inMemoryDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to execute DataHybridNoPage with high threshold: %v", err)
	}

	// Test with threshold below dataset size (should use database)
	lowThreshold := 10
	start = time.Now()
	resultsDatabase, err := handler.DataHybridNoPage(db, lowThreshold, filterRoot)
	databaseDuration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to execute DataHybridNoPage with low threshold: %v", err)
	}

	// Both should return the same results
	if len(resultsInMemory) != len(resultsDatabase) {
		t.Errorf("Result count mismatch: in-memory %d vs database %d",
			len(resultsInMemory), len(resultsDatabase))
	}

	t.Logf("✅ DataHybridNoPage threshold behavior:")
	t.Logf("  High threshold (%d): %d results in %v (in-memory mode)",
		highThreshold, len(resultsInMemory), inMemoryDuration)
	t.Logf("  Low threshold (%d): %d results in %v (database mode)",
		lowThreshold, len(resultsDatabase), databaseDuration)
	t.Logf("  Results consistency: %t", len(resultsInMemory) == len(resultsDatabase))
}
