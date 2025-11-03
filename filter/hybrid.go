package filter

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// FilterHybrid intelligently chooses between in-memory (FilterDataQuery) and database (FilterDataGorm)
// filtering based on estimated table size. If estimated rows <= threshold, it fetches all data and
// uses in-memory filtering for better performance. Otherwise, it uses database filtering.
func (f *FilterHandler[T]) FilterHybrid(
	db *gorm.DB,
	threshold int64,
	filterRoot FilterRoot,
	pageIndex int,
	pageSize int,
) (*PaginationResult[T], error) {
	// Get table name from the model
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return nil, fmt.Errorf("failed to parse model: %w", err)
	}
	tableName := stmt.Table

	// Estimate row count based on database type
	estimatedRows, err := f.estimateTableRows(db, tableName)
	if err != nil {
		// If estimation fails, fall back to database filtering
		return f.FilterDataGorm(db, filterRoot, pageIndex, pageSize)
	}

	// Decide which strategy to use
	if estimatedRows <= threshold {
		// Use in-memory filtering for better performance on small datasets
		var allData []*T
		if err := db.Find(&allData).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch data for in-memory filtering: %w", err)
		}
		return f.FilterDataQuery(allData, filterRoot, pageIndex, pageSize)
	}

	// Use database filtering for large datasets
	return f.FilterDataGorm(db, filterRoot, pageIndex, pageSize)
}

// estimateTableRows returns an estimated row count for a table.
// It uses database-specific methods for fast estimation without scanning the entire table.
func (f *FilterHandler[T]) estimateTableRows(db *gorm.DB, tableName string) (int64, error) {
	// Get the database driver name
	dialectName := db.Name()

	type Estimate struct {
		Rows int64
	}
	var est Estimate

	switch dialectName {
	case "postgres":
		// PostgreSQL: Use pg_class for instant estimation
		query := fmt.Sprintf(`
			SELECT reltuples::BIGINT AS rows
			FROM pg_class
			WHERE relname = '%s'
		`, tableName)

		if err := db.Raw(query).Scan(&est).Error; err != nil {
			return 0, fmt.Errorf("postgres estimation failed: %w", err)
		}
		return est.Rows, nil

	case "mysql":
		// MySQL/MariaDB: Use INFORMATION_SCHEMA
		query := `
			SELECT TABLE_ROWS AS rows
			FROM INFORMATION_SCHEMA.TABLES
			WHERE TABLE_SCHEMA = DATABASE()
			  AND TABLE_NAME = ?
		`

		if err := db.Raw(query, tableName).Scan(&est).Error; err != nil {
			return 0, fmt.Errorf("mysql estimation failed: %w", err)
		}
		return est.Rows, nil

	case "sqlite":
		// SQLite: Query sqlite_stat1 if available, otherwise fall back to COUNT(*)
		// First try sqlite_stat1 (if ANALYZE has been run)
		query := fmt.Sprintf(`
			SELECT stat AS rows
			FROM sqlite_stat1
			WHERE tbl = '%s'
			LIMIT 1
		`, tableName)

		var statRows string
		if err := db.Raw(query).Scan(&statRows).Error; err == nil && statRows != "" {
			// Parse the first number from the stat string
			parts := strings.Split(statRows, " ")
			if len(parts) > 0 {
				var rowCount int64
				if _, err := fmt.Sscanf(parts[0], "%d", &rowCount); err == nil {
					return rowCount, nil
				}
			}
		}

		// Fall back to COUNT(*) for SQLite (it's usually fast for small databases)
		if err := db.Table(tableName).Count(&est.Rows).Error; err != nil {
			return 0, fmt.Errorf("sqlite count failed: %w", err)
		}
		return est.Rows, nil

	case "sqlserver":
		// SQL Server: Use system views
		query := fmt.Sprintf(`
			SELECT SUM(p.rows) AS rows
			FROM sys.partitions p
			INNER JOIN sys.objects o ON p.object_id = o.object_id
			WHERE o.name = '%s'
			  AND p.index_id IN (0, 1)
		`, tableName)

		if err := db.Raw(query).Scan(&est).Error; err != nil {
			return 0, fmt.Errorf("sqlserver estimation failed: %w", err)
		}
		return est.Rows, nil

	default:
		// Unsupported database: fall back to COUNT(*)
		if err := db.Table(tableName).Count(&est.Rows).Error; err != nil {
			return 0, fmt.Errorf("count fallback failed: %w", err)
		}
		return est.Rows, nil
	}
}
