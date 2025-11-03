# Database Seeder & Performance Testing

This package provides tools for seeding test data and stress-testing the GORM filtering system.

## Quick Start

### 1. Seed Test Data

```bash
# Small dataset (100 records) - for quick testing
make seed-small

# Default dataset (1,000 records)
make seed

# Large dataset (1,000,000 records) - for stress testing
make seed-stress

# Custom size
make seed-custom RECORDS=50000 BATCH=1000
```

### 2. Run Performance Tests

```bash
# Run all filter tests
make test

# Test specific page
make test-page PAGE=2 SIZE=100
```

### 3. Clean Database

```bash
make clean
```

## Manual Commands

### Seeding

```bash
# Using presets
go run cmd/seed/main.go -preset=small      # 100 records
go run cmd/seed/main.go -preset=default    # 1,000 records
go run cmd/seed/main.go -preset=stress     # 1,000,000 records

# Custom configuration
go run cmd/seed/main.go -count=10000 -batch=1000 -clear -progress

# Flags:
#   -count      Number of records to generate
#   -batch      Batch size for inserting records
#   -clear      Clear existing data before seeding
#   -progress   Show progress during seeding
#   -stats      Show statistics after seeding
```

### Testing

```bash
# Run performance tests
go run cmd/test/main.go

# With pagination
go run cmd/test/main.go -page=1 -page-size=50

# Flags:
#   -page       Page number to fetch (default: 1)
#   -page-size  Records per page (default: 50)
```

## Seeder Configuration

The seeder supports three preset configurations:

### Small Test Config

```go
RecordCount:   100
BatchSize:     50
ClearExisting: true
ShowProgress:  false
```

### Default Config

```go
RecordCount:   1000
BatchSize:     500
ClearExisting: false
ShowProgress:  true
```

### Stress Test Config

```go
RecordCount:   1_000_000
BatchSize:     10_000
ClearExisting: true
ShowProgress:  true
```

## Custom Usage in Code

```go
package main

import (
    "github.com/Lands-Horizon-Corp/golang-filtering/seeder"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

    // Create custom config
    config := seeder.SeederConfig{
        RecordCount:   50000,
        BatchSize:     5000,
        ClearExisting: true,
        ShowProgress:  true,
    }

    // Seed data
    userSeeder := seeder.NewUserSeeder(db, config)
    userSeeder.Seed()

    // Get statistics
    stats, _ := userSeeder.GetStats()
    fmt.Printf("Total users: %v\n", stats["total_users"])
}
```

## Performance Tests Included

The test suite includes:

1. **Simple text filter** - Name contains "John"
2. **Multiple AND filters** - Name, age, and active status
3. **Range filter** - Age between 30-40
4. **Date filter** - Created in last 30 days
5. **OR logic** - Name contains "John" OR "Smith"
6. **Complex filter** - Multiple conditions with range and text filters

Each test reports:

- Query execution time
- Total matching records
- Number of pages
- Records on current page
- Sample first record

## Generated Data Distribution

The seeder creates realistic data:

- **Names**: Mix of common names (John, Jane, etc.) and random names
- **Ages**:
  - 10% teens (13-17)
  - 65% young adults (18-35)
  - 15% middle aged (36-55)
  - 8% seniors (56-75)
  - 2% elderly (76-90)
- **Status**: 80% active users
- **Created dates**: Spread over last 3 years
- **Emails**: Generated from names with common domains

## Database Indexes

The seeder automatically creates indexes on:

- `name`
- `age`
- `email`
- `is_active`
- `created_at`

This ensures optimal query performance even with large datasets.

## Stress Testing Tips

For stress testing with 1M+ records:

1. Use SQLite with WAL mode for better write performance
2. Increase batch size (10,000-50,000) for faster seeding
3. Monitor memory usage during seeding
4. Test queries with different page sizes
5. Compare performance across different filter combinations

## Example Output

```bash
$ make seed-stress
Using STRESS TEST preset (1,000,000 records)

Clearing existing users...
Seeding 1000000 users in batches of 10000...
Progress: 1.00% (10000/1000000) - Elapsed: 2s
Progress: 2.00% (20000/1000000) - Elapsed: 4s
...
✓ Seeded 1000000 users in 3m42s
  Average: 4504.50 records/second

=== Database Statistics ===
  total_users: 1000000
  active_users: 800156
  inactive_users: 199844
  average_age: 35.42
  min_age: 13
  max_age: 90

✓ Seeding completed successfully!
```

```bash
$ make test
Test 1: Finding users with 'John' in name
  ✓ Query time: 15.2ms
  • Total matches: 45230
  • Total pages: 905
  • Current page: 1/905
  • Records on this page: 50
  • First record: John Adams (Age: 28, Active: true)
```
