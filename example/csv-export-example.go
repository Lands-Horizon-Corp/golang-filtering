package example

import (
	"fmt"
	"log"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// User represents a sample user model for CSV export
type User struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"`
}

// CSVExportExample demonstrates how to use DataQueryNoPageCSV for exporting filtered data
// The function automatically uses all available fields from the struct as CSV columns
func CSVExportExample() {
	// Sample data
	users := []*User{
		{ID: 1, Name: "Alice Johnson", Email: "alice@example.com", Age: 30, IsActive: true, Role: "admin"},
		{ID: 2, Name: "Bob Smith", Email: "bob@example.com", Age: 25, IsActive: false, Role: "user"},
		{ID: 3, Name: "Charlie Brown", Email: "charlie@example.com", Age: 35, IsActive: true, Role: "moderator"},
		{ID: 4, Name: "Diana Prince", Email: "diana@example.com", Age: 28, IsActive: true, Role: "user"},
		{ID: 5, Name: "Eve Adams", Email: "eve@example.com", Age: 32, IsActive: true, Role: "admin"},
	}

	// Create filter handler
	handler := filter.NewFilter[User](filter.GolangFilteringConfig{})

	// Example 1: Export all users
	fmt.Println("=== Example 1: Export All Users ===")
	filterRoot := filter.Root{
		Logic: filter.LogicAnd,
	}

	csvData, err := handler.DataQueryNoPageCSV(users, filterRoot)
	if err != nil {
		log.Fatal("Error generating CSV:", err)
	}

	fmt.Println(string(csvData))

	// Example 2: Export only active users, sorted by age
	fmt.Println("\n=== Example 2: Active Users Sorted by Age ===")
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "is_active",
				DataType: filter.DataTypeBool,
				Mode:     filter.ModeEqual,
				Value:    true,
			},
		},
		SortFields: []filter.SortField{
			{
				Field: "age",
				Order: filter.SortOrderAsc,
			},
		},
	}

	csvData, err = handler.DataQueryNoPageCSV(users, filterRoot)
	if err != nil {
		log.Fatal("Error generating filtered CSV:", err)
	}

	fmt.Println(string(csvData))

	// Example 3: Export admins only
	fmt.Println("\n=== Example 3: Admin Users Only ===")
	filterRoot = filter.Root{
		Logic: filter.LogicAnd,
		FieldFilters: []filter.FieldFilter{
			{
				Field:    "role",
				DataType: filter.DataTypeText,
				Mode:     filter.ModeEqual,
				Value:    "admin",
			},
		},
	}

	csvData, err = handler.DataQueryNoPageCSV(users, filterRoot)
	if err != nil {
		log.Fatal("Error generating admin CSV:", err)
	}

	fmt.Println(string(csvData))

	// Example 4: Hybrid CSV export (intelligent strategy selection)
	// Note: This would typically use a real database connection
	fmt.Println("\n=== Example 4: Hybrid CSV Export (simulated) ===")
	fmt.Println("// For hybrid CSV export, you would use:")
	fmt.Println("// csvData, err := handler.HybridCSV(db, 10000, filterRoot)")
	fmt.Println("// This automatically chooses between in-memory and database filtering")
	fmt.Println("// based on estimated table size vs threshold")
}
