# Golang Filtering Package

<div align="center">
  <h1>🚀 Golang Filtering</h1>
  <p>A high-performance, memory-efficient filtering package for Go with automatic field detection and support for both in-memory and GORM database filtering.</p>

  [![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://go.dev/)
  [![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
  [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
</div>

## ✨ Features

- **Automatic field getter generation** using reflection
- **Parallel processing** for in-memory filtering
- **Memory efficient** – zero data cloning, pointer-based operations
- **GORM integration** for database queries
- **Rich filter modes** – text, number, boolean, date, time
- **Sorting** with multiple fields
- **Pagination** built-in
- **JSON tag support**
- **Nested struct support**

[Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Examples](#examples) • [Performance](#performance) • [API Reference](#api-reference)

---

## 📦 Installation

```bash
go get github.com/Lands-Horizon-Corp/golang-filtering/filter
```

**Requirements:**
- Go 1.18+ (for generics support)
- GORM v2 (optional, for database filtering)

---

## 🚀 Quick Start

### Define Your Model

```go
type User struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    Age       int       `json:"age"`
    Email     string    `json:"email"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
    Role      string    `json:"role"`
}
```

### Create Filter Handler

```go
filterHandler := filter.NewFilter[User]()
```

---

## 📚 Examples

### 1. In-Memory Filtering (`FilterDataQuery`)

Perfect for filtering data already loaded in memory with **parallel processing**.

```go
package main

import (
    "fmt"
    "github.com/Lands-Horizon-Corp/golang-filtering/filter"
)

func main() {
    users := []*filter.User{
        {Name: "John Doe", Age: 25, Email: "john@example.com", IsActive: true},
        {Name: "Jane Smith", Age: 30, Email: "jane@example.com", IsActive: true},
        {Name: "Bob Johnson", Age: 35, Email: "bob@example.com", IsActive: false},
    }

    filterHandler := filter.NewFilter[filter.User]()

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

    result, err := filterHandler.FilterDataQuery(users, filterRoot, 1, 10)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Total: %d, Page: %d/%d\n", result.TotalSize, result.PageIndex, result.TotalPage)
    for _, user := range result.Data {
        fmt.Printf("- %s (Age: %d)\n", user.Name, user.Age)
    }
}
```

**Key Benefits:**
- Parallel processing using all CPU cores
- Memory efficient – no data cloning
- Perfect for already-loaded data

---

### 2. GORM Database Filtering (`FilterDataGorm`)

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
    db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    db.AutoMigrate(&filter.User{})

    filterHandler := filter.NewFilter[filter.User]()

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
- Direct database queries – efficient SQL generation
- Built-in `COUNT` for pagination
- Database-level filtering (no loading all records)
- Perfect for large datasets

---

### 3. Hybrid Filtering (`FilterHybrid`) – Auto-Switching

**Automatically chooses** between in-memory and database filtering based on table size.

```go
result, err := filterHandler.FilterHybrid(db, 10000, filterRoot, 1, 30)
```

**How it works:**
1. Estimates table size using database metadata (instant)
2. If ≤ threshold → `FilterDataQuery` (parallel processing)
3. If > threshold → `FilterDataGorm` (SQL queries)

**Supported databases:**
- PostgreSQL (`pg_class`)
- MySQL/MariaDB (`INFORMATION_SCHEMA`)
- SQLite (`sqlite_stat1`)
- SQL Server (`sys.partitions`)

**Key Benefits:**
- Automatic optimization
- Fast estimation using metadata
- Smart switching
- Scalable from dev to production

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
filter.FilterModeEqual    // =
filter.FilterModeNotEqual // !=
filter.FilterModeGT       // >
filter.FilterModeGTE      // >=
filter.FilterModeLT       // <
filter.FilterModeLTE      // <=
filter.FilterModeRange    // Between
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
filter.FilterModeEqual    // true/false
filter.FilterModeNotEqual // not
```

### Date/DateTime Filters
```go
filter.FilterModeEqual    // =
filter.FilterModeNotEqual // !=
filter.FilterModeBefore   // <
filter.FilterModeAfter    // >
filter.FilterModeGTE      // >=
filter.FilterModeLTE      // <=
filter.FilterModeRange    // Between
```

---

## 🔀 Logic Operators

### AND Logic
```go
filterRoot := filter.FilterRoot{
    Logic: filter.FilterLogicAnd,
    Filters: []filter.Filter{ /* all must match */ },
}
```

### OR Logic
```go
filterRoot := filter.FilterRoot{
    Logic: filter.FilterLogicOr,
    Filters: []filter.Filter{ /* any can match */ },
}
```

---

## 📄 Pagination

```go
result, err := filterHandler.FilterDataQuery(data, filterRoot, pageIndex, pageSize)
// or FilterDataGorm / FilterHybrid
```

**Result includes:**
```go
result.Data       // []*T
result.TotalSize  // int
result.TotalPage  // int
result.PageIndex  // int (1-based)
result.PageSize   // int
```

**Defaults:**
- `pageIndex`: 1 (if ≤ 0)
- `pageSize`: 30 (if ≤ 0)

---

## 🎯 Sorting

```go
filterRoot := filter.FilterRoot{
    SortFields: []filter.SortField{
        {Field: "age", Order: filter.FilterSortOrderDesc},
        {Field: "name", Order: filter.FilterSortOrderAsc},
    },
}
```

---

## ⚡ Performance

### In-Memory (`FilterDataQuery`)
- Parallel processing across all CPU cores
- Zero data cloning – pointer-based
- Pre-allocated slices
- Reflection cached once

### Database (`FilterDataGorm`)
- Efficient SQL generation
- Leverages database indexes
- Single `COUNT(*)` query for pagination

### Hybrid Mode
- Metadata-based size estimation (~1-2ms)
- Auto-selects optimal strategy

| Records | In-Memory | Database | Winner     |
|--------|-----------|----------|------------|
| 100    | 50µs      | 200µs    | In-Memory  |
| 10K    | 10ms      | 15ms     | In-Memory  |
| 100K   | 100ms     | 50ms     | Database   |
| 1M     | 1s        | 100ms    | **Database** |
| 10M    | OOM       | 500ms    | **Database** |

**Use `FilterHybrid` for automatic optimization!**

---

## 📖 API Reference

### Core Types

```go
type Handler[T any] struct { /* cached getters */ }

type FilterRoot struct {
    Logic      FilterLogic
    Filters    []Filter
    SortFields []SortField
}

type Filter struct {
    Field          string
    Value          any
    Mode           FilterMode
    FilterDataType FilterDataType
}

type PaginationResult[T any] struct {
    Data      []*T
    TotalSize int
    TotalPage int
    PageIndex int
    PageSize  int
}
```

### Methods

| Method | Description |
|-------|-------------|
| `NewFilter[T any]()` | Creates handler with cached reflection |
| `FilterDataQuery(...)` | In-memory parallel filtering |
| `FilterDataGorm(...)` | GORM database filtering |
| `FilterHybrid(...)` | Auto-switches based on size |

---

## 🎯 When to Use Each Method

| Use Case | Recommended |
|--------|-------------|
| Data in memory, < 100K | `FilterDataQuery` |
| Large DB tables | `FilterDataGorm` |
| Unknown size / multi-tenant | `FilterHybrid` |

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)

```bash
git clone https://github.com/Lands-Horizon-Corp/golang-filtering.git
cd golang-filtering
go test ./...
```

---

## 📄 License

[MIT License](LICENSE)

---

## 💬 Support

- Issues: [GitHub Issues](https://github.com/Lands-Horizon-Corp/golang-filtering/issues)
- Email: support@landshorizon.com

---

<div align="center">
  <p><strong>Made with ❤️ by <a href="https://github.com/Lands-Horizon-Corp">Lands Horizon Corp</a></strong></p>
  <p>Star this project if you find it useful! ⭐</p>
</div>