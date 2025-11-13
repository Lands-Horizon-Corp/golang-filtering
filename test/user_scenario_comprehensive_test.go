package test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// TestUserScenarioBugFix tests the exact scenario reported by the user
// where date range filtering was returning all records instead of filtering properly
func TestUserScenarioBugFix(t *testing.T) {
	handler := filter.NewFilter[User](filter.GolangFilteringConfig{})

	// Create test data similar to user's actual API response
	users := []*User{
		{ID: "d0af469c-0361-4a76-baac-701db5817fdd", CreatedAt: time.Date(2025, 11, 13, 15, 32, 25, 0, time.FixedZone("+08:00", 8*3600)), Email: "user1@example.com"},
		{ID: "user-may", CreatedAt: time.Date(2025, 5, 15, 10, 0, 0, 0, time.UTC), Email: "user2@example.com"},
		{ID: "user-june-4", CreatedAt: time.Date(2025, 6, 4, 12, 30, 0, 0, time.UTC), Email: "user3@example.com"},
		{ID: "user-july", CreatedAt: time.Date(2025, 7, 10, 14, 0, 0, 0, time.UTC), Email: "user4@example.com"},
	}

	// Test 1: The exact failing filter from the user's request
	t.Run("UserExactFailingFilter", func(t *testing.T) {
		filterJSON := `{"filters":[{"field":"created_at","mode":"range","dataType":"date","value":{"from":"2025-06-03T16:00:00.000Z","to":"2025-06-05T16:00:00.000Z"}}],"logic":"AND"}`

		var filterRoot filter.Root
		err := json.Unmarshal([]byte(filterJSON), &filterRoot)
		if err != nil {
			t.Fatalf("Failed to parse user's filter JSON: %v", err)
		}

		result, err := handler.DataQuery(users, filterRoot, 1, 100)
		if err != nil {
			t.Fatalf("Error applying filter: %v", err)
		}

		// Should return exactly 1 user (the June 4th user)
		if result.TotalSize != 1 {
			t.Errorf("Expected 1 user in June 3-5 timestamp range, got %d", result.TotalSize)
		}

		if result.TotalSize > 0 && result.Data[0].ID != "user-june-4" {
			t.Errorf("Expected user-june-4, got %s", result.Data[0].ID)
		}
	})

	// Test 2: Verify empty range returns 0 (not all users like the bug did)
	t.Run("EmptyRangeBugFix", func(t *testing.T) {
		filterRoot := filter.Root{
			Logic: filter.LogicAnd,
			FieldFilters: []filter.FieldFilter{
				{
					Field:    "created_at",
					Mode:     filter.ModeRange,
					DataType: filter.DataTypeDate,
					Value: filter.Range{
						From: "2025-08-01T00:00:00.000Z",
						To:   "2025-08-31T23:59:59.000Z",
					},
				},
			},
		}

		result, err := handler.DataQuery(users, filterRoot, 1, 100)
		if err != nil {
			t.Fatalf("Error filtering: %v", err)
		}

		// Should return 0 users (no users in August)
		if result.TotalSize != 0 {
			t.Errorf("Expected 0 users in empty range, got %d (this was the original bug)", result.TotalSize)
		}
	})
}
