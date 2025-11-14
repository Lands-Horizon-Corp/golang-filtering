package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestPaginationSample demonstrates pagination with 5 items per page and 18 total items
func TestPaginationSample(t *testing.T) {
	// Create test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&TestUser{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	// Create exactly 18 test users
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	users := make([]*TestUser, 18)
	for i := 0; i < 18; i++ {
		users[i] = &TestUser{
			ID:        uint(i + 1),
			Name:      fmt.Sprintf("User %02d", i+1),
			Email:     fmt.Sprintf("user%02d@example.com", i+1),
			Age:       25 + (i % 20), // Vary ages
			IsActive:  true,          // All active for simplicity
			Role:      "user",
			CreatedAt: baseTime.AddDate(0, 0, i),
		}
	}

	// Insert all users into database
	for _, user := range users {
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
	}

	// Create filter handler
	handler := filter.NewFilter[TestUser](filter.GolangFilteringConfig{})

	// Filter configuration with sorting by ID
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

	pageSize := 5
	totalPages := 4 // 18 items / 5 per page = 3.6 → 4 pages

	t.Logf("🧪 Testing DataGorm Pagination: 18 items, %d per page", pageSize)
	t.Logf("📊 Expected pages: %d", totalPages)

	// Test each page
	for pageIndex := 0; pageIndex < totalPages; pageIndex++ {
		result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
		if err != nil {
			t.Fatalf("DataGorm failed for page %d: %v", pageIndex, err)
		}

		// Log page details
		t.Logf("📄 Page %d: %d records (PageIndex: %d, TotalSize: %d, TotalPage: %d)",
			pageIndex, len(result.Data), result.PageIndex, result.TotalSize, result.TotalPage)

		// Show first and last records on this page
		if len(result.Data) > 0 {
			firstRecord := result.Data[0]
			lastRecord := result.Data[len(result.Data)-1]

			if len(result.Data) == 1 {
				t.Logf("   📝 Record: ID=%d %s", firstRecord.ID, firstRecord.Name)
			} else {
				t.Logf("   📝 First: ID=%d %s, Last: ID=%d %s",
					firstRecord.ID, firstRecord.Name, lastRecord.ID, lastRecord.Name)
			}

			// Show all records on this page
			recordIDs := make([]uint, len(result.Data))
			for i, user := range result.Data {
				recordIDs[i] = user.ID
			}
			t.Logf("   🔢 All IDs on page: %v", recordIDs)
		} else {
			t.Logf("   📭 Empty page")
		}

		// Validate page structure
		if result.PageIndex != pageIndex {
			t.Errorf("Expected PageIndex %d, got %d", pageIndex, result.PageIndex)
		}
		if result.PageSize != pageSize {
			t.Errorf("Expected PageSize %d, got %d", pageSize, result.PageSize)
		}
		if result.TotalSize != 18 {
			t.Errorf("Expected TotalSize 18, got %d", result.TotalSize)
		}

		// Expected record count per page
		var expectedCount int
		switch pageIndex {
		case 0, 1, 2: // Pages 0, 1, 2 should have 5 records each
			expectedCount = 5
		case 3: // Page 3 should have 3 records (18 - 15 = 3)
			expectedCount = 3
		default:
			expectedCount = 0
		}

		if len(result.Data) != expectedCount {
			t.Errorf("Page %d: Expected %d records, got %d", pageIndex, expectedCount, len(result.Data))
		}
	}

	// Test empty page beyond the data
	t.Run("Empty_Page_Beyond_Data", func(t *testing.T) {
		result, err := handler.DataGorm(db, filterRoot, 10, pageSize) // Way beyond data
		if err != nil {
			t.Fatalf("DataGorm failed for page 10: %v", err)
		}

		t.Logf("📄 Page 10: %d records (should be empty)", len(result.Data))
		if len(result.Data) != 0 {
			t.Errorf("Expected page 10 to be empty, got %d records", len(result.Data))
		}
	})

	// Verify no duplicate records across pages
	t.Run("No_Duplicates_Across_Pages", func(t *testing.T) {
		allRecords := make(map[uint]*TestUser)

		for pageIndex := 0; pageIndex < totalPages; pageIndex++ {
			result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)
			if err != nil {
				t.Fatalf("DataGorm failed for page %d: %v", pageIndex, err)
			}

			for _, user := range result.Data {
				if existingUser, exists := allRecords[user.ID]; exists {
					t.Errorf("Duplicate user ID %d found on multiple pages. First: %s, Second: %s",
						user.ID, existingUser.Name, user.Name)
				}
				allRecords[user.ID] = user
			}
		}

		t.Logf("✅ Verified %d unique records across all pages (no duplicates)", len(allRecords))

		if len(allRecords) != 18 {
			t.Errorf("Expected 18 unique records, found %d", len(allRecords))
		}
	})
}
