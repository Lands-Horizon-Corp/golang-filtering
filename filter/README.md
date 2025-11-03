# Filter Package

A flexible filtering, sorting, and pagination package for Go with **automatic field getter generation**.

## Installation

```bash
go get github.com/Lands-Horizon-Corp/golang-filtering/filter
```

## Features

- ✨ **Automatic getter generation using reflection** - No code generation needed!
- 🔍 Advanced filtering with multiple data types (text, number, bool, date, time)
- 📊 Sorting support
- 📄 Built-in pagination
- 🏷️ JSON tag support
- 🔗 Nested struct support
- 🎯 GORM integration

## Quick Start

### Automatic Getters (Recommended - Zero Configuration)

Simply pass your type to `NewFilter[T]()` and getters are automatically generated:

```go
package main

import (
    "github.com/Lands-Horizon-Corp/golang-filtering/filter"
    "gorm.io/gorm"
)

type User struct {
    Name     string `json:"name"`
    Age      int    `json:"age"`
    Email    string `json:"email"`
    IsActive bool   `json:"is_active"`
}

func main() {
    // Automatic getter generation - no setup needed!
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
        },
        SortFields: []filter.SortField{
            {Field: "age", Order: filter.FilterSortOrderDesc},
        },
    }

    // Use with GORM
    result, err := filterHandler.FilterDataGorm(db, filterRoot, 1, 10)
    // ... handle result
}
```

### How It Works

1. **Automatic Field Detection**: Uses reflection to detect all exported fields
2. **JSON Tag Support**: Respects `json` tags for field names
3. **Multiple Key Formats**: Supports both original field names and lowercase versions
4. **Nested Structs**: Automatically generates `parent.child` accessors

Example with nested structs:

```go
type Address struct {
    City    string `json:"city"`
    Country string `json:"country"`
}

type User struct {
    Name    string  `json:"name"`
    Address Address `json:"address"`
}

// You can filter by:
// - "name" or "Name"
// - "address.city" or "address.City"
// - "address.country" or "address.Country"
```

## Filter Modes

### Text Filters

- `equal`, `notEqual`
- `contains`, `notContains`
- `startsWith`, `endsWith`
- `isEmpty`, `isNotEmpty`

### Number Filters

- `equal`, `notEqual`
- `gt` (greater than), `gte` (greater than or equal)
- `lt` (less than), `lte` (less than or equal)
- `range`

### Boolean Filters

- `equal`, `notEqual`

### Date/DateTime Filters

- `equal`, `notEqual`
- `before`, `after`
- `gte`, `lte`, `gt`, `lt`
- `range`

### Time Filters

- `equal`, `notEqual`
- `before`, `after`
- `gte`, `lte`, `gt`, `lt`
- `range`

## Performance Notes

- **Reflection overhead is minimal**: Getters are generated once at initialization
- No runtime performance penalty during filtering
- Perfect for all use cases
