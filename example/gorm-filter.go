// Package example provides sample implementations for the filtering package
package example

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/gorm"
)

// Product represents a sample product model for GORM filtering examples
type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	InStock     bool      `json:"in_stock"`
	Rating      float64   `json:"rating"`
	LaunchedAt  time.Time `json:"launched_at"`
	Description string    `json:"description"`
}

// GormFilterSample demonstrates database filtering using FilterDataGorm
func GormFilterSample(db *gorm.DB) {
	fmt.Println("=== GORM Database Filter Example ===")

	// Create filter handler
	filterHandler := filter.NewFilter[Product]()

	// Example 1: Find products with "Pro" in name
	fmt.Println("Example 1: Find products with 'Pro' in their name")
	gormExample1(filterHandler, db)

	// Example 2: Find products in price range $500-$1500
	fmt.Println("\nExample 2: Products priced between $500 and $1500")
	gormExample2(filterHandler, db)

	// Example 3: In-stock products with rating >= 4.0
	fmt.Println("\nExample 3: In-stock products with rating >= 4.0")
	gormExample3(filterHandler, db)

	// Example 4: Electronics OR Computers category
	fmt.Println("\nExample 4: Products in Electronics OR Computers category")
	gormExample4(filterHandler, db)

	// Example 5: Products launched in last 6 months
	fmt.Println("\nExample 5: Products launched in last 6 months")
	gormExample5(filterHandler, db)
}

func gormExample1(filterHandler *filter.FilterHandler[Product], db *gorm.DB) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "name",
				Value:          "Pro",
				Mode:           filter.FilterModeContains,
				FilterDataType: filter.FilterDataTypeText,
			},
		},
	}

	result, err := filterHandler.FilterDataGorm(db, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printGormResults(result)
}

func gormExample2(filterHandler *filter.FilterHandler[Product], db *gorm.DB) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "price",
				Value:          filter.FilterRange{From: 500, To: 1500},
				Mode:           filter.FilterModeRange,
				FilterDataType: filter.FilterDataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "price", Order: filter.FilterSortOrderAsc},
		},
	}

	result, err := filterHandler.FilterDataGorm(db, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printGormResults(result)
}

func gormExample3(filterHandler *filter.FilterHandler[Product], db *gorm.DB) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "in_stock",
				Value:          true,
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeBool,
			},
			{
				Field:          "rating",
				Value:          4.0,
				Mode:           filter.FilterModeGTE,
				FilterDataType: filter.FilterDataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "rating", Order: filter.FilterSortOrderDesc},
		},
	}

	result, err := filterHandler.FilterDataGorm(db, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printGormResults(result)
}

func gormExample4(filterHandler *filter.FilterHandler[Product], db *gorm.DB) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicOr,
		Filters: []filter.Filter{
			{
				Field:          "category",
				Value:          "Electronics",
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeText,
			},
			{
				Field:          "category",
				Value:          "Computers",
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "category", Order: filter.FilterSortOrderAsc},
			{Field: "name", Order: filter.FilterSortOrderAsc},
		},
	}

	result, err := filterHandler.FilterDataGorm(db, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printGormResults(result)
}

func gormExample5(filterHandler *filter.FilterHandler[Product], db *gorm.DB) {
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)

	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "launched_at",
				Value:          sixMonthsAgo,
				Mode:           filter.FilterModeAfter,
				FilterDataType: filter.FilterDataTypeDate,
			},
		},
		SortFields: []filter.SortField{
			{Field: "launched_at", Order: filter.FilterSortOrderDesc},
		},
	}

	result, err := filterHandler.FilterDataGorm(db, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printGormResults(result)
}

func printGormResults(result *filter.PaginationResult[Product]) {
	fmt.Printf("Total: %d records, Page: %d/%d, Page Size: %d\n",
		result.TotalSize, result.PageIndex, result.TotalPage, result.PageSize)
	fmt.Println("Results:")

	if len(result.Data) == 0 {
		fmt.Println("  (no results)")
		return
	}

	for i, product := range result.Data {
		fmt.Printf("  %d. ID: %d, Name: %-30s Price: $%.2f, Category: %-15s InStock: %-5v Rating: %.1f\n",
			i+1, product.ID, product.Name, product.Price, product.Category, product.InStock, product.Rating)
	}
}
