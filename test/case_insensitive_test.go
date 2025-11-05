package test

import (
	"testing"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	ID          uint   `json:"id" gorm:"primarykey"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
}

func setupProductDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	products := []Product{
		{ID: 1, Name: "iPhone 15 Pro", Category: "Electronics", Description: "Latest Apple smartphone", Brand: "Apple"},
		{ID: 2, Name: "Samsung Galaxy S24", Category: "Electronics", Description: "Premium Android phone", Brand: "Samsung"},
		{ID: 3, Name: "MacBook Pro M3", Category: "Computers", Description: "Professional laptop", Brand: "Apple"},
		{ID: 4, Name: "Dell XPS 15", Category: "Computers", Description: "Windows laptop", Brand: "Dell"},
		{ID: 5, Name: "iPad Air", Category: "Tablets", Description: "Versatile tablet", Brand: "Apple"},
	}

	db.Create(&products)
	return db
}

// TestCaseInsensitiveGORM tests case-insensitive text filtering with GORM
func TestCaseInsensitiveGORM(t *testing.T) {
	db := setupProductDB(t)
	handler := filter.NewFilter[Product]()

	tests := []struct {
		name          string
		searchValue   string
		mode          filter.Mode
		field         string
		expectedCount int
		description   string
	}{
		{
			name:          "Equal - lowercase search",
			searchValue:   "iphone 15 pro",
			mode:          filter.ModeEqual,
			field:         "name",
			expectedCount: 1,
			description:   "Should find 'iPhone 15 Pro' with lowercase search",
		},
		{
			name:          "Equal - UPPERCASE search",
			searchValue:   "IPHONE 15 PRO",
			mode:          filter.ModeEqual,
			field:         "name",
			expectedCount: 1,
			description:   "Should find 'iPhone 15 Pro' with UPPERCASE search",
		},
		{
			name:          "Equal - MiXeD case",
			searchValue:   "IpHoNe 15 pRo",
			mode:          filter.ModeEqual,
			field:         "name",
			expectedCount: 1,
			description:   "Should find 'iPhone 15 Pro' with mixed case",
		},
		{
			name:          "Contains - lowercase",
			searchValue:   "apple",
			mode:          filter.ModeContains,
			field:         "brand",
			expectedCount: 3,
			description:   "Should find all Apple products with lowercase",
		},
		{
			name:          "Contains - UPPERCASE",
			searchValue:   "APPLE",
			mode:          filter.ModeContains,
			field:         "brand",
			expectedCount: 3,
			description:   "Should find all Apple products with UPPERCASE",
		},
		{
			name:          "StartsWith - lowercase",
			searchValue:   "mac",
			mode:          filter.ModeStartsWith,
			field:         "name",
			expectedCount: 1,
			description:   "Should find MacBook with lowercase start",
		},
		{
			name:          "StartsWith - UPPERCASE",
			searchValue:   "MAC",
			mode:          filter.ModeStartsWith,
			field:         "name",
			expectedCount: 1,
			description:   "Should find MacBook with UPPERCASE start",
		},
		{
			name:          "EndsWith - lowercase",
			searchValue:   "pro",
			mode:          filter.ModeEndsWith,
			field:         "name",
			expectedCount: 1,
			description:   "Should find MacBook Pro ending with 'Pro' using lowercase",
		},
		{
			name:          "EndsWith - UPPERCASE",
			searchValue:   "PRO",
			mode:          filter.ModeEndsWith,
			field:         "name",
			expectedCount: 1,
			description:   "Should find MacBook Pro ending with 'Pro' using UPPERCASE",
		},
		{
			name:          "NotEqual - lowercase",
			searchValue:   "apple",
			mode:          filter.ModeNotEqual,
			field:         "brand",
			expectedCount: 2,
			description:   "Should exclude Apple products with lowercase",
		},
		{
			name:          "NotContains - UPPERCASE",
			searchValue:   "APPLE",
			mode:          filter.ModeNotContains,
			field:         "brand",
			expectedCount: 2,
			description:   "Should exclude Apple products with UPPERCASE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    tt.field,
						Value:    tt.searchValue,
						Mode:     tt.mode,
						DataType: filter.DataTypeText,
					},
				},
			}

			result, err := handler.DataGorm(db, filterRoot, 1, 10)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}

			if result.TotalSize != tt.expectedCount {
				t.Errorf("%s: expected %d products, got %d", tt.description, tt.expectedCount, result.TotalSize)
			}
		})
	}
}

// TestCaseInsensitiveDataQuery tests case-insensitive text filtering with in-memory filtering
func TestCaseInsensitiveDataQuery(t *testing.T) {
	// Create test data in memory
	products := []*Product{
		{ID: 1, Name: "iPhone 15 Pro", Category: "Electronics", Description: "Latest Apple smartphone", Brand: "Apple"},
		{ID: 2, Name: "Samsung Galaxy S24", Category: "Electronics", Description: "Premium Android phone", Brand: "Samsung"},
		{ID: 3, Name: "MacBook Pro M3", Category: "Computers", Description: "Professional laptop", Brand: "Apple"},
		{ID: 4, Name: "Dell XPS 15", Category: "Computers", Description: "Windows laptop", Brand: "Dell"},
		{ID: 5, Name: "iPad Air", Category: "Tablets", Description: "Versatile tablet", Brand: "Apple"},
	}

	handler := filter.NewFilter[Product]()

	tests := []struct {
		name          string
		searchValue   string
		mode          filter.Mode
		field         string
		expectedCount int
		description   string
	}{
		{
			name:          "Equal - lowercase search",
			searchValue:   "iphone 15 pro",
			mode:          filter.ModeEqual,
			field:         "name",
			expectedCount: 1,
			description:   "DataQuery: Should find 'iPhone 15 Pro' with lowercase",
		},
		{
			name:          "Equal - UPPERCASE search",
			searchValue:   "IPHONE 15 PRO",
			mode:          filter.ModeEqual,
			field:         "name",
			expectedCount: 1,
			description:   "DataQuery: Should find 'iPhone 15 Pro' with UPPERCASE",
		},
		{
			name:          "Contains - lowercase",
			searchValue:   "apple",
			mode:          filter.ModeContains,
			field:         "brand",
			expectedCount: 3,
			description:   "DataQuery: Should find all Apple products with lowercase",
		},
		{
			name:          "Contains - UPPERCASE",
			searchValue:   "APPLE",
			mode:          filter.ModeContains,
			field:         "brand",
			expectedCount: 3,
			description:   "DataQuery: Should find all Apple products with UPPERCASE",
		},
		{
			name:          "StartsWith - lowercase",
			searchValue:   "mac",
			mode:          filter.ModeStartsWith,
			field:         "name",
			expectedCount: 1,
			description:   "DataQuery: Should find MacBook with lowercase",
		},
		{
			name:          "EndsWith - UPPERCASE",
			searchValue:   "PRO",
			mode:          filter.ModeEndsWith,
			field:         "name",
			expectedCount: 1,
			description:   "DataQuery: Should find MacBook Pro ending with 'Pro' using UPPERCASE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    tt.field,
						Value:    tt.searchValue,
						Mode:     tt.mode,
						DataType: filter.DataTypeText,
					},
				},
			}

			result, err := handler.DataQuery(products, filterRoot, 1, 10)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}

			if result.TotalSize != tt.expectedCount {
				t.Errorf("%s: expected %d products, got %d", tt.description, tt.expectedCount, result.TotalSize)
			}
		})
	}
}

// TestCaseInsensitiveConsistency ensures GORM and DataQuery return the same results
func TestCaseInsensitiveConsistency(t *testing.T) {
	db := setupProductDB(t)
	handler := filter.NewFilter[Product]()

	// Fetch all data for in-memory comparison
	var allProducts []*Product
	db.Find(&allProducts)

	testCases := []struct {
		searchValue string
		mode        filter.Mode
		field       string
	}{
		{"APPLE", filter.ModeEqual, "brand"},
		{"iphone", filter.ModeContains, "name"},
		{"PRO", filter.ModeEndsWith, "name"},
		{"mac", filter.ModeStartsWith, "name"},
		{"samsung", filter.ModeNotEqual, "brand"},
	}

	for _, tc := range testCases {
		t.Run(tc.searchValue+"_"+string(tc.mode), func(t *testing.T) {
			filterRoot := filter.Root{
				Logic: filter.LogicAnd,
				FieldFilters: []filter.FieldFilter{
					{
						Field:    tc.field,
						Value:    tc.searchValue,
						Mode:     tc.mode,
						DataType: filter.DataTypeText,
					},
				},
			}

			// Test with GORM
			gormResult, err := handler.DataGorm(db, filterRoot, 1, 100)
			if err != nil {
				t.Fatalf("GORM error: %v", err)
			}

			// Test with DataQuery
			queryResult, err := handler.DataQuery(allProducts, filterRoot, 1, 100)
			if err != nil {
				t.Fatalf("DataQuery error: %v", err)
			}

			// Compare results
			if gormResult.TotalSize != queryResult.TotalSize {
				t.Errorf("Inconsistent results: GORM returned %d, DataQuery returned %d",
					gormResult.TotalSize, queryResult.TotalSize)
			}
		})
	}
}
