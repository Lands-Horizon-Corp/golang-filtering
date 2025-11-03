# Golang Filtering Package

A powerful and flexible filtering package for Go with GORM support, providing dynamic filtering, sorting, and pagination capabilities.

## Features

- ✅ Dynamic filtering with multiple data types (number, text, bool, date, time)
- ✅ Multiple filter modes (equal, contains, greater than, range, etc.)
- ✅ AND/OR logic support
- ✅ Sorting with multiple fields
- ✅ Pagination support
- ✅ GORM integration for efficient database queries
- ✅ Type-safe with Go generics

## Installation

```bash
go get github.com/Lands-Horizon-Corp/golang-filtering
```

## Usage

### Basic Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/Lands-Horizon-Corp/golang-filtering/filter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primarykey"`
	Name      string
	Age       int
	Email     string
	IsActive  bool
	CreatedAt time.Time
}

func main() {
	// Initialize database
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Auto migrate
	db.AutoMigrate(&User{})

	// Create filter handler
	filterHandler := filter.NewFilter[User]()

	// Define filters
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
			{Field: "name", Order: filter.FilterSortOrderAsc},
		},
	}

	// Execute filtering with pagination
	result, err := filterHandler.FilterDataGorm(db, filterRoot, 1, 20)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total: %d, Pages: %d, Current Page: %d\n",
		result.TotalSize, result.TotalPage, result.PageIndex)
	fmt.Printf("Found %d users\n", len(result.Data))

	for _, user := range result.Data {
		fmt.Printf("User: %s, Age: %d\n", user.Name, user.Age)
	}
}
```

## Filter Modes

### Text Filters

- `equal` - Exact match (case-insensitive)
- `notEqual` - Not equal (case-insensitive)
- `contains` - Contains substring
- `notContains` - Does not contain substring
- `startsWith` - Starts with
- `endsWith` - Ends with
- `isEmpty` - Is empty or null
- `isNotEmpty` - Is not empty

### Number Filters

- `equal` - Equal to
- `notEqual` - Not equal to
- `gt` - Greater than
- `gte` - Greater than or equal
- `lt` - Less than
- `lte` - Less than or equal
- `range` - Between two values

### Boolean Filters

- `equal` - Equal to
- `notEqual` - Not equal to

### Date/Time Filters

- `equal` - Equal to date
- `notEqual` - Not equal to date
- `before` - Before date
- `after` - After date
- `gte` - Greater than or equal
- `lte` - Less than or equal
- `range` - Between two dates

## Filter Logic

```go
// AND logic - all filters must match
filterRoot := filter.FilterRoot{
	Logic: filter.FilterLogicAnd,
	Filters: []filter.Filter{...},
}

// OR logic - any filter can match
filterRoot := filter.FilterRoot{
	Logic: filter.FilterLogicOr,
	Filters: []filter.Filter{...},
}
```

## Sorting

```go
filterRoot.SortFields = []filter.SortField{
	{Field: "age", Order: filter.FilterSortOrderDesc},
	{Field: "name", Order: filter.FilterSortOrderAsc},
}
```

## Pagination

```go
// Get page 2 with 50 items per page
result, err := filterHandler.FilterDataGorm(db, filterRoot, 2, 50)

// Result contains:
// - result.Data: The actual data
// - result.TotalSize: Total matching records
// - result.TotalPage: Total number of pages
// - result.PageIndex: Current page
// - result.PageSize: Items per page
```

## Range Filters

```go
// Age between 18 and 65
filter.Filter{
	Field: "age",
	Value: filter.FilterRange{
		From: 18,
		To:   65,
	},
	Mode:           filter.FilterModeRange,
	FilterDataType: filter.FilterDataTypeNumber,
}

// Created between two dates
filter.Filter{
	Field: "created_at",
	Value: filter.FilterRange{
		From: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
	},
	Mode:           filter.FilterModeRange,
	FilterDataType: filter.FilterDataTypeDate,
}
```

## License

MIT License
