package example

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/gorm"
)

// Order represents a sample order model for hybrid filtering examples
type Order struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	OrderNumber string    `json:"order_number"`
	CustomerID  uint      `json:"customer_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	IsPaid      bool      `json:"is_paid"`
	OrderDate   time.Time `json:"order_date"`
	ShipDate    time.Time `json:"ship_date"`
}

// HybridFilterSample demonstrates automatic strategy selection using FilterHybrid
func HybridFilterSample(db *gorm.DB) {
	fmt.Println("=== Hybrid Filter Example ===")
	fmt.Println("Automatically switches between in-memory and database filtering based on data size")
	fmt.Println()

	// Create filter handler
	filterHandler := filter.NewFilter[Order]()

	// Example 1: Small dataset (threshold = 10000)
	// If table has <= 10k rows, uses in-memory filtering for speed
	// If table has > 10k rows, uses database filtering to save memory
	fmt.Println("Example 1: Find paid orders with amount >= $100 (Auto-switch at 10k threshold)")
	hybridExample1(filterHandler, db)

	// Example 2: Very small threshold (threshold = 100)
	// Forces database filtering for tables with > 100 rows
	fmt.Println("\nExample 2: Find pending orders (Force DB filtering at 100 threshold)")
	hybridExample2(filterHandler, db)

	// Example 3: Large threshold (threshold = 1000000)
	// Uses in-memory filtering for most datasets
	fmt.Println("\nExample 3: Orders from specific customer (Prefer in-memory at 1M threshold)")
	hybridExample3(filterHandler, db)

	// Example 4: Complex filter with date range
	fmt.Println("\nExample 4: Recent orders with complex filters (Smart auto-detection)")
	hybridExample4(filterHandler, db)
}

func hybridExample1(filterHandler *filter.FilterHandler[Order], db *gorm.DB) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "is_paid",
				Value:          true,
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeBool,
			},
			{
				Field:          "total_amount",
				Value:          100.0,
				Mode:           filter.FilterModeGTE,
				FilterDataType: filter.FilterDataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "total_amount", Order: filter.FilterSortOrderDesc},
		},
	}

	// Threshold: 10,000 rows
	// If orders table has <= 10k rows: uses FilterDataQuery (in-memory, parallel processing)
	// If orders table has > 10k rows: uses FilterDataGorm (database query)
	result, err := filterHandler.FilterHybrid(db, 10000, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printHybridResults(result, "Paid orders >= $100")
}

func hybridExample2(filterHandler *filter.FilterHandler[Order], db *gorm.DB) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "status",
				Value:          "pending",
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "order_date", Order: filter.FilterSortOrderDesc},
		},
	}

	// Low threshold: 100 rows
	// Forces database filtering for most real-world scenarios
	result, err := filterHandler.FilterHybrid(db, 100, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printHybridResults(result, "Pending orders")
}

func hybridExample3(filterHandler *filter.FilterHandler[Order], db *gorm.DB) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "customer_id",
				Value:          12345,
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "order_date", Order: filter.FilterSortOrderDesc},
		},
	}

	// High threshold: 1,000,000 rows
	// Prefers in-memory filtering unless dataset is massive
	result, err := filterHandler.FilterHybrid(db, 1000000, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printHybridResults(result, "Customer #12345 orders")
}

func hybridExample4(filterHandler *filter.FilterHandler[Order], db *gorm.DB) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "order_date",
				Value:          thirtyDaysAgo,
				Mode:           filter.FilterModeAfter,
				FilterDataType: filter.FilterDataTypeDate,
			},
			{
				Field:          "is_paid",
				Value:          true,
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeBool,
			},
			{
				Field:          "total_amount",
				Value:          filter.Range{From: 50, To: 500},
				Mode:           filter.FilterModeRange,
				FilterDataType: filter.FilterDataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "order_date", Order: filter.FilterSortOrderDesc},
			{Field: "total_amount", Order: filter.FilterSortOrderDesc},
		},
	}

	// Balanced threshold: 50,000 rows
	// Good default for most applications
	result, err := filterHandler.FilterHybrid(db, 50000, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printHybridResults(result, "Recent orders ($50-$500)")
}

func printHybridResults(result *filter.PaginationResult[Order], description string) {
	fmt.Printf("Results for: %s\n", description)
	fmt.Printf("Total: %d records, Page: %d/%d, Page Size: %d\n",
		result.TotalSize, result.PageIndex, result.TotalPage, result.PageSize)
	fmt.Println("Orders:")

	if len(result.Data) == 0 {
		fmt.Println("  (no results)")
		return
	}

	for i, order := range result.Data {
		fmt.Printf("  %d. Order #%-15s Customer: %5d, Amount: $%8.2f, Status: %-10s Paid: %-5v Date: %s\n",
			i+1, order.OrderNumber, order.CustomerID, order.TotalAmount, order.Status, order.IsPaid,
			order.OrderDate.Format("2006-01-02"))
	}
	fmt.Println()
}
