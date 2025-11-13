# Golang Filtering

A high-performance filtering package for Go with support for both in-memory and database filtering.

## Installation

```bash
go get github.com/Lands-Horizon-Corp/golang-filtering/filter
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

type User struct {
    ID       uint   `json:"id"`
    Name     string `json:"name"`
    Age      int    `json:"age"`
    Email    string `json:"email"`
    IsActive bool   `json:"is_active"`
}

func main() {
    users := []*User{
        {ID: 1, Name: "John Doe", Age: 25, Email: "john@example.com", IsActive: true},
        {ID: 2, Name: "Jane Smith", Age: 30, Email: "jane@example.com", IsActive: false},
    }

    handler := filter.NewFilter[User](filter.GolangFilteringConfig{})

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
                Field:    "is_active",
                Value:    true,
                Mode:     filter.ModeEqual,
                DataType: filter.DataTypeBool,
            },
        },
    }

    result, err := handler.DataQuery(users, filterRoot, 1, 10)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d users\n", result.TotalSize)
    for _, user := range result.Data {
        fmt.Printf("- %s\n", user.Name)
    }
}
```

## Features

- **In-Memory Filtering** - Filter data already loaded in memory
- **Database Filtering** - Filter directly at database level with GORM
- **Hybrid Mode** - Automatically choose between in-memory and database filtering
- **CSV Export** - Export filtered results to CSV format
- **Custom CSV** - Define custom field mappings for CSV export
- **Parallel Processing** - Multi-core processing for in-memory filtering
- **Type Safety** - Full Go generics support
- **Security** - Built-in protection against SQL injection and XSS

## Methods

### In-Memory Filtering
```go
// Filter data in memory
result, err := handler.DataQuery(data, filterRoot, pageIndex, pageSize)

// Export to CSV
csvData, err := handler.DataQueryNoPageCSV(data, filterRoot)

// Custom CSV with field mapping
csvData, err := handler.DataQueryNoPageCSVCustom(data, filterRoot, func(item *T) map[string]any {
    return map[string]any{
        "Custom Name": item.Name,
        "Custom Email": item.Email,
    }
})
```

### Database Filtering
```go
// Filter at database level
result, err := handler.DataGorm(db, filterRoot, pageIndex, pageSize)

// Export to CSV
csvData, err := handler.GormNoPaginationCSV(db, filterRoot)

// With preset conditions
csvData, err := handler.GormNoPaginationCSVWithPreset(db, presetConditions, filterRoot)

// Custom CSV
csvData, err := handler.GormNoPaginationCSVCustom(db, filterRoot, customMapper)
```

### Hybrid Filtering
```go
// Auto-choose strategy based on table size
threshold := 10000
result, err := handler.Hybrid(db, threshold, filterRoot, pageIndex, pageSize)

// Hybrid CSV export
csvData, err := handler.HybridCSV(db, threshold, filterRoot)
csvData, err := handler.HybridCSVCustom(db, threshold, filterRoot, customMapper)
```

## Filter Modes

### Text
- `ModeEqual`, `ModeNotEqual`
- `ModeContains`, `ModeNotContains`
- `ModeStartsWith`, `ModeEndsWith`
- `ModeIsEmpty`, `ModeIsNotEmpty`

### Number
- `ModeEqual`, `ModeNotEqual`
- `ModeGT`, `ModeGTE`, `ModeLT`, `ModeLTE`
- `ModeRange`

### Boolean
- `ModeEqual`, `ModeNotEqual`

### Date/Time
- `ModeEqual`, `ModeNotEqual`
- `ModeBefore`, `ModeAfter`
- `ModeRange`

## License

MIT License