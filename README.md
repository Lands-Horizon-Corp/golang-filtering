# Golang Filtering Package

A high-performance, memory-efficient filtering package for Go with automatic field detection and support for both in-memory and GORM database filtering.

## Features

- ✨ **Automatic field getter generation** using reflection
- 🚀 **Parallel processing** for in-memory filtering
- 💾 **Memory efficient** - zero data cloning, pointer-based operations
- �️ **GORM integration** for database queries
- 🔍 **Rich filter modes** - text, number, boolean, date, time
- 📊 **Sorting** with multiple fields
- 📄 **Pagination** built-in
- 🏷️ **JSON tag support**
- 🔗 **Nested struct support**

## Installation

```bash
go get github.com/Lands-Horizon-Corp/golang-filtering/filter
```

## Quick Start

### Define Your Model

```go
type User struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Age       int       `json:"age"`
    Email     string    `json:"email"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Create Filter Handler

```go
// Automatic getter generation - no setup needed!
filterHandler := filter.NewFilter[User]()
```

---

## Usage Examples

### 1. In-Memory Filtering (FilterData)

Perfect for filtering data already loaded in memory with parallel processing.

```go
package main

import (
    "fmt"
    "github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

func main() {
    // Your data
    users := []*User{
        {Name: "John Doe", Age: 25, Email: "john@example.com", IsActive: true},
        {Name: "Jane Smith", Age: 30, Email: "jane@example.com", IsActive: true},
        {Name: "Bob Johnson", Age: 35, Email: "bob@example.com", IsActive: false},
    }

    // Create filter handler
    filterHandler := filter.NewFilter[User]()

    // Define filters
    filterRoot := filter.FilterRoot{
        Logic: filter.FilterLogicAnd, // or FilterLogicOr
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

    // Apply filtering with pagination
    result, err := filterHandler.FilterDataQuery(users, filterRoot, 1, 10)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Total: %d, Page: %d/%d\n",
        result.TotalSize, result.PageIndex, result.TotalPage)

    for _, user := range result.Data {
        fmt.Printf("- %s (Age: %d)\n", user.Name, user.Age)
    }
}
```

**Key Benefits:**

- ⚡ Parallel processing using all CPU cores
- 💾 Memory efficient - no data cloning
- 🎯 Perfect for already-loaded data

---

### 2. GORM Database Filtering (FilterDataGorm)

Perfect for querying databases with efficient SQL generation.

```go
package main

import (
    "fmt"
    "log"
    "github.com/Lands-Horizon-Corp/golang-filtering/filter"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

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
        },
        SortFields: []filter.SortField{
            {Field: "age", Order: filter.FilterSortOrderDesc},
        },
    }

    // Execute database query with pagination
    result, err := filterHandler.FilterDataGorm(db, filterRoot, 1, 10)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Total: %d, Pages: %d\n", result.TotalSize, result.TotalPage)
    for _, user := range result.Data {
        fmt.Printf("- %s (Age: %d)\n", user.Name, user.Age)
    }
}
```

**Key Benefits:**

- 🗄️ Direct database queries - efficient SQL generation
- 📊 Built-in COUNT for pagination
- 🔍 Database-level filtering (no loading all records)
- 🚀 Perfect for large datasets (billions of records)

---

### 3. Hybrid Filtering (FilterHybrid) - Auto-Switching

**NEW!** Automatically chooses between in-memory and database filtering based on table size.

```go
package main

import (
    "fmt"
    "log"
    "github.com/Lands-Horizon-Corp/golang-filtering/filter"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    // Initialize database
    db, err := gorm.Open(postgres.Open("your-connection-string"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Create filter handler
    filterHandler := filter.NewFilter[User]()

    // Define filters
    filterRoot := filter.FilterRoot{
        Logic: filter.FilterLogicAnd,
        Filters: []filter.Filter{
            {
                Field:          "is_active",
                Value:          true,
                Mode:           filter.FilterModeEqual,
                FilterDataType: filter.FilterDataTypeBool,
            },
        },
    }

    // Hybrid filtering with 10,000 row threshold
    // - If table has ≤10k rows → uses in-memory filtering (faster)
    // - If table has >10k rows → uses database filtering (memory efficient)
    result, err := filterHandler.FilterHybrid(db, 10000, filterRoot, 1, 30)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Total: %d, Pages: %d\n", result.TotalSize, result.TotalPage)
}
```

**Key Benefits:**

- 🎯 **Automatic optimization** - no manual decision needed
- ⚡ **Fast estimation** - uses database metadata (pg_class, INFORMATION_SCHEMA)
- 🧠 **Smart switching** - best strategy for each dataset size
- 🔄 **Scalable** - works from development to production without code changes

**Supported Databases:**

- PostgreSQL (pg_class estimation)
- MySQL/MariaDB (INFORMATION_SCHEMA)
- SQLite (sqlite_stat1)
- SQL Server (sys.partitions)

**Recommended Thresholds:**

- Small tables (users, settings): 50,000
- Medium tables (orders, products): 10,000
- Large tables (logs, events): 1,000

**For more details, see:** [HYBRID-FILTER.md](./HYBRID-FILTER.md)

---

## Filter Modes

### Text Filters

```go
filter.FilterModeEqual        // Exact match (case-insensitive)
filter.FilterModeNotEqual     // Not equal
filter.FilterModeContains     // Contains substring
filter.FilterModeNotContains  // Does not contain
filter.FilterModeStartsWith   // Starts with
filter.FilterModeEndsWith     // Ends with
filter.FilterModeIsEmpty      // Empty or null
filter.FilterModeIsNotEmpty   // Not empty
```

### Number Filters

```go
filter.FilterModeEqual    // Equal to
filter.FilterModeNotEqual // Not equal to
filter.FilterModeGT       // Greater than
filter.FilterModeGTE      // Greater than or equal
filter.FilterModeLT       // Less than
filter.FilterModeLTE      // Less than or equal
filter.FilterModeRange    // Between two values
```

**Range Example:**

```go
{
    Field:          "age",
    Value:          filter.FilterRange{From: 18, To: 65},
    Mode:           filter.FilterModeRange,
    FilterDataType: filter.FilterDataTypeNumber,
}
```

### Boolean Filters

```go
filter.FilterModeEqual    // Equal to true/false
filter.FilterModeNotEqual // Not equal
```

### Date/DateTime Filters

```go
filter.FilterModeEqual    // Exact date match
filter.FilterModeNotEqual // Not equal
filter.FilterModeBefore   // Before date
filter.FilterModeAfter    // After date
filter.FilterModeGTE      // Greater than or equal
filter.FilterModeLTE      // Less than or equal
filter.FilterModeRange    // Between two dates
```

### Time Filters

```go
filter.FilterModeEqual    // Exact time
filter.FilterModeNotEqual // Not equal
filter.FilterModeBefore   // Before time
filter.FilterModeAfter    // After time
filter.FilterModeGTE      // Greater than or equal
filter.FilterModeLTE      // Less than or equal
filter.FilterModeRange    // Between two times
```

---

## Logic Operators

### AND Logic

All filters must match:

```go
filterRoot := filter.FilterRoot{
    Logic: filter.FilterLogicAnd,
    Filters: []filter.Filter{
        // All these must be true
    },
}
```

### OR Logic

Any filter can match:

```go
filterRoot := filter.FilterRoot{
    Logic: filter.FilterLogicOr,
    Filters: []filter.Filter{
        // Any of these can be true
    },
}
```

---

## Pagination

Both methods support pagination:

```go
result, err := filterHandler.FilterDataQuery(data, filterRoot, pageIndex, pageSize)
// or
result, err := filterHandler.FilterDataGorm(db, filterRoot, pageIndex, pageSize)

// Result contains:
result.Data       // []*T - Current page data
result.TotalSize  // int - Total matching records
result.TotalPage  // int - Total pages
result.PageIndex  // int - Current page (1-based)
result.PageSize   // int - Records per page
```

**Defaults:**

- `pageIndex`: defaults to 1 if <= 0
- `pageSize`: defaults to 30 if <= 0

---

## When to Use Each Method

### Use `FilterDataQuery` (In-Memory) When:

- ✅ Data is already loaded in memory
- ✅ You need parallel processing for speed
- ✅ Working with small to medium datasets (< 1 million records)
- ✅ You want to filter API responses or cached data

### Use `FilterDataGorm` (Database) When:

- ✅ Data is in a database
- ✅ Working with large datasets (millions/billions of records)
- ✅ You want to leverage database indexes
- ✅ Memory is limited
- ✅ You need efficient pagination over large result sets

### Use `FilterHybrid` (Auto-Switching) When:

- ✅ **Unknown dataset size** - let it decide automatically
- ✅ **Multi-tenant applications** - different tenants have different data volumes
- ✅ **Development to production** - small data in dev, large in prod
- ✅ **Want simplicity** - one method that always works optimally
- ✅ **Tables that grow over time** - automatic adaptation

---

## Performance

### In-Memory Filtering:

- ⚡ Parallel processing across all CPU cores
- 💾 Zero data cloning - only pointer manipulation
- 🎯 Pre-allocated slices prevent reallocation
- ⚙️ Reflection overhead is minimal (one-time at initialization)

### Database Filtering:

- 📊 Efficient SQL generation
- 🔍 Leverages database indexes
- 💡 Add indexes on frequently filtered fields for best performance

```bash
rm -f test.db && go run .
```
