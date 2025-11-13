package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// TestPaginationPageIndex0To5 tests pagination from page index 0 to 5 with both DataQuery and DataGorm
func TestPaginationPageIndex0To5(t *testing.T) {
	// Setup test data - create enough records to span multiple pages
	maxDepth := 3
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{
		MaxDepth: &maxDepth,
	})

	// Create 25 test users to ensure we have data across multiple pages
	users := []*TestUser{
		{ID: 1, Name: "Alice Johnson", Age: 30, IsActive: true, CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: 2, Name: "Bob Smith", Age: 25, IsActive: true, CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		{ID: 3, Name: "Charlie Brown", Age: 35, IsActive: false, CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
		{ID: 4, Name: "Diana Prince", Age: 28, IsActive: true, CreatedAt: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)},
		{ID: 5, Name: "Eve Adams", Age: 32, IsActive: true, CreatedAt: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 6, Name: "Frank Miller", Age: 27, IsActive: false, CreatedAt: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)},
		{ID: 7, Name: "Grace Kelly", Age: 29, IsActive: true, CreatedAt: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)},
		{ID: 8, Name: "Henry Ford", Age: 45, IsActive: true, CreatedAt: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)},
		{ID: 9, Name: "Ivy League", Age: 22, IsActive: false, CreatedAt: time.Date(2024, 1, 9, 0, 0, 0, 0, time.UTC)},
		{ID: 10, Name: "Jack Black", Age: 38, IsActive: true, CreatedAt: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)},
		{ID: 11, Name: "Kate Moss", Age: 31, IsActive: true, CreatedAt: time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)},
		{ID: 12, Name: "Leo DiCaprio", Age: 49, IsActive: false, CreatedAt: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)},
		{ID: 13, Name: "Mary Jane", Age: 26, IsActive: true, CreatedAt: time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC)},
		{ID: 14, Name: "Nick Fury", Age: 55, IsActive: true, CreatedAt: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC)},
		{ID: 15, Name: "Olivia Newton", Age: 33, IsActive: false, CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		{ID: 16, Name: "Peter Parker", Age: 24, IsActive: true, CreatedAt: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)},
		{ID: 17, Name: "Queen Elizabeth", Age: 96, IsActive: false, CreatedAt: time.Date(2024, 1, 17, 0, 0, 0, 0, time.UTC)},
		{ID: 18, Name: "Robert Downey", Age: 58, IsActive: true, CreatedAt: time.Date(2024, 1, 18, 0, 0, 0, 0, time.UTC)},
		{ID: 19, Name: "Sarah Connor", Age: 41, IsActive: true, CreatedAt: time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC)},
		{ID: 20, Name: "Tony Stark", Age: 53, IsActive: false, CreatedAt: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)},
		{ID: 21, Name: "Uma Thurman", Age: 54, IsActive: true, CreatedAt: time.Date(2024, 1, 21, 0, 0, 0, 0, time.UTC)},
		{ID: 22, Name: "Vincent Vega", Age: 36, IsActive: true, CreatedAt: time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)},
		{ID: 23, Name: "Wonder Woman", Age: 30, IsActive: false, CreatedAt: time.Date(2024, 1, 23, 0, 0, 0, 0, time.UTC)},
		{ID: 24, Name: "Xavier Professor", Age: 60, IsActive: true, CreatedAt: time.Date(2024, 1, 24, 0, 0, 0, 0, time.UTC)},
		{ID: 25, Name: "Yoda Master", Age: 900, IsActive: true, CreatedAt: time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC)},
	}

	// Filter configuration - get active users only, sorted by ID
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
		SortFields: []filter.SortField{
			{Field: "id", Order: filter.SortOrderAsc},
		},
	}

	pageSize := 3 // Use small page size to test multiple pages

	t.Logf("🧪 Testing Pagination from Page Index 0 to 5 (PageSize: %d)", pageSize)
	t.Logf("📊 Total users: %d, Expected active users: %d", len(users), countActiveUsers(users))

	// Test DataQuery pagination for pages 0-5
	t.Run("DataQuery_Pagination_Pages_0_to_5", func(t *testing.T) {
		var totalRecordsFound int
		var lastNonEmptyPage int

		for pageIndex := 0; pageIndex <= 5; pageIndex++ {
			result, err := handler.DataQuery(users, filterRoot, pageIndex, pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed for page %d: %v", pageIndex, err)
			}

			// Validate pagination result structure
			if result.PageSize != pageSize {
				t.Errorf("Page %d: Expected page size %d, got %d", pageIndex, pageSize, result.PageSize)
			}

			// PageIndex should remain as-is for 0-based indexing
			if result.PageIndex != pageIndex {
				t.Errorf("Page %d: Expected page index %d (0-based), got %d", pageIndex, pageIndex, result.PageIndex)
			}

			// Log page details
			t.Logf("📄 Page %d: Returned %d records (PageIndex: %d, TotalSize: %d, TotalPage: %d)",
				pageIndex, len(result.Data), result.PageIndex, result.TotalSize, result.TotalPage)

			if len(result.Data) > 0 {
				lastNonEmptyPage = pageIndex
				totalRecordsFound += len(result.Data)

				// Log first and last record of this page
				firstRecord := result.Data[0]
				lastRecord := result.Data[len(result.Data)-1]
				t.Logf("   📝 First: ID=%d %s, Last: ID=%d %s",
					firstRecord.ID, firstRecord.Name, lastRecord.ID, lastRecord.Name)

				// Verify all records are active
				for i, user := range result.Data {
					if !user.IsActive {
						t.Errorf("Page %d, Record %d: Expected active user, got inactive user ID=%d %s",
							pageIndex, i, user.ID, user.Name)
					}
				}

				// Verify sorting (by ID ascending)
				for i := 1; i < len(result.Data); i++ {
					if result.Data[i-1].ID >= result.Data[i].ID {
						t.Errorf("Page %d: Records not sorted correctly by ID. %d >= %d",
							pageIndex, result.Data[i-1].ID, result.Data[i].ID)
					}
				}
			} else {
				t.Logf("   📭 Empty page")
			}

			// Validate page boundaries
			if pageIndex > 0 && result.TotalPage > 0 && pageIndex <= result.TotalPage {
				// Pages within range should have data unless we're at the end
				if len(result.Data) == 0 && pageIndex <= result.TotalPage {
					t.Logf("   ⚠️  Page %d is empty but within total page range (%d)", pageIndex, result.TotalPage)
				}
			}
		}

		t.Logf("✅ DataQuery Summary: Last non-empty page: %d, Total records found across pages: %d",
			lastNonEmptyPage, totalRecordsFound)
	})

	// Test DataGorm pagination for pages 0-5 (with database)
	t.Run("DataGorm_Pagination_Pages_0_to_5", func(t *testing.T) {
		db := setupTestDB(t)
		var totalRecordsFound int
		var lastNonEmptyPage int

		for pageIndex := 0; pageIndex <= 5; pageIndex++ {
			result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
			if err != nil {
				t.Fatalf("DataGorm failed for page %d: %v", pageIndex, err)
			}

			// Validate pagination result structure
			if result.PageSize != pageSize {
				t.Errorf("Page %d: Expected page size %d, got %d", pageIndex, pageSize, result.PageSize)
			}

			// PageIndex should remain as-is for 0-based indexing
			if result.PageIndex != pageIndex {
				t.Errorf("Page %d: Expected page index %d (0-based), got %d", pageIndex, pageIndex, result.PageIndex)
			}

			// Log page details
			t.Logf("📄 Page %d: Returned %d records (PageIndex: %d, TotalSize: %d, TotalPage: %d)",
				pageIndex, len(result.Data), result.PageIndex, result.TotalSize, result.TotalPage)

			if len(result.Data) > 0 {
				lastNonEmptyPage = pageIndex
				totalRecordsFound += len(result.Data)

				// Log first and last record of this page
				firstRecord := result.Data[0]
				lastRecord := result.Data[len(result.Data)-1]
				t.Logf("   📝 First: ID=%d %s, Last: ID=%d %s",
					firstRecord.ID, firstRecord.Name, lastRecord.ID, lastRecord.Name)

				// Verify all records are active
				for i, user := range result.Data {
					if !user.IsActive {
						t.Errorf("Page %d, Record %d: Expected active user, got inactive user ID=%d %s",
							pageIndex, i, user.ID, user.Name)
					}
				}

				// Verify sorting (by ID ascending)
				for i := 1; i < len(result.Data); i++ {
					if result.Data[i-1].ID >= result.Data[i].ID {
						t.Errorf("Page %d: Records not sorted correctly by ID. %d >= %d",
							pageIndex, result.Data[i-1].ID, result.Data[i].ID)
					}
				}
			} else {
				t.Logf("   📭 Empty page")
			}
		}

		t.Logf("✅ DataGorm Summary: Last non-empty page: %d, Total records found across pages: %d",
			lastNonEmptyPage, totalRecordsFound)
	})

	// Test edge cases
	t.Run("Edge_Cases", func(t *testing.T) {
		// Test page 0 specifically
		t.Run("Page_0_Behavior", func(t *testing.T) {
			result, err := handler.DataQuery(users, filterRoot, 0, pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed for page 0: %v", err)
			}
			t.Logf("🔍 Page 0 behavior - PageIndex: %d, Records: %d", result.PageIndex, len(result.Data))
		})

		// Test negative page index
		t.Run("Negative_Page_Index", func(t *testing.T) {
			result, err := handler.DataQuery(users, filterRoot, -1, pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed for page -1: %v", err)
			}
			t.Logf("🔍 Page -1 behavior - PageIndex: %d, Records: %d", result.PageIndex, len(result.Data))
		})

		// Test very large page index
		t.Run("Large_Page_Index", func(t *testing.T) {
			result, err := handler.DataQuery(users, filterRoot, 100, pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed for page 100: %v", err)
			}
			t.Logf("🔍 Page 100 behavior - PageIndex: %d, Records: %d", result.PageIndex, len(result.Data))

			if len(result.Data) != 0 {
				t.Errorf("Expected page 100 to be empty, got %d records", len(result.Data))
			}
		})
	})

	// Test consistency between pages
	t.Run("Page_Consistency", func(t *testing.T) {
		// Collect all records from pages 0-4 and verify no duplicates
		allRecords := make(map[uint]*TestUser)

		for pageIndex := 0; pageIndex <= 4; pageIndex++ {
			result, err := handler.DataQuery(users, filterRoot, pageIndex, pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed for page %d: %v", pageIndex, err)
			}

			for _, user := range result.Data {
				if existingUser, exists := allRecords[user.ID]; exists {
					t.Errorf("Duplicate user ID %d found on multiple pages. First: %s, Second: %s",
						user.ID, existingUser.Name, user.Name)
				}
				allRecords[user.ID] = user
			}
		}

		t.Logf("✅ Collected %d unique records across pages 0-4", len(allRecords))
	})
}

// TestPaginationBoundaryConditions tests specific boundary conditions for pagination
func TestPaginationBoundaryConditions(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	// Create exactly 10 users for predictable pagination testing
	users := make([]*TestUser, 10)
	for i := 0; i < 10; i++ {
		users[i] = &TestUser{
			ID:        uint(i + 1),
			Name:      fmt.Sprintf("User_%02d", i+1),
			Age:       25 + i,
			IsActive:  true,
			CreatedAt: time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
		}
	}

	filterRoot := filter.Root{Logic: filter.LogicAnd}

	testCases := []struct {
		name        string
		pageIndex   int
		pageSize    int
		expectCount int
		expectEmpty bool
		description string
	}{
		{"Page_0_Size_3", 0, 3, 3, false, "Page 0 (first page) with 3 records"},
		{"Page_1_Size_3", 1, 3, 3, false, "Page 1 (second page) with 3 records"},
		{"Page_2_Size_3", 2, 3, 3, false, "Page 2 (third page) with 3 records"},
		{"Page_3_Size_3", 3, 3, 1, false, "Page 3 (fourth page) with remaining 1 record"},
		{"Page_4_Size_3", 4, 3, 0, true, "Page 4 should be empty"},
		{"Page_5_Size_3", 5, 3, 0, true, "Page 5 should be empty"},
		{"Page_0_Size_10", 0, 10, 10, false, "Page 0 (first page) with all records"},
		{"Page_1_Size_10", 1, 10, 0, true, "Page 1 should be empty when all fit on page 0"},
		{"Page_0_Size_15", 0, 15, 10, false, "Page 0 with page size larger than dataset"},
		{"Page_0_Size_0", 0, 0, 10, false, "Page 0 with zero page size should use default (30), fit all records"},
		{"Page_1_Size_0", 1, 0, 0, true, "Page 1 with zero page size should be empty since all fit on page 0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := handler.DataQuery(users, filterRoot, tc.pageIndex, tc.pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed: %v", err)
			}

			actualCount := len(result.Data)
			isEmpty := actualCount == 0

			t.Logf("📊 %s: PageIndex=%d, PageSize=%d -> ActualCount=%d, TotalSize=%d, TotalPage=%d",
				tc.description, result.PageIndex, result.PageSize, actualCount, result.TotalSize, result.TotalPage)

			if tc.expectEmpty && !isEmpty {
				t.Errorf("Expected empty page, got %d records", actualCount)
			}

			if !tc.expectEmpty && isEmpty {
				t.Errorf("Expected non-empty page, got 0 records")
			}

			if !tc.expectEmpty && tc.expectCount > 0 && actualCount != tc.expectCount {
				t.Errorf("Expected %d records, got %d", tc.expectCount, actualCount)
			}

			// Verify TotalSize is consistent
			if result.TotalSize != 10 {
				t.Errorf("Expected TotalSize=10, got %d", result.TotalSize)
			}
		})
	}
}

// Helper function to count active users
func countActiveUsers(users []*TestUser) int {
	count := 0
	for _, user := range users {
		if user.IsActive {
			count++
		}
	}
	return count
}
