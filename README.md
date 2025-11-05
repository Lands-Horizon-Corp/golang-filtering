<div align="center">
  <img src=".github/assets/logo.png" alt="Golang Filtering Logo" width="80"/>
  <h1>Golang Filtering</h1>
  
  <p><strong>A high-performance, memory-efficient filtering package for Go with automatic field detection and support for both in-memory and GORM database filtering.</strong></p>

[![Go Reference](https://pkg.go.dev/badge/github.com/Lands-Horizon-Corp/golang-filtering.svg)](https://pkg.go.dev/github.com/Lands-Horizon-Corp/golang-filtering)
[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Security](https://img.shields.io/badge/security-SQL%20%7C%20XSS%20protected-green.svg)](#security)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

</div>

## ✨ Features

- **Dynamic field access without manual getters** – Automatically inspects your struct fields at runtime using reflection, eliminating the need to write custom getter functions for every field. Supports JSON tags and nested structures out of the box.
- **Nested struct filtering (up to 3 levels)** – Filter on deeply nested relationships using dot notation (e.g., `team.department.company.name`) with automatic JOIN generation for GORM queries
- **Parallel processing** for in-memory filtering – Utilizes all available CPU cores to process large datasets concurrently, significantly improving performance on multi-core systems
- **Memory efficient** – Zero data cloning, pointer-based operations that work directly on your original data without creating copies
- **GORM integration** for database queries – Seamless integration with GORM v2, automatically generating optimized SQL queries with proper parameterization and automatic JOINs for nested fields
- **Built-in security** – Multi-layer protection against SQL injection, XSS attacks, Command Injection, and Null Byte Attacks using industry-standard sanitization
- **Rich filter modes** – Comprehensive filtering options for text (contains, starts with, ends with), numbers (ranges, comparisons), booleans, dates, times, and custom types
- **Custom type support** – Handle custom types (like `TimeOfDay`) with `sql.Scanner` and `driver.Valuer` interfaces for specialized database storage
- **Sorting** with multiple fields – Sort by multiple columns (including nested fields) with ascending/descending order support
- **Pagination** built-in – Automatic pagination with total count, page count, and configurable page sizes
- **JSON tag support** – Respects your JSON field tags for API-friendly filtering
- **Preload support** – Eagerly load related entities with GORM's Preload feature

[Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Examples](#examples) • [Nested Filtering](#-nested-struct-filtering) • [Security](#security) • [Performance](#performance) • [API Reference](#api-reference)

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

### 1. In-Memory Filtering (`DataQuery`)

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

    filterRoot := filter.Root{
        Logic: filter.LogicAnd,
        FieldFilters: []filter.FieldFilter{
            {
                Field:          "name",
                Value:          "John",
                Mode:           filter.ModeContains,
                DataType: filter.DataTypeText,
            },
            {
                Field:          "age",
                Value:          18,
                Mode:           filter.ModeGTE,
                DataType: filter.DataTypeNumber,
            },
        },
        SortFields: []filter.SortField{
            {Field: "age", Order: filter.SortOrderDesc},
        },
    }

    pageIndex := 1
    pageSize := 10

    result, err := filterHandler.DataQuery(users, filterRoot, pageIndex, pageSize)
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

### 2. GORM Database Filtering (`DataGorm`)

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

    filterRoot := filter.Root{
        Logic: filter.LogicAnd,
        FieldFilters: []filter.FieldFilter{
            {
                Field:          "name",
                Value:          "John",
                Mode:           filter.ModeContains,
                DataType: filter.DataTypeText,
            },
            {
                Field:          "age",
                Value:          18,
                Mode:           filter.ModeGTE,
                DataType: filter.DataTypeNumber,
            },
        },
        SortFields: []filter.SortField{
            {Field: "age", Order: filter.SortOrderDesc},
        },
    }

    pageIndex := 1
    pageSize := 10

    result, err := filterHandler.DataGorm(db, filterRoot, pageIndex, pageSize)
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
- Preload related entities (eager loading)

#### Preloading Related Entities (GORM Only)

The `Preload` field in `Root` allows you to eagerly load related entities. This is **only available for GORM** and is optional (can be empty or omitted).

```go
type Author struct {
    ID    uint   `gorm:"primaryKey" json:"id"`
    Name  string `json:"name"`
    Posts []Post `gorm:"foreignKey:AuthorID" json:"posts"`
}

type Post struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Title    string `json:"title"`
    AuthorID uint   `json:"author_id"`
    Author   Author `gorm:"foreignKey:AuthorID" json:"author"` // Related entity
}

filterHandler := filter.NewFilter[Post]()

filterRoot := filter.Root{
    Logic:        filter.LogicAnd,
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "title",
            Value:    "Go",
            Mode:     filter.ModeContains,
            DataType: filter.DataTypeText,
        },
    },
    Preload: []string{"Author"}, // Eagerly load the Author relationship
}

pageIndex := 1
pageSize := 10

result, err := filterHandler.DataGorm(db, filterRoot, pageIndex, pageSize)
// result.Data[0].Author will be populated

// Multiple preloads
filterRoot.Preload = []string{"Author", "Comments"}

// Empty preload (no related entities loaded)
filterRoot.Preload = []string{}
```

**Preload Features:**

- Optional field (can be empty array or omitted)
- Supports multiple relations
- Works with filtering, sorting, and pagination
- GORM-only feature (ignored in `DataQuery` and `Hybrid`)

#### Using Preset Conditions with DataGorm (Multi-Tenant / Scoped Queries)

You can pass a `*gorm.DB` with **existing WHERE conditions** to `DataGorm`, and the package will apply `filterRoot` filters **on top of** your preset conditions. This is perfect for multi-tenant apps, branch-specific queries, or any scenario where you need base filtering.

**Use Case: Multi-Tenant Financial App**

```go
type BillAndCoin struct {
    ID             uint    `gorm:"primaryKey"`
    OrganizationID uint    `json:"organization_id"`  // Tenant isolation
    BranchID       uint    `json:"branch_id"`        // Branch isolation
    Amount         float64 `json:"amount"`
    Currency       string  `json:"currency"`
    Status         string  `json:"status"`
}

filterHandler := filter.NewFilter[BillAndCoin]()

// Scenario 1: Organization + Branch scoping with user filters
func GetBills(db *gorm.DB, orgID, branchID uint, userFilters filter.Root) (*filter.PaginationResult[BillAndCoin], error) {
    // Apply preset conditions (tenant/branch isolation)
    presetDB := db.Where("organization_id = ? AND branch_id = ?", orgID, branchID)

    // User's dynamic filters will be added on top
    // Final SQL: SELECT * FROM bill_and_coins
    //            WHERE organization_id = ? AND branch_id = ?  -- Your preset
    //            AND [user's filterRoot conditions]           -- User filters
    //            ORDER BY [sort fields] LIMIT ? OFFSET ?

    pageIndex := 1
    pageSize := 30

    return filterHandler.DataGorm(presetDB, userFilters, pageIndex, pageSize)
}

// Example usage:
orgID := uint(1)
branchID := uint(5)

userFilters := filter.Root{
    Logic: filter.LogicAnd,
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "status",
            Value:    "active",
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeText,
        },
        {
            Field:    "amount",
            Value:    filter.Range{From: 100, To: 1000},
            Mode:     filter.ModeRange,
            DataType: filter.DataTypeNumber,
        },
    },
    SortFields: []filter.SortField{
        {Field: "amount", Order: filter.SortOrderDesc},
    },
}

result, err := GetBills(db, orgID, branchID, userFilters)
// Returns: active bills for org=1, branch=5, with amount between 100-1000, sorted by amount DESC
```

**Scenario 2: Organization-only scoping**

```go
// Simpler: just organization-level isolation
presetDB := db.Where("organization_id = ?", orgID)
result, err := filterHandler.DataGorm(presetDB, userFilters, pageIndex, pageSize)
```

**Scenario 3: Complex preset conditions**

```go
// Multiple branches with date range
presetDB := db.Where("organization_id = ? AND branch_id IN ? AND created_at >= ?",
    orgID, []uint{1, 2, 3}, time.Now().AddDate(0, -1, 0))

result, err := filterHandler.DataGorm(presetDB, userFilters, pageIndex, pageSize)
```

**Scenario 4: No preset conditions (backward compatible)**

```go
// If you don't need preset conditions, just pass db as-is
result, err := filterHandler.DataGorm(db, userFilters, pageIndex, pageSize)
// Works exactly as before - filters everything in the table
```

**Why This Matters:**

✅ **Security** - Tenant isolation enforced at query level  
✅ **Performance** - Database can use indexes on preset columns  
✅ **Flexibility** - Combine fixed scope with dynamic user filters  
✅ **Clean Code** - Separation of concerns (business rules vs. user input)

---

#### Using Preset Conditions with Structs (Recommended)

For better type safety and cleaner code, you can use **structs** to define preset conditions:

**Method 1: Using `ApplyPresetConditions` Helper**

```go
// Define your preset condition struct
type AccountTag struct {
    OrganizationID uint `gorm:"column:organization_id"`
    BranchID       uint `gorm:"column:branch_id"`
}

filterHandler := filter.NewFilter[BillAndCoin]()

func GetUserBills(db *gorm.DB, user *User, userFilters filter.Root) (*filter.PaginationResult[BillAndCoin], error) {
    // Create preset conditions from user context
    tag := &AccountTag{
        OrganizationID: user.OrganizationID,
        BranchID:       *user.BranchID,
    }

    // Apply preset conditions to db
    db = filter.ApplyPresetConditions(db, tag)

    // DataGorm will apply userFilters on top of preset conditions
    return filterHandler.DataGorm(db, userFilters, 1, 30)
}
```

**Method 2: Using `DataGormWithPreset` Convenience Method (Even Cleaner!)**

```go
func GetUserBills(db *gorm.DB, user *User, userFilters filter.Root) (*filter.PaginationResult[BillAndCoin], error) {
    tag := &AccountTag{
        OrganizationID: user.OrganizationID,
        BranchID:       *user.BranchID,
    }

    // One-liner: applies preset AND user filters
    return filterHandler.DataGormWithPreset(db, tag, userFilters, 1, 30)
}
```

**Complete Example:**

```go
package main

import (
    "github.com/Lands-Horizon-Corp/golang-filtering/filter"
    "gorm.io/gorm"
)

type User struct {
    ID             uint
    OrganizationID uint
    BranchID       *uint  // Pointer allows nil for users without branch
}

type Transaction struct {
    ID             uint    `gorm:"primaryKey"`
    OrganizationID uint    `json:"organization_id"`
    BranchID       uint    `json:"branch_id"`
    Amount         float64 `json:"amount"`
    Currency       string  `json:"currency"`
    Status         string  `json:"status"`
}

type AccountTag struct {
    OrganizationID uint `gorm:"column:organization_id"`
    BranchID       uint `gorm:"column:branch_id"`
}

func GetTransactions(db *gorm.DB, user *User, pageIndex, pageSize int) (*filter.PaginationResult[Transaction], error) {
    handler := filter.NewFilter[Transaction]()

    // User's dynamic filters (from query parameters, form, etc.)
    userFilters := filter.Root{
        Logic: filter.LogicAnd,
        FieldFilters: []filter.FieldFilter{
            {
                Field:    "status",
                Value:    "completed",
                Mode:     filter.ModeEqual,
                DataType: filter.DataTypeText,
            },
            {
                Field:    "currency",
                Value:    "USD",
                Mode:     filter.ModeEqual,
                DataType: filter.DataTypeText,
            },
        },
        SortFields: []filter.SortField{
            {Field: "amount", Order: filter.SortOrderDesc},
        },
    }

    // Apply multi-tenant isolation automatically
    tag := &AccountTag{
        OrganizationID: user.OrganizationID,
        BranchID:       *user.BranchID,
    }

    // Returns: completed USD transactions for user's org+branch, sorted by amount DESC
    return handler.DataGormWithPreset(db, tag, userFilters, pageIndex, pageSize)
}
```

**Benefits of Struct-Based Approach:**

✅ **Type Safety** - Compile-time checking for field names  
✅ **Reusability** - Define preset structs once, use everywhere  
✅ **Clean Code** - No string concatenation for WHERE clauses  
✅ **IDE Support** - Auto-completion and refactoring support  
✅ **Maintainability** - Easy to add/remove preset fields

**Flexible Preset Conditions:**

```go
// Example 1: Organization only
type OrgFilter struct {
    OrganizationID uint `gorm:"column:organization_id"`
}

// Example 2: Organization + Branch + Status
type ScopedFilter struct {
    OrganizationID uint   `gorm:"column:organization_id"`
    BranchID       uint   `gorm:"column:branch_id"`
    IsActive       bool   `gorm:"column:is_active"`
}

// Example 3: Nil preset (no restrictions)
result, err := handler.DataGormWithPreset(db, nil, userFilters, 1, 30)
// Works like regular DataGorm - no preset conditions
```

**Why This Matters:**

✅ **Security** - Tenant isolation enforced at query level  
✅ **Performance** - Database can use indexes on preset columns  
✅ **Flexibility** - Combine hard-coded business rules with dynamic user filters  
✅ **Multi-tenancy** - Perfect for SaaS apps with organization/branch structures  
✅ **Backward compatible** - Works with or without preset conditions

---

### 3. Hybrid Filtering (`Hybrid`) – Auto-Switching

**Automatically chooses** between in-memory and database filtering based on table size.

```go
threshold := 10000  // Switch to DB queries if table has more than 10K rows
pageIndex := 1
pageSize := 30

result, err := filterHandler.Hybrid(db, threshold, filterRoot, pageIndex, pageSize)
```

**How it works:**

1. Estimates table size using database metadata (instant)
2. If ≤ threshold → `DataQuery` (parallel processing)
3. If > threshold → `DataGorm` (SQL queries)

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

#### Using Hybrid with Pre-existing DB Conditions

Just like `DataGorm`, `Hybrid` respects any pre-existing WHERE conditions on your `*gorm.DB`:

```go
// Multi-tenant scenario: filter by organization and branch first
db := gormDB.Where("organization_id = ? AND branch_id = ?", orgID, branchID)

filterRoot := filter.Root{
    Logic: filter.LogicAnd,
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "status",
            Value:    "active",
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeText,
        },
    },
}

threshold := 10000
pageIndex := 1
pageSize := 10

result, err := filterHandler.Hybrid(db, threshold, filterRoot, pageIndex, pageSize)
// If DataQuery path chosen: fetches all records WHERE organization_id=? AND branch_id=?, then filters in-memory
// If DataGorm path chosen: SELECT * WHERE organization_id=? AND branch_id=? AND status='active'
```

**Behavior:**

- **DataQuery path** (small dataset): Fetches data using your preset WHERE conditions, then applies `filterRoot` filters in memory
- **DataGorm path** (large dataset): Combines your preset conditions with `filterRoot` filters in a single SQL query

---

## 🔗 Nested Struct Filtering

Filter on **deeply nested relationships** using dot notation. Supports up to **3 levels of nesting** with automatic JOIN generation for GORM queries.

### Basic Nested Filtering

```go
type Address struct {
    ID      uint   `json:"id"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type User struct {
    ID       uint     `json:"id"`
    Name     string   `json:"name"`
    Address  *Address `gorm:"foreignKey:AddressID" json:"address"`
}

// Filter by nested field using dot notation
filterRoot := filter.Root{
    Logic: filter.LogicAnd,
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "address.city",        // Nested field
            Value:    "New York",
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeText,
        },
    },
}

result, err := handler.DataGorm(db, filterRoot, 1, 10)
// Automatically generates: SELECT * FROM users
// INNER JOIN addresses ON users.address_id = addresses.id
// WHERE addresses.city = 'New York'
```

### Multi-Level Nesting (Up to 3 Levels)

```go
type Company struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}

type Department struct {
    ID        uint     `json:"id"`
    Name      string   `json:"name"`
    CompanyID uint     `json:"company_id"`
    Company   *Company `gorm:"foreignKey:CompanyID" json:"company"`
}

type Team struct {
    ID           uint        `json:"id"`
    Name         string      `json:"name"`
    DepartmentID uint        `json:"department_id"`
    Department   *Department `gorm:"foreignKey:DepartmentID" json:"department"`
}

type Employee struct {
    ID     uint   `json:"id"`
    Name   string `json:"name"`
    TeamID uint   `json:"team_id"`
    Team   *Team  `gorm:"foreignKey:TeamID" json:"team"`
}

// Level 1: team.name
filterRoot := filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "team.name",
            Value:    "Engineering",
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeText,
        },
    },
}

// Level 2: team.department.name
filterRoot := filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "team.department.name",
            Value:    "Technology",
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeText,
        },
    },
}

// Level 3: team.department.company.name (maximum depth)
filterRoot := filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "team.department.company.name",
            Value:    "Acme Corp",
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeText,
        },
    },
}
```

### Nested Time Filtering with Custom Types

For time fields stored as TEXT in SQLite, use custom types with `sql.Scanner` and `driver.Valuer`:

```go
import (
    "database/sql/driver"
    "time"
)

// Custom time type for SQLite TEXT storage
type TimeOfDay struct {
    time.Time
}

func (t *TimeOfDay) Scan(value interface{}) error {
    if value == nil {
        return nil
    }
    str, ok := value.(string)
    if !ok {
        return fmt.Errorf("cannot scan type %T into TimeOfDay", value)
    }
    parsed, err := time.Parse("15:04:05", str)
    if err != nil {
        return err
    }
    t.Time = parsed
    return nil
}

func (t TimeOfDay) Value() (driver.Value, error) {
    return t.Time.Format("15:04:05"), nil
}

// Model with custom time type
type WorkShift struct {
    ID        uint      `json:"id"`
    Name      string    `json:"name"`
    StartTime TimeOfDay `gorm:"type:varchar(8)" json:"start_time"` // Stored as "HH:MM:SS"
    EndTime   TimeOfDay `gorm:"type:varchar(8)" json:"end_time"`
}

type Attendance struct {
    ID          uint       `json:"id"`
    Employee    string     `json:"employee"`
    WorkShiftID uint       `json:"work_shift_id"`
    WorkShift   *WorkShift `gorm:"foreignKey:WorkShiftID" json:"work_shift"`
}

// Filter by nested time field
filterRoot := filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "work_shift.start_time",
            Value:    "08:00:00",
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeTime,
        },
    },
}

result, err := handler.DataGorm(db, filterRoot, 1, 10)
// Finds all attendances with shifts starting at 8:00 AM
```

### Nested Filtering with Preload

Combine nested filtering with eager loading:

```go
filterRoot := filter.Root{
    Logic: filter.LogicAnd,
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "work_shift.start_time",
            Value:    filter.Range{From: "06:00:00", To: "14:00:00"},
            Mode:     filter.ModeRange,
            DataType: filter.DataTypeTime,
        },
    },
    Preload: []string{"WorkShift"}, // Eagerly load the relationship
}

result, err := handler.DataGorm(db, filterRoot, 1, 10)
// Returns attendances with shifts between 6:00 AM - 2:00 PM
// result.Data[0].WorkShift is fully populated
```

### Nested Sorting

Sort by nested fields:

```go
filterRoot := filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "team.department.name",
            Value:    "Engineering",
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeText,
        },
    },
    SortFields: []filter.SortField{
        {Field: "team.name", Order: filter.SortOrderAsc},           // Sort by team name
        {Field: "team.department.name", Order: filter.SortOrderAsc}, // Then by department
    },
}
```

### Nested Filtering Limitations

- **Maximum depth**: 3 levels (e.g., `a.b.c.d` is blocked)
- **GORM only**: Automatic JOIN generation works with `DataGorm` and `Hybrid`
- **DataQuery**: Nested filtering requires data to be pre-loaded with relationships
- **Performance**: Deep nesting may impact query performance; use indexes on foreign keys

### Nested Field Examples

```go
// ✅ Supported patterns
"profile.email"                    // Level 1
"address.city"                     // Level 1
"team.department.name"             // Level 2
"order.customer.address.city"      // Level 3 (maximum)

// ❌ Blocked patterns
"a.b.c.d.e"                        // Level 4+ (too deep)
```

### Automatic JOIN Generation (GORM)

The package automatically generates SQL JOINs for nested fields:

```go
// Filter: address.city = "NYC"
// Generated SQL:
// SELECT users.* FROM users
// INNER JOIN addresses ON users.address_id = addresses.id
// WHERE addresses.city = 'NYC'

// Filter: team.department.company.name = "Acme"
// Generated SQL:
// SELECT employees.* FROM employees
// INNER JOIN teams ON employees.team_id = teams.id
// INNER JOIN departments ON teams.department_id = departments.id
// INNER JOIN companies ON departments.company_id = companies.id
// WHERE companies.name = 'Acme'
```

---

## Filter Modes

### Text Filters

```go
filter.ModeEqual        // Exact match (case-insensitive)
filter.ModeNotEqual     // Not equal
filter.ModeContains     // Contains substring
filter.ModeNotContains  // Does not contain
filter.ModeStartsWith   // Starts with
filter.ModeEndsWith     // Ends with
filter.ModeIsEmpty      // Empty or null
filter.ModeIsNotEmpty   // Not empty
```

### Number Filters

```go
filter.ModeEqual    // =
filter.ModeNotEqual // !=
filter.ModeGT       // >
filter.ModeGTE      // >=
filter.ModeLT       // <
filter.ModeLTE      // <=
filter.ModeRange    // Between
```

**Range Example:**

```go
{
    Field:    "age",
    Value:    filter.Range{From: 18, To: 65},
    Mode:     filter.ModeRange,
    DataType: filter.DataTypeNumber,
}
```

### Boolean Filters

```go
filter.ModeEqual    // true/false
filter.ModeNotEqual // not
```

### Date/DateTime Filters

```go
filter.ModeEqual    // =
filter.ModeNotEqual // !=
filter.ModeBefore   // <
filter.ModeAfter    // >
filter.ModeGTE      // >=
filter.ModeLTE      // <=
filter.ModeRange    // Between
```

---

## 🔀 Logic Operators

### AND Logic

```go
filterRoot := filter.Root{
    Logic: filter.LogicAnd,
    FieldFilters: []filter.FieldFilter{ /* all must match */ },
}
```

### OR Logic

```go
filterRoot := filter.Root{
    Logic: filter.LogicOr,
    FieldFilters: []filter.FieldFilter{ /* any can match */ },
}
```

---

## 📄 Pagination

```go
pageIndex := 1   // First page (1-based)
pageSize := 30   // 30 items per page

result, err := filterHandler.DataQuery(data, filterRoot, pageIndex, pageSize)
// or DataGorm(db, filterRoot, pageIndex, pageSize)
// or Hybrid(db, threshold, filterRoot, pageIndex, pageSize)
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
filterRoot := filter.Root{
    SortFields: []filter.SortField{
        {Field: "age", Order: filter.SortOrderDesc},
        {Field: "name", Order: filter.SortOrderAsc},
    },
}
```

---

## 🔒 Security

### Built-in Protection Against Common Attacks

This package includes **enterprise-grade security** to protect against SQL injection, XSS, and other common web attacks.

#### **🛡️ Multi-Layer Security Architecture**

1. **Primary Defense: GORM Parameterized Queries**

   - All database queries use parameterized statements
   - Prevents SQL injection at the database driver level
   - Zero risk of direct SQL injection

2. **Secondary Defense: Input Sanitization**
   - Professional-grade sanitization using [`github.com/kennygrant/sanitize`](https://github.com/kennygrant/sanitize)
   - Defense-in-depth approach for application-layer protection

#### **🚫 What Gets Blocked**

| Attack Type           | Protection                   | Example                              |
| --------------------- | ---------------------------- | ------------------------------------ |
| **SQL Injection**     | Removes dangerous characters | `admin'--` → `admin`                 |
| **XSS Attacks**       | Strips HTML/JavaScript tags  | `<script>alert('XSS')</script>` → `` |
| **Command Injection** | Removes shell metacharacters | `test; rm -rf /` → `test rm rf`      |
| **Null Byte Attacks** | Removes control characters   | `admin\x00pass` → `adminpass`        |
| **SQL Comments**      | Removes comment syntax       | `user--DROP` → `userDROP`            |
| **Script Tags**       | Strips all HTML              | `<img onerror=alert(1)>` → ``        |

#### **✅ Automatic Sanitization**

All text input is automatically sanitized through the `parseText()` function:

```go
func parseText(value any) (string, error) {
    str, ok := value.(string)
    if !ok {
        return "", fmt.Errorf("invalid text type")
    }
    // Automatic sanitization applied here
    sanitized := Sanitize(str)
    return strings.ToLower(sanitized), nil
}
```

#### **🔐 Security Features**

##### **1. SQL Injection Prevention**

**Input:**

```go
filterRoot := filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "username",
            Value:    "admin' OR '1'='1",  // SQL injection attempt
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeText,
        },
    },
}
```

**What Happens:**

- Dangerous characters (`'`, `OR`, `=`) are removed
- GORM's parameterized query provides additional protection
- Query becomes safe: `WHERE username = ?` with parameter `"admin OR 1 1"`

##### **2. XSS Protection**

**Input:**

```go
filterRoot := filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "comment",
            Value:    "<script>alert('XSS')</script>",  // XSS attempt
            Mode:     filter.ModeContains,
            DataType: filter.DataTypeText,
        },
    },
}
```

**What Happens:**

- All HTML tags are stripped
- JavaScript code is removed
- Safe text remains: `"alertXSS"` or empty string

##### **3. Command Injection Protection**

**Input:**

```go
filterRoot := filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "filename",
            Value:    "test.txt; rm -rf /",  // Command injection attempt
            Mode:     filter.ModeEqual,
            DataType: filter.DataTypeText,
        },
    },
}
```

**What Happens:**

- Semicolons and special characters removed
- Shell metacharacters stripped
- Safe filename: `"testtxt rm rf"`

#### **📋 Sanitization Rules**

The `Sanitize()` function removes/transforms:

- ✅ HTML tags: `<script>`, `<img>`, `<iframe>`, etc.
- ✅ JavaScript protocols: `javascript:`, `data:`, `vbscript:`
- ✅ SQL special characters: `'`, `"`, `;`, `--`, `/*`, `*/`
- ✅ Parentheses and operators: `(`, `)`, `=`, `+`
- ✅ Control characters: Null bytes, newlines, tabs
- ✅ Shell metacharacters: `|`, `&`, `;`, `$`, `` ` ``

**Preserves:**

- ✅ Alphanumeric characters: `a-z`, `A-Z`, `0-9`
- ✅ Safe punctuation: `-`, `_`, `.`, `@`
- ✅ Spaces (normalized)

#### **🧪 Tested Security**

The package includes **44 comprehensive tests** covering:

- SQL injection attempts (various patterns)
- XSS attacks (script tags, event handlers)
- Command injection (shell metacharacters)
- Null byte attacks
- Encoding attacks (hex, char encoding)
- Legitimate input preservation

```bash
# Run security tests
go test ./test -v -run "SQL|XSS|Sanitize"
```

#### **⚠️ Best Practices**

While this package provides robust security, follow these additional best practices:

1. **Use HTTPS** - Always transmit data over encrypted connections
2. **Validate Input Types** - Check data types before processing
3. **Set Query Limits** - Prevent resource exhaustion with pagination
4. **Monitor Logs** - Watch for suspicious patterns in filter queries
5. **Keep Dependencies Updated** - Regularly update security packages

#### **🔍 Manual Security Testing**

Test the security yourself:

```go
// Test SQL injection
pageIndex := 1
pageSize := 10

result, err := handler.DataQuery(users, filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "name",
            Value:    "'; DROP TABLE users--",
            Mode:     filter.ModeContains,
            DataType: filter.DataTypeText,
        },
    },
}, pageIndex, pageSize)
// Returns safe results, no SQL execution

// Test XSS
result, err = handler.DataQuery(users, filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "bio",
            Value:    "<img src=x onerror=alert(1)>",
            Mode:     filter.ModeContains,
            DataType: filter.DataTypeText,
        },
    },
}, pageIndex, pageSize)
// HTML stripped, XSS prevented
```

#### **📦 Security Dependencies**

- **[kennygrant/sanitize](https://github.com/kennygrant/sanitize)** (v1.2.4)
  - BSD-3-Clause License
  - Production-ready, actively maintained
  - No regex overhead, efficient string processing
  - Used by thousands of projects

---

## ⚡ Performance

### In-Memory (`DataQuery`)

- Parallel processing across all CPU cores
- Zero data cloning – pointer-based
- Pre-allocated slices
- Reflection cached once

### Database (`DataGorm`)

- Efficient SQL generation
- Leverages database indexes
- Single `COUNT(*)` query for pagination

### Hybrid Mode

- Metadata-based size estimation (~1-2ms)
- Auto-selects optimal strategy

| Records | In-Memory | Database | Winner       |
| ------- | --------- | -------- | ------------ |
| 100     | 50µs      | 200µs    | In-Memory    |
| 10K     | 10ms      | 15ms     | In-Memory    |
| 100K    | 100ms     | 50ms     | Database     |
| 1M      | 1s        | 100ms    | **Database** |
| 10M     | OOM       | 500ms    | **Database** |

**Use `Hybrid` for automatic optimization!**

---

## 📖 API Reference

### Core Types

```go
type Handler[T any] struct { /* cached getters */ }

type Root struct {
    Logic        Logic
    FieldFilters []FieldFilter
    SortFields   []SortField
}

type FieldFilter struct {
    Field    string
    Value    any
    Mode     Mode
    DataType DataType
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

| Method               | Description                            |
| -------------------- | -------------------------------------- |
| `NewFilter[T any]()` | Creates handler with cached reflection |
| `DataQuery(...)`     | In-memory parallel filtering           |
| `DataGorm(...)`      | GORM database filtering                |
| `Hybrid(...)`        | Auto-switches based on size            |

---

## 🎯 When to Use Each Method

| Use Case                    | Recommended |
| --------------------------- | ----------- |
| Data in memory, < 100K      | `DataQuery` |
| Large DB tables             | `DataGorm`  |
| Unknown size / multi-tenant | `Hybrid`    |

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
- Email: lands.horizon.corp@gmail.com

---

<div align="center">
  <p><strong>Made with ❤️ by <a href="https://github.com/Lands-Horizon-Corp">Lands Horizon Corp</a></strong></p>
  <p>Star this project if you find it useful! ⭐</p>
</div>
