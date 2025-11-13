package test

import (
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// User model matching your actual data structure
type User struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
}

// TestDateRangeBugFix tests the specific bug reported where date range filtering
// was returning all records instead of filtering properly
func TestDateRangeBugFix(t *testing.T) {
	handler := filter.NewFilter[User](filter.GolangFilteringConfig{})

	// Create test data with dates in November 2024 to November 2025 (similar to your actual data)
	users := []*User{
		{ID: "1", CreatedAt: time.Date(2024, 11, 13, 15, 32, 25, 0, time.UTC), Email: "user1@example.com"},
		{ID: "2", CreatedAt: time.Date(2024, 12, 15, 10, 0, 0, 0, time.UTC), Email: "user2@example.com"},
		{ID: "3", CreatedAt: time.Date(2025, 1, 20, 14, 30, 0, 0, time.UTC), Email: "user3@example.com"},
		{ID: "4", CreatedAt: time.Date(2025, 3, 10, 9, 15, 0, 0, time.UTC), Email: "user4@example.com"},
		{ID: "5", CreatedAt: time.Date(2025, 6, 1, 8, 0, 0, 0, time.UTC), Email: "user5@example.com"},
		{ID: "6", CreatedAt: time.Date(2025, 11, 13, 16, 8, 47, 0, time.UTC), Email: "user6@example.com"},
	}

	// Test the exact filter that was failing: June 3-5, 2025
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "created_at",
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
				Value: filter.Range{
					From: "2025-06-03T16:00:00.000Z",
					To:   "2025-06-05T16:00:00.000Z",
				},
			},
		},
	}

	result, err := handler.DataQuery(users, filterRoot, 0, 100)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should return 0 users since none are in the June 3-5 range
	if result.TotalSize != 0 {
		t.Errorf("Expected 0 users in June 3-5 range, got %d", result.TotalSize)
	}

	// Test a range that should match some data
	filterRoot2 := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "created_at",
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
				Value: filter.Range{
					From: "2025-11-12",
					To:   "2025-11-14",
				},
			},
		},
	}

	result2, err := handler.DataQuery(users, filterRoot2, 0, 100)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should return 1 user (user6 created on 2025-11-13)
	if result2.TotalSize != 1 {
		t.Errorf("Expected 1 user in Nov 12-14 range, got %d", result2.TotalSize)
	}

	// Test a range that should match multiple records
	filterRoot3 := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "created_at",
				Mode:     filter.ModeRange,
				DataType: filter.DataTypeDate,
				Value: filter.Range{
					From: "2024-11-01",
					To:   "2025-01-31",
				},
			},
		},
	}

	result3, err := handler.DataQuery(users, filterRoot3, 0, 100)
	if err != nil {
		t.Fatalf("Error filtering: %v", err)
	}

	// Should return 3 users (user1, user2, user3)
	if result3.TotalSize != 3 {
		t.Errorf("Expected 3 users in Nov 2024 - Jan 2025 range, got %d", result3.TotalSize)
	}
}
