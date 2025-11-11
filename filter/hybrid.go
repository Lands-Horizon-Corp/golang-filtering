package filter

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Hybrid intelligently chooses between in-memory (DataQuery) and database (DataGorm)
// filtering based on estimated table size.
//
// IMPORTANT: Respects pre-existing WHERE conditions on the db parameter.
// - If DataQuery is chosen (small dataset): fetches data using existing conditions, then filters in-memory
// - If DataGorm is chosen (large dataset): combines existing conditions with filterRoot filters in SQL
//
// Example with pre-existing conditions:
//
//	db := gormDB.Where("organization_id = ? AND branch_id = ?", orgID, branchID)
//	result, err := handler.Hybrid(db, 10000, filterRoot, pageIndex, pageSize)
//	// DataQuery path: SELECT * FROM table WHERE organization_id = ? AND branch_id = ? (fetch all, filter in-memory)
//	// DataGorm path: SELECT * FROM table WHERE organization_id = ? AND branch_id = ? AND [filterRoot conditions]
func (f *Handler[T]) Hybrid(
	db *gorm.DB,
	threshold int,
	filterRoot Root,
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
	// NOTE: Estimation uses the full table, not filtered by existing WHERE conditions
	// This is intentional - we want to estimate total table size for strategy selection
	estimatedRows, err := f.estimateTableRows(db, tableName)
	if err != nil {
		// If estimation fails, fall back to database filtering
		return f.DataGorm(db, filterRoot, pageIndex, pageSize)
	}

	// Decide which strategy to use
	if estimatedRows <= int64(threshold) {
		// Use in-memory filtering for better performance on small datasets
		// IMPORTANT: This respects any pre-existing WHERE conditions on db
		// Example: if db has .Where("org_id = ?", 123), only records matching that will be fetched
		var allData []*T

		// Apply preload relationships before fetching data
		queryDB := db
		for _, relation := range filterRoot.Preload {
			queryDB = queryDB.Preload(relation)
		}

		if err := queryDB.Find(&allData).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch data for in-memory filtering: %w", err)
		}
		return f.DataQuery(allData, filterRoot, pageIndex, pageSize)
	}

	// Use database filtering for large datasets
	// DataGorm will combine existing WHERE conditions with filterRoot filters
	return f.DataGorm(db, filterRoot, pageIndex, pageSize)
}

// DataHybridNoPage intelligently chooses between in-memory (DataQueryNoPage) and database (DataGormNoPage)
// filtering based on estimated table size, returning results without pagination.
//
// IMPORTANT: Respects pre-existing WHERE conditions on the db parameter.
// - If DataQueryNoPage is chosen (small dataset): fetches data using existing conditions, then filters in-memory
// - If DataGormNoPage is chosen (large dataset): combines existing conditions with filterRoot filters in SQL
//
// Example with pre-existing conditions:
//
//	db := gormDB.Where("organization_id = ? AND branch_id = ?", orgID, branchID)
//	results, err := handler.DataHybridNoPage(db, 10000, filterRoot)
//	// DataQueryNoPage path: SELECT * FROM table WHERE organization_id = ? AND branch_id = ? (fetch all, filter in-memory)
//	// DataGormNoPage path: SELECT * FROM table WHERE organization_id = ? AND branch_id = ? AND [filterRoot conditions]
func (f *Handler[T]) DataHybridNoPage(
	db *gorm.DB,
	threshold int,
	filterRoot Root,
) ([]*T, error) {
	// Get table name from the model
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return nil, fmt.Errorf("failed to parse model: %w", err)
	}
	tableName := stmt.Table

	// Estimate row count based on database type
	// NOTE: Estimation uses the full table, not filtered by existing WHERE conditions
	// This is intentional - we want to estimate total table size for strategy selection
	estimatedRows, err := f.estimateTableRows(db, tableName)
	if err != nil {
		// If estimation fails, fall back to database filtering
		return f.DataGormNoPage(db, filterRoot)
	}

	// Decide which strategy to use
	if estimatedRows <= int64(threshold) {
		// Use in-memory filtering for better performance on small datasets
		// IMPORTANT: This respects any pre-existing WHERE conditions on db
		// Example: if db has .Where("org_id = ?", 123), only records matching that will be fetched
		var allData []*T

		// Apply preload relationships before fetching data
		queryDB := db
		for _, relation := range filterRoot.Preload {
			queryDB = queryDB.Preload(relation)
		}

		if err := queryDB.Find(&allData).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch data for in-memory filtering: %w", err)
		}
		return f.DataQueryNoPage(allData, filterRoot)
	}

	// Use database filtering for large datasets
	// DataGormNoPage will combine existing WHERE conditions with filterRoot filters
	return f.DataGormNoPage(db, filterRoot)
}

// estimateTableRows returns an estimated row count for a table.
// It uses database-specific methods for fast estimation without scanning the entire table.
// NOTE: This estimates the FULL table size, ignoring any WHERE conditions on the db parameter.
func (f *Handler[T]) estimateTableRows(db *gorm.DB, tableName string) (int64, error) {
	// Get the database driver name
	dialectName := db.Name()

	type Estimate struct {
		Rows int64
	}
	var est Estimate

	// Create a fresh session without any WHERE conditions for estimation
	// We want to estimate the full table size, not filtered results
	freshDB := db.Session(&gorm.Session{NewDB: true})

	switch dialectName {
	case "postgres":
		// PostgreSQL: Use pg_class for instant estimation
		query := fmt.Sprintf(`
			SELECT reltuples::BIGINT AS rows
			FROM pg_class
			WHERE relname = '%s'
		`, tableName)

		if err := freshDB.Raw(query).Scan(&est).Error; err != nil {
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

		if err := freshDB.Raw(query, tableName).Scan(&est).Error; err != nil {
			return 0, fmt.Errorf("mysql estimation failed: %w", err)
		}
		return est.Rows, nil

	case "sqlite":
		// SQLite: Query sqlite_stat1 if available, otherwise fall back to COUNT(*)
		// First try sqlite_stat1 (if ANALYZE has been run)
		// Note: Suppress error logging since sqlite_stat1 may not exist
		query := fmt.Sprintf(`
			SELECT stat AS rows
			FROM sqlite_stat1
			WHERE tbl = '%s'
			LIMIT 1
		`, tableName)

		var statRows string
		// Use a session with disabled logging for this query to avoid "no such table" warnings
		silentDB := freshDB.Session(&gorm.Session{Logger: freshDB.Logger.LogMode(1)}) // Silent mode
		err := silentDB.Raw(query).Scan(&statRows).Error

		// If sqlite_stat1 exists and has data, parse it
		if err == nil && statRows != "" {
			// Parse the first number from the stat string
			parts := strings.Split(statRows, " ")
			if len(parts) > 0 {
				var rowCount int64
				if _, scanErr := fmt.Sscanf(parts[0], "%d", &rowCount); scanErr == nil {
					return rowCount, nil
				}
			}
		}

		// Fall back to COUNT(*) for SQLite
		// This works even if sqlite_stat1 doesn't exist
		var countRows int64
		if err := freshDB.Table(tableName).Count(&countRows).Error; err != nil {
			return 0, fmt.Errorf("sqlite count failed: %w", err)
		}
		return countRows, nil

	case "sqlserver":
		// SQL Server: Use system views
		query := fmt.Sprintf(`
			SELECT SUM(p.rows) AS rows
			FROM sys.partitions p
			INNER JOIN sys.objects o ON p.object_id = o.object_id
			WHERE o.name = '%s'
			  AND p.index_id IN (0, 1)
		`, tableName)

		if err := freshDB.Raw(query).Scan(&est).Error; err != nil {
			return 0, fmt.Errorf("sqlserver estimation failed: %w", err)
		}
		return est.Rows, nil

	default:
		// Unsupported database: fall back to COUNT(*)
		if err := freshDB.Table(tableName).Count(&est.Rows).Error; err != nil {
			return 0, fmt.Errorf("count fallback failed: %w", err)
		}
		return est.Rows, nil
	}
}
