// Package main demonstrates Hybrid functionality with SQLite
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        uint `gorm:"primarykey"`
	Name      string
	Age       int
	Email     string
	IsActive  bool
	CreatedAt time.Time
}

func main() {
	fmt.Println("=== Hybrid SQLite Test ===")

	// Initialize database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Auto migrate
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal(err)
	}

	// Seed test data
	seedData(db)

	// Run ANALYZE to populate sqlite_stat1 for better estimation
	db.Exec("ANALYZE")
	fmt.Println("Ran ANALYZE to populate SQLite statistics")

	// Get actual row count for comparison
	var actualCount int64
	db.Model(&User{}).Count(&actualCount)
	fmt.Printf("Actual users in database: %d\n\n", actualCount)

	// Create filter handler
	filterHandler := filter.NewFilter[User]()

	// Define filters
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "name",
				Value:    "John",
				Mode:     filter.ModeContains,
				DataType: filter.DataTypeText,
			},
			{
				Field:    "age",
				Value:    18,
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.SortOrderDesc},
			{Field: "name", Order: filter.SortOrderAsc},
		},
	}

	fmt.Println("--- Test 1: Low Threshold (50) - Should use DATABASE filtering ---")
	testHybrid(filterHandler, db, filterRoot, 50, "Low threshold forces DB filtering")

	fmt.Println("\n--- Test 2: Medium Threshold (100) - Should use IN-MEMORY filtering ---")
	testHybrid(filterHandler, db, filterRoot, 100, "Medium threshold uses in-memory")

	fmt.Println("\n--- Test 3: High Threshold (1000) - Should use IN-MEMORY filtering ---")
	testHybrid(filterHandler, db, filterRoot, 1000, "High threshold uses in-memory")

	fmt.Println("\n--- Test 4: Compare all three methods ---")
	compareingMethods(filterHandler, db, filterRoot)

	fmt.Println("\n--- Test 5: Active users only (different filter) ---")
	testActiveUsers(filterHandler, db)
}

func testHybrid(filterHandler *filter.Handler[User], db *gorm.DB, filterRoot filter.Root, threshold int64, description string) {
	fmt.Printf("Description: %s\n", description)
	fmt.Printf("Threshold: %d rows\n", threshold)

	// Get estimated rows to see what strategy will be used
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(User)); err != nil {
		log.Printf("Warning: failed to parse model: %v\n", err)
		return
	}

	// Try to get estimation (simplified version without exposing internal method)
	var est struct{ Rows int64 }
	db.Raw("SELECT COUNT(*) AS rows FROM " + stmt.Table).Scan(&est)
	estimatedRows := est.Rows

	expectedStrategy := "DATABASE"
	if estimatedRows <= threshold {
		expectedStrategy = "IN-MEMORY"
	}
	fmt.Printf("Estimated rows: %d, Expected strategy: %s\n", estimatedRows, expectedStrategy)

	start := time.Now()
	result, err := filterHandler.Hybrid(db, threshold, filterRoot, 1, 10)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Execution time: %v\n", elapsed)
	fmt.Printf("Total: %d, Pages: %d, Current Page: %d/%d\n",
		result.TotalSize, result.TotalPage, result.PageIndex, result.TotalPage)
	fmt.Printf("Found %d users on this page\n", len(result.Data))

	for i, user := range result.Data {
		fmt.Printf("  %d. Name: %-20s Age: %d, Active: %v\n",
			i+1, user.Name, user.Age, user.IsActive)
	}
}

func compareingMethods(filterHandler *filter.Handler[User], db *gorm.DB, filterRoot filter.Root) {
	fmt.Println("Comparing DataQuery vs DataGorm vs Hybrid:")
	fmt.Println()

	// Test 1: In-Memory (DataQuery)
	fmt.Println("1. DataQuery (In-Memory):")
	var allUsers []*User
	db.Find(&allUsers)
	start := time.Now()
	result1, err := filterHandler.DataQuery(allUsers, filterRoot, 1, 10)
	elapsed1 := time.Since(start)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   Time: %v, Results: %d\n", elapsed1, len(result1.Data))
	}

	// Test 2: Database (DataGorm)
	fmt.Println("2. DataGorm (Database):")
	start = time.Now()
	result2, err := filterHandler.DataGorm(db, filterRoot, 1, 10)
	elapsed2 := time.Since(start)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   Time: %v, Results: %d\n", elapsed2, len(result2.Data))
	}

	// Test 3: Hybrid with low threshold (forces DB)
	fmt.Println("3. Hybrid (threshold=50, expects DB):")
	start = time.Now()
	result3, err := filterHandler.Hybrid(db, 50, filterRoot, 1, 10)
	elapsed3 := time.Since(start)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   Time: %v, Results: %d\n", elapsed3, len(result3.Data))
	}

	// Test 4: Hybrid with high threshold (forces in-memory)
	fmt.Println("4. Hybrid (threshold=200, expects In-Memory):")
	start = time.Now()
	result4, err := filterHandler.Hybrid(db, 200, filterRoot, 1, 10)
	elapsed4 := time.Since(start)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   Time: %v, Results: %d\n", elapsed4, len(result4.Data))
	}

	// Verify all methods return same results
	fmt.Println()
	if result1 != nil && result2 != nil && result3 != nil && result4 != nil {
		if result1.TotalSize == result2.TotalSize && result2.TotalSize == result3.TotalSize && result3.TotalSize == result4.TotalSize {
			fmt.Printf("✓ All methods returned consistent results (%d total records)\n", result1.TotalSize)
		} else {
			fmt.Printf("✗ Warning: Methods returned different totals: Q=%d, G=%d, H1=%d, H2=%d\n",
				result1.TotalSize, result2.TotalSize, result3.TotalSize, result4.TotalSize)
		}
	}
}

func testActiveUsers(filterHandler *filter.Handler[User], db *gorm.DB) {
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
				Value:    25,
				Mode:     filter.ModeGTE,
				DataType: filter.DataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.SortOrderAsc},
		},
	}

	result, err := filterHandler.Hybrid(db, 10, filterRoot, 1, 10)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Active users aged 25+:\n")
	fmt.Printf("Total: %d users\n", result.TotalSize)
	for i, user := range result.Data {
		fmt.Printf("  %d. Name: %-20s Age: %d, Email: %s\n",
			i+1, user.Name, user.Age, user.Email)
	}
}

func seedData(db *gorm.DB) {
	// Check if data already exists
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		fmt.Printf("Database already has %d users, skipping seed\n", count)
		return
	}

	users := []User{
		{Name: "John Doe", Age: 25, Email: "john@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "Jane Smith", Age: 30, Email: "jane@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "John Smith", Age: 22, Email: "johnsmith@example.com", IsActive: false, CreatedAt: time.Now()},
		{Name: "Bob Johnson", Age: 35, Email: "bob@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "Alice Wonder", Age: 28, Email: "alice@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "Johnny Walker", Age: 40, Email: "johnny@example.com", IsActive: true, CreatedAt: time.Now()},
		{Name: "Sarah Connor", Age: 32, Email: "sarah@example.com", IsActive: false, CreatedAt: time.Now()},
		{Name: "John Connor", Age: 18, Email: "johnconnor@example.com", IsActive: true, CreatedAt: time.Now()},
	}

	// Add many more test users for better threshold demonstration
	baseTime := time.Now().AddDate(0, -12, 0)
	for i := range 100 {
		users = append(users, User{
			Name:      fmt.Sprintf("User %d", i),
			Age:       20 + (i % 50),
			Email:     fmt.Sprintf("user%d@example.com", i),
			IsActive:  i%3 != 0, // ~66% active
			CreatedAt: baseTime.AddDate(0, 0, i),
		})
	}

	db.Create(&users)
	fmt.Printf("Seeded %d test users\n", len(users))
}
