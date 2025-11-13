package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// TestPaginationStressTest performs comprehensive pagination stress testing
func TestPaginationStressTest(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	t.Run("1_Item_Per_Page_20_Pages", func(t *testing.T) {
		// Create exactly 20 users for 1 item per page testing
		users := make([]*TestUser, 20)
		for i := 0; i < 20; i++ {
			users[i] = &TestUser{
				ID:        uint(i + 1),
				Name:      fmt.Sprintf("User_%02d", i+1),
				Age:       25 + i,
				IsActive:  true,
				CreatedAt: time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			}
		}

		filterRoot := filter.Root{Logic: filter.LogicAnd}
		pageSize := 1

		t.Logf("🧪 Testing 1 item per page across 20 pages (20 total items)")

		// Test pages 0-19 (should all have data) and a few beyond
		for pageIndex := 0; pageIndex < 25; pageIndex++ {
			result, err := handler.DataQuery(users, filterRoot, pageIndex, pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed for page %d: %v", pageIndex, err)
			}

			expectedCount := 1
			if pageIndex >= 20 {
				expectedCount = 0 // Pages beyond data should be empty
			}

			actualCount := len(result.Data)

			t.Logf("📄 Page %d: Expected %d, Got %d items (TotalSize: %d, TotalPage: %d)",
				pageIndex, expectedCount, actualCount, result.TotalSize, result.TotalPage)

			if actualCount != expectedCount {
				t.Errorf("Page %d: Expected %d items, got %d", pageIndex, expectedCount, actualCount)
			}

			if result.TotalSize != 20 {
				t.Errorf("Page %d: Expected TotalSize=20, got %d", pageIndex, result.TotalSize)
			}

			if result.TotalPage != 20 {
				t.Errorf("Page %d: Expected TotalPage=20, got %d", pageIndex, result.TotalPage)
			}

			if result.PageIndex != pageIndex {
				t.Errorf("Page %d: Expected PageIndex=%d, got %d", pageIndex, pageIndex, result.PageIndex)
			}

			// Verify correct item is returned
			if actualCount > 0 {
				expectedID := uint(pageIndex + 1)
				if result.Data[0].ID != expectedID {
					t.Errorf("Page %d: Expected user ID %d, got %d", pageIndex, expectedID, result.Data[0].ID)
				}
			}
		}
	})

	t.Run("Odd_Numbers_Test_13_Items_3_Per_Page", func(t *testing.T) {
		// Create 13 users (odd number) with 3 items per page
		users := make([]*TestUser, 13)
		for i := 0; i < 13; i++ {
			users[i] = &TestUser{
				ID:        uint(i + 1),
				Name:      fmt.Sprintf("OddUser_%02d", i+1),
				Age:       30 + i,
				IsActive:  true,
				CreatedAt: time.Date(2024, 2, i+1, 0, 0, 0, 0, time.UTC),
			}
		}

		filterRoot := filter.Root{Logic: filter.LogicAnd}
		pageSize := 3

		// Expected distribution: Page 0: 3 items, Page 1: 3 items, Page 2: 3 items, Page 3: 3 items, Page 4: 1 item
		expectedCounts := []int{3, 3, 3, 3, 1, 0, 0} // Pages 0-6

		t.Logf("🧪 Testing 13 items with 3 per page (odd numbers)")

		for pageIndex, expectedCount := range expectedCounts {
			result, err := handler.DataQuery(users, filterRoot, pageIndex, pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed for page %d: %v", pageIndex, err)
			}

			actualCount := len(result.Data)

			t.Logf("📄 Page %d: Expected %d, Got %d items (TotalSize: %d, TotalPage: %d)",
				pageIndex, expectedCount, actualCount, result.TotalSize, result.TotalPage)

			if actualCount != expectedCount {
				t.Errorf("Page %d: Expected %d items, got %d", pageIndex, expectedCount, actualCount)
			}

			if result.TotalSize != 13 {
				t.Errorf("Page %d: Expected TotalSize=13, got %d", pageIndex, result.TotalSize)
			}

			expectedTotalPages := 5 // ceil(13/3) = 5
			if result.TotalPage != expectedTotalPages {
				t.Errorf("Page %d: Expected TotalPage=%d, got %d", pageIndex, expectedTotalPages, result.TotalPage)
			}

			// Verify correct items are returned
			if actualCount > 0 {
				startID := uint(pageIndex*pageSize + 1)
				for i, user := range result.Data {
					expectedID := startID + uint(i)
					if user.ID != expectedID {
						t.Errorf("Page %d, Position %d: Expected user ID %d, got %d",
							pageIndex, i, expectedID, user.ID)
					}
				}
			}
		}
	})

	t.Run("Prime_Numbers_Test_17_Items_5_Per_Page", func(t *testing.T) {
		// Create 17 users (prime number) with 5 items per page
		users := make([]*TestUser, 17)
		for i := 0; i < 17; i++ {
			users[i] = &TestUser{
				ID:        uint(i + 1),
				Name:      fmt.Sprintf("PrimeUser_%02d", i+1),
				Age:       35 + i,
				IsActive:  true,
				CreatedAt: time.Date(2024, 3, i+1, 0, 0, 0, 0, time.UTC),
			}
		}

		filterRoot := filter.Root{Logic: filter.LogicAnd}
		pageSize := 5

		// Expected distribution: Page 0: 5 items, Page 1: 5 items, Page 2: 5 items, Page 3: 2 items
		expectedCounts := []int{5, 5, 5, 2, 0, 0} // Pages 0-5

		t.Logf("🧪 Testing 17 items (prime) with 5 per page")

		for pageIndex, expectedCount := range expectedCounts {
			result, err := handler.DataQuery(users, filterRoot, pageIndex, pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed for page %d: %v", pageIndex, err)
			}

			actualCount := len(result.Data)

			t.Logf("📄 Page %d: Expected %d, Got %d items (TotalSize: %d, TotalPage: %d)",
				pageIndex, expectedCount, actualCount, result.TotalSize, result.TotalPage)

			if actualCount != expectedCount {
				t.Errorf("Page %d: Expected %d items, got %d", pageIndex, expectedCount, actualCount)
			}

			if result.TotalSize != 17 {
				t.Errorf("Page %d: Expected TotalSize=17, got %d", pageIndex, result.TotalSize)
			}

			expectedTotalPages := 4 // ceil(17/5) = 4
			if result.TotalPage != expectedTotalPages {
				t.Errorf("Page %d: Expected TotalPage=%d, got %d", pageIndex, expectedTotalPages, result.TotalPage)
			}
		}
	})

	t.Run("Database_GORM_Stress_Test", func(t *testing.T) {
		db := setupTestDB(t)

		t.Logf("🧪 Testing GORM pagination with database")

		filterRoot := filter.Root{Logic: filter.LogicAnd}
		pageSize := 2

		// Test pages 0-6 to cover all data and beyond
		for pageIndex := 0; pageIndex < 7; pageIndex++ {
			result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
			if err != nil {
				t.Fatalf("DataGorm failed for page %d: %v", pageIndex, err)
			}

			actualCount := len(result.Data)

			t.Logf("📄 GORM Page %d: Got %d items (TotalSize: %d, TotalPage: %d)",
				pageIndex, actualCount, result.TotalSize, result.TotalPage)

			// Verify pagination consistency
			if result.PageIndex != pageIndex {
				t.Errorf("Page %d: Expected PageIndex=%d, got %d", pageIndex, pageIndex, result.PageIndex)
			}

			if result.PageSize != pageSize {
				t.Errorf("Page %d: Expected PageSize=%d, got %d", pageIndex, pageSize, result.PageSize)
			}

			// Check that pages beyond the last page are empty
			if pageIndex >= result.TotalPage && actualCount != 0 {
				t.Errorf("Page %d: Expected empty page beyond TotalPage=%d, got %d items",
					pageIndex, result.TotalPage, actualCount)
			}

			// Verify no duplicates across pages by checking IDs are sequential within pages
			if actualCount > 1 {
				for i := 1; i < len(result.Data); i++ {
					if result.Data[i].ID <= result.Data[i-1].ID {
						t.Errorf("Page %d: Items not properly ordered. ID %d should be > %d",
							pageIndex, result.Data[i].ID, result.Data[i-1].ID)
					}
				}
			}
		}
	})

	t.Run("Filtered_Data_Pagination", func(t *testing.T) {
		// Create mixed active/inactive users to test pagination with filtering
		users := make([]*TestUser, 15)
		for i := 0; i < 15; i++ {
			users[i] = &TestUser{
				ID:        uint(i + 1),
				Name:      fmt.Sprintf("FilterUser_%02d", i+1),
				Age:       25 + i,
				IsActive:  i%2 == 0, // Only even indexes are active (0,2,4,6,8,10,12,14) = 8 active users
				CreatedAt: time.Date(2024, 5, i+1, 0, 0, 0, 0, time.UTC),
			}
		}

		// Filter for active users only
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
		pageSize := 3

		t.Logf("🧪 Testing pagination with filtered data (8 active out of 15 total)")

		// Should have 8 active users: pages 0,1,2 with 3 each, page 3 with 2 (but it's actually page 2)
		expectedCounts := []int{3, 3, 2, 0, 0} // Pages 0-4

		for pageIndex, expectedCount := range expectedCounts {
			result, err := handler.DataQuery(users, filterRoot, pageIndex, pageSize)
			if err != nil {
				t.Fatalf("DataQuery failed for page %d: %v", pageIndex, err)
			}

			actualCount := len(result.Data)

			t.Logf("📄 Filtered Page %d: Expected %d, Got %d items (TotalSize: %d, TotalPage: %d)",
				pageIndex, expectedCount, actualCount, result.TotalSize, result.TotalPage)

			if actualCount != expectedCount {
				t.Errorf("Page %d: Expected %d items, got %d", pageIndex, expectedCount, actualCount)
			}

			if result.TotalSize != 8 { // Only active users
				t.Errorf("Page %d: Expected TotalSize=8 (active users), got %d", pageIndex, result.TotalSize)
			}

			expectedTotalPages := 3 // ceil(8/3) = 3
			if result.TotalPage != expectedTotalPages {
				t.Errorf("Page %d: Expected TotalPage=%d, got %d", pageIndex, expectedTotalPages, result.TotalPage)
			}

			// Verify all returned users are active
			for i, user := range result.Data {
				if !user.IsActive {
					t.Errorf("Page %d, Position %d: Expected active user, got inactive user ID=%d",
						pageIndex, i, user.ID)
				}
			}
		}
	})
}

// TestPaginationSequentialAccess tests sequential page access patterns
func TestPaginationSequentialAccess(t *testing.T) {
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	// Create 23 users (another odd prime-ish number)
	users := make([]*TestUser, 23)
	for i := 0; i < 23; i++ {
		users[i] = &TestUser{
			ID:        uint(i + 1),
			Name:      fmt.Sprintf("SeqUser_%03d", i+1),
			Age:       20 + i,
			IsActive:  true,
			CreatedAt: time.Date(2024, 6, i+1, 0, 0, 0, 0, time.UTC),
		}
	}

	filterRoot := filter.Root{Logic: filter.LogicAnd}
	pageSize := 4

	t.Logf("🧪 Testing sequential access pattern: 23 items, 4 per page")

	// Collect all items by iterating through pages sequentially
	allCollectedUsers := make([]*TestUser, 0, 23)
	pageIndex := 0

	for {
		result, err := handler.DataQuery(users, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("DataQuery failed for page %d: %v", pageIndex, err)
		}

		t.Logf("📄 Sequential Page %d: Got %d items", pageIndex, len(result.Data))

		if len(result.Data) == 0 {
			break // No more data
		}

		allCollectedUsers = append(allCollectedUsers, result.Data...)
		pageIndex++

		// Safety check to prevent infinite loop
		if pageIndex > 10 {
			t.Fatalf("Too many pages, possible infinite loop")
		}
	}

	// Verify we collected all users
	if len(allCollectedUsers) != 23 {
		t.Errorf("Expected to collect 23 users sequentially, got %d", len(allCollectedUsers))
	}

	// Verify no duplicates and correct order
	seenIDs := make(map[uint]bool)
	for i, user := range allCollectedUsers {
		if seenIDs[user.ID] {
			t.Errorf("Duplicate user ID %d found at position %d", user.ID, i)
		}
		seenIDs[user.ID] = true

		expectedID := uint(i + 1)
		if user.ID != expectedID {
			t.Errorf("Position %d: Expected user ID %d, got %d", i, expectedID, user.ID)
		}
	}

	t.Logf("✅ Sequential access collected all %d users correctly", len(allCollectedUsers))
}
