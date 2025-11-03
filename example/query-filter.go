package example

import (
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

// User1 represents a sample user model for in-memory filtering examples
type User1 struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	Role      string    `json:"role"`
}

// QueryFilterSample demonstrates in-memory filtering using FilterDataQuery
func QueryFilterSample() {
	fmt.Println("=== In-Memory Query Filter Example ===")

	// Sample data
	users := []*User1{
		{ID: 1, Name: "John Doe", Age: 25, Email: "john@example.com", IsActive: true, CreatedAt: time.Now().AddDate(0, -6, 0), Role: "admin"},
		{ID: 2, Name: "Jane Smith", Age: 30, Email: "jane@example.com", IsActive: true, CreatedAt: time.Now().AddDate(0, -3, 0), Role: "user"},
		{ID: 3, Name: "John Smith", Age: 22, Email: "johnsmith@example.com", IsActive: false, CreatedAt: time.Now().AddDate(0, -1, 0), Role: "user"},
		{ID: 4, Name: "Bob Johnson", Age: 35, Email: "bob@example.com", IsActive: true, CreatedAt: time.Now().AddDate(0, -12, 0), Role: "admin"},
		{ID: 5, Name: "Alice Wonder", Age: 28, Email: "alice@example.com", IsActive: true, CreatedAt: time.Now().AddDate(0, -2, 0), Role: "moderator"},
		{ID: 6, Name: "Johnny Walker", Age: 40, Email: "johnny@example.com", IsActive: true, CreatedAt: time.Now().AddDate(0, -8, 0), Role: "user"},
		{ID: 7, Name: "Sarah Connor", Age: 32, Email: "sarah@example.com", IsActive: false, CreatedAt: time.Now().AddDate(0, -5, 0), Role: "user"},
		{ID: 8, Name: "John Connor", Age: 18, Email: "johnconnor@example.com", IsActive: true, CreatedAt: time.Now().AddDate(0, -1, 0), Role: "user"},
		{ID: 9, Name: "Mike Wilson", Age: 45, Email: "mike@example.com", IsActive: true, CreatedAt: time.Now().AddDate(-1, 0, 0), Role: "admin"},
		{ID: 10, Name: "Emma Brown", Age: 27, Email: "emma@example.com", IsActive: true, CreatedAt: time.Now().AddDate(0, -4, 0), Role: "moderator"},
	}

	// Create filter handler
	filterHandler := filter.NewFilter[User1]()

	// Example 1: Simple text contains filter
	fmt.Println("Example 1: Find all users with 'John' in their name")
	queryExample1(filterHandler, users)

	// Example 2: Number range filter with sorting
	fmt.Println("\nExample 2: Find users aged between 25 and 35, sorted by age")
	queryExample2(filterHandler, users)

	// Example 3: Multiple filters with AND logic
	fmt.Println("\nExample 3: Active users named 'John' who are 18 or older")
	queryExample3(filterHandler, users)

	// Example 4: OR logic filter
	fmt.Println("\nExample 4: Users who are either admins OR moderators")
	queryExample4(filterHandler, users)

	// Example 5: Date filter
	fmt.Println("\nExample 5: Users created in the last 3 months")
	queryExample5(filterHandler, users)

	// Example 6: Complex filter with pagination
	fmt.Println("\nExample 6: Active users aged 25+, sorted by age (descending), page 1 of 2")
	queryExample6(filterHandler, users)

	// Example 7: Empty/Not Empty filters
	fmt.Println("\nExample 7: Users with non-empty roles")
	queryExample7(filterHandler, users)
}

func queryExample1(filterHandler *filter.FilterHandler[User1], users []*User1) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "name",
				Value:          "John",
				Mode:           filter.FilterModeContains,
				FilterDataType: filter.FilterDataTypeText,
			},
		},
	}

	result, err := filterHandler.FilterDataQuery(users, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printQueryResults(result)
}

func queryExample2(filterHandler *filter.FilterHandler[User1], users []*User1) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "age",
				Value:          filter.FilterRange{From: 25, To: 35},
				Mode:           filter.FilterModeRange,
				FilterDataType: filter.FilterDataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.FilterSortOrderAsc},
		},
	}

	result, err := filterHandler.FilterDataQuery(users, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printQueryResults(result)
}

func queryExample3(filterHandler *filter.FilterHandler[User1], users []*User1) {
	filterRoot := filter.FilterRoot{
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
				Value:          18,
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
		},
	}

	result, err := filterHandler.FilterDataQuery(users, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printQueryResults(result)
}

func queryExample4(filterHandler *filter.FilterHandler[User1], users []*User1) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicOr, // OR logic
		Filters: []filter.Filter{
			{
				Field:          "role",
				Value:          "admin",
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeText,
			},
			{
				Field:          "role",
				Value:          "moderator",
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "role", Order: filter.FilterSortOrderAsc},
			{Field: "name", Order: filter.FilterSortOrderAsc},
		},
	}

	result, err := filterHandler.FilterDataQuery(users, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printQueryResults(result)
}

func queryExample5(filterHandler *filter.FilterHandler[User1], users []*User1) {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)

	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "created_at",
				Value:          threeMonthsAgo,
				Mode:           filter.FilterModeAfter,
				FilterDataType: filter.FilterDataTypeDate,
			},
		},
		SortFields: []filter.SortField{
			{Field: "created_at", Order: filter.FilterSortOrderDesc},
		},
	}

	result, err := filterHandler.FilterDataQuery(users, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printQueryResults(result)
}

func queryExample6(filterHandler *filter.FilterHandler[User1], users []*User1) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "is_active",
				Value:          true,
				Mode:           filter.FilterModeEqual,
				FilterDataType: filter.FilterDataTypeBool,
			},
			{
				Field:          "age",
				Value:          25,
				Mode:           filter.FilterModeGTE,
				FilterDataType: filter.FilterDataTypeNumber,
			},
		},
		SortFields: []filter.SortField{
			{Field: "age", Order: filter.FilterSortOrderDesc},
			{Field: "name", Order: filter.FilterSortOrderAsc},
		},
	}

	// Request page 1 with page size of 3
	result, err := filterHandler.FilterDataQuery(users, filterRoot, 1, 3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printQueryResults(result)
}

func queryExample7(filterHandler *filter.FilterHandler[User1], users []*User1) {
	filterRoot := filter.FilterRoot{
		Logic: filter.FilterLogicAnd,
		Filters: []filter.Filter{
			{
				Field:          "role",
				Value:          "",
				Mode:           filter.FilterModeIsNotEmpty,
				FilterDataType: filter.FilterDataTypeText,
			},
		},
		SortFields: []filter.SortField{
			{Field: "role", Order: filter.FilterSortOrderAsc},
			{Field: "name", Order: filter.FilterSortOrderAsc},
		},
	}

	result, err := filterHandler.FilterDataQuery(users, filterRoot, 1, 10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	printQueryResults(result)
}

func printQueryResults(result *filter.PaginationResult[User1]) {
	fmt.Printf("Total: %d records, Page: %d/%d, Page Size: %d\n",
		result.TotalSize, result.PageIndex, result.TotalPage, result.PageSize)
	fmt.Println("Results:")

	if len(result.Data) == 0 {
		fmt.Println("  (no results)")
		return
	}

	for i, user := range result.Data {
		fmt.Printf("  %d. ID: %d, Name: %-20s Age: %d, Active: %-5v Role: %s, Created: %s\n",
			i+1, user.ID, user.Name, user.Age, user.IsActive, user.Role,
			user.CreatedAt.Format("2006-01-02"))
	}
}
