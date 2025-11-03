package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"github.com/Lands-Horizon-Corp/golang-filtering/seeder"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Command line flags
	pageSize := flag.Int("page-size", 50, "Number of records per page")
	pageIndex := flag.Int("page", 1, "Page number to fetch")

	flag.Parse()

	// Initialize database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create filter handler
	filterHandler := filter.NewFilter[seeder.User]()

	fmt.Println("=== Filter Performance Test ===")

	// Test 1: Simple text filter
	fmt.Println("Test 1: Finding users with 'John' in name")
	test1 := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "name",
				Value:          "John",
				Mode:           filter.FilterModeContains,
				FilterDataType: filter.FilterDataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.FilterSortOrderAsc},
		},
	}
	runTest(db, filterHandler, test1, *pageIndex, *pageSize)

	// Test 2: Multiple filters with AND logic
	fmt.Println("\nTest 2: Users named John, age >= 25, and active")
	test2 := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "name",
				Value:          "John",
				Mode:           filter.FilterModeContains,
				FilterDataType: filter.FilterDataTypeText,
			},
			{
				Field:          "age",
				Value:          25,
				Mode:           filter.FilterModeGTE,
				FilterDataType: filter.FilterDataTypeNumber,
			},
			{
				Field:          "is_active",
				Value:          true,
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeBool,
			},
		},
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.FilterSortOrderDesc},
			{Field: "name", Order: filter.FilterSortOrderAsc},
		},
	}
	runTest(db, filterHandler, test2, *pageIndex, *pageSize)

	// Test 3: Age range filter
	fmt.Println("\nTest 3: Users aged between 30 and 40")
	test3 := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field: "age",
				Value: filter.FilterRange{
					From: 30,
					To:   40,
				},
				Mode:           filter.FilterModeRange,
				FilterDataType: filter.FilterDataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.FilterSortOrderAsc},
		},
	}
	runTest(db, filterHandler, test3, *pageIndex, *pageSize)

	// Test 4: Date filter (created in last 30 days)
	fmt.Println("\nTest 4: Users created in the last 30 days")
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	test4 := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "created_at",
				Value:          thirtyDaysAgo,
				Mode:           filter.FilterModeAfter,
				FilterDataType: filter.FilterDataTypeDate,
			},
		},
		SortFields: []filter.SortField{
			{Field: "created_at", Order: filter.FilterSortOrderDesc},
		},
	}
	runTest(db, filterHandler, test4, *pageIndex, *pageSize)

	// Test 5: OR logic
	fmt.Println("\nTest 5: Users named John OR Smith (OR logic)")
	test5 := filter.FilterRoot{
		Logic: filter.FilterLogicOr,
		Filters: []filter.Filter{
			{
				Field:          "name",
				Value:          "John",
				Mode:           filter.FilterModeContains,
				FilterDataType: filter.FilterDataTypeText,
			},
			{
				Field:          "name",
				Value:          "Smith",
				Mode:           filter.FilterModeContains,
				FilterDataType: filter.FilterDataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "name", Order: filter.FilterSortOrderAsc},
		},
	}
	runTest(db, filterHandler, test5, *pageIndex, *pageSize)

	// Test 6: Complex filter
	fmt.Println("\nTest 6: Complex filter (age 20-50, active, email contains gmail)")
	test6 := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field: "age",
				Value: filter.FilterRange{
					From: 20,
					To:   50,
				},
				Mode:           filter.FilterModeRange,
				FilterDataType: filter.FilterDataTypeNumber,
			},
			{
				Field:          "is_active",
				Value:          true,
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeBool,
			},
			{
				Field:          "email",
				Value:          "gmail",
				Mode:           filter.FilterModeContains,
				FilterDataType: filter.FilterDataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.FilterSortOrderAsc},
		},
	}
	runTest(db, filterHandler, test6, *pageIndex, *pageSize)
}

func runTest(db *gorm.DB, handler *filter.FilterHandler[seeder.User], filterRoot filter.FilterRoot, pageIndex, pageSize int) {
	startTime := time.Now()

	result, err := handler.FilterDataGorm(db, filterRoot, pageIndex, pageSize)
	if err != nil {
		log.Printf("  ✗ Error: %v\n", err)
		return
	}

	elapsed := time.Since(startTime)

	fmt.Printf("  ✓ Query time: %s\n", elapsed.Round(time.Microsecond))
	fmt.Printf("  • Total matches: %d\n", result.TotalSize)
	fmt.Printf("  • Total pages: %d\n", result.TotalPage)
	fmt.Printf("  • Current page: %d/%d\n", result.PageIndex, result.TotalPage)
	fmt.Printf("  • Records on this page: %d\n", len(result.Data))

	if len(result.Data) > 0 {
		fmt.Printf("  • First record: %s (Age: %d, Active: %v)\n",
			result.Data[0].Name, result.Data[0].Age, result.Data[0].IsActive)
	}
}
