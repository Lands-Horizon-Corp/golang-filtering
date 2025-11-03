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

- **Automatic field getter generation** using reflection
- **Parallel processing** for in-memory filtering
- **Memory efficient** – zero data cloning, pointer-based operations
- **GORM integration** for database queries
- **Built-in security** – SQL injection, XSS protection, Command Injection, Null Byte Attacks
- **Rich filter modes** – text, number, boolean, date, time
- **Sorting** with multiple fields
- **Pagination** built-in
- **JSON tag support**
- **Nested struct support**

[Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Examples](#examples) • [Security](#security) • [Performance](#performance) • [API Reference](#api-reference)

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

    result, err := filterHandler.DataQuery(users, filterRoot, 1, 10)
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

    result, err := filterHandler.DataGorm(db, filterRoot, 1, 10)
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

### 3. Hybrid Filtering (`Hybrid`) – Auto-Switching

**Automatically chooses** between in-memory and database filtering based on table size.

```go
result, err := filterHandler.Hybrid(db, 10000, filterRoot, 1, 30)
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
    Field:          "age",
    Value:          filter.Range{From: 18, To: 65},
    Mode:           filter.ModeRange,
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
result, err := filterHandler.DataQuery(data, filterRoot, pageIndex, pageSize)
// or DataGorm / Hybrid
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
result, err := handler.DataQuery(users, filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "name",
            Value:    "'; DROP TABLE users--",
            Mode:     filter.ModeContains,
            DataType: filter.DataTypeText,
        },
    },
}, 1, 10)
// Returns safe results, no SQL execution

// Test XSS
result, err := handler.DataQuery(users, filter.Root{
    FieldFilters: []filter.FieldFilter{
        {
            Field:    "bio",
            Value:    "<img src=x onerror=alert(1)>",
            Mode:     filter.ModeContains,
            DataType: filter.DataTypeText,
        },
    },
}, 1, 10)
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
