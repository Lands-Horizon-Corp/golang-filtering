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

// HybridCSV intelligently chooses between in-memory (DataQueryNoPageCSV) and database (GormNoPaginationCSV)
// filtering based on estimated table size, returning results as CSV bytes.
//
// IMPORTANT: Respects pre-existing WHERE conditions on the db parameter.
// - If DataQueryNoPageCSV is chosen (small dataset): fetches data using existing conditions, then filters in-memory and exports to CSV
// - If GormNoPaginationCSV is chosen (large dataset): combines existing conditions with filterRoot filters in SQL and exports to CSV
//
// Example with pre-existing conditions:
//
//	db := gormDB.Where("organization_id = ? AND branch_id = ?", orgID, branchID)
//	csvData, err := handler.HybridCSV(db, 10000, filterRoot)
//	// DataQueryNoPageCSV path: SELECT * FROM table WHERE organization_id = ? AND branch_id = ? (fetch all, filter in-memory, export CSV)
//	// GormNoPaginationCSV path: SELECT * FROM table WHERE organization_id = ? AND branch_id = ? AND [filterRoot conditions] (export CSV)
func (f *Handler[T]) HybridCSV(
	db *gorm.DB,
	threshold int,
	filterRoot Root,
) ([]byte, error) {
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
		// If estimation fails, fall back to database filtering with CSV export
		return f.GormNoPaginationCSV(db, filterRoot)
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
		return f.DataQueryNoPageCSV(allData, filterRoot)
	}

	// Use database filtering for large datasets with CSV export
	// GormNoPaginationCSV will combine existing WHERE conditions with filterRoot filters
	return f.GormNoPaginationCSV(db, filterRoot)
}

// HybridCSVWithPreset is a convenience method that combines preset conditions with HybridCSV.
// It accepts preset conditions as a struct and applies them before filtering, returning CSV results using hybrid strategy.
//
// Example usage:
//
//	type AccountTag struct {
//	    OrganizationID uint `gorm:"column:organization_id"`
//	    BranchID       uint `gorm:"column:branch_id"`
//	}
//
//	tag := &AccountTag{
//	    OrganizationID: user.OrganizationID,
//	    BranchID:       *user.BranchID,
//	}
//	csvData, err := handler.HybridCSVWithPreset(db, tag, 10000, filterRoot)
func (f *Handler[T]) HybridCSVWithPreset(
	db *gorm.DB,
	presetConditions any,
	threshold int,
	filterRoot Root,
) ([]byte, error) {
	// Apply preset conditions to db
	if presetConditions != nil {
		db = db.Where(presetConditions)
	}

	// Call HybridCSV with the modified db
	return f.HybridCSV(db, threshold, filterRoot)
}

// HybridCSVCustom intelligently chooses between in-memory (DataQueryNoPageCSVCustom) and database (GormNoPaginationCSVCustom)
// approaches for CSV export based on estimated table size, using a custom callback function for field mapping.
// For small tables (below threshold), it uses in-memory filtering with full dataset retrieval.
// For large tables (above threshold), it uses database-level filtering to minimize memory usage.
//
// Parameters:
//   - db: GORM database instance with any preset conditions
//   - threshold: row count threshold for switching between in-memory and database strategies
//   - filterRoot: filter configuration defining conditions, logic, and sorting
//   - customGetter: callback function that defines custom CSV field mapping
//
// Strategy Selection:
//   - If estimated table rows <= threshold: DataQueryNoPageCSVCustom (in-memory processing)
//   - If estimated table rows > threshold: GormNoPaginationCSVCustom (database processing)
//   - If estimation fails: Falls back to GormNoPaginationCSVCustom (database processing)
//
// Example usage:
//
//	csvData, err := handler.HybridCSVCustom(db, 10000, filterRoot, func(user *User) map[string]any {
//	    return map[string]any{
//	        "Employee Name": fmt.Sprintf("%s %s", user.FirstName, user.LastName),
//	        "Contact Email": user.Email,
//	        "Department": user.Department.Name,
//	        "Join Date": user.CreatedAt.Format("2006-01-02"),
//	    }
//	})
func (f *Handler[T]) HybridCSVCustom(
	db *gorm.DB,
	threshold int,
	filterRoot Root,
	customGetter func(*T) map[string]any,
) ([]byte, error) {
	// Get table name from the model
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return nil, fmt.Errorf("failed to parse model: %w", err)
	}
	tableName := stmt.Table

	// Estimate row count based on database type
	estimatedRows, err := f.estimateTableRows(db, tableName)
	if err != nil {
		// If estimation fails, fall back to database filtering with CSV export
		return f.GormNoPaginationCSVCustom(db, filterRoot, customGetter)
	}

	if int(estimatedRows) <= threshold {
		// Small table: use in-memory filtering with custom CSV export
		var allData []*T
		if err := db.Find(&allData).Error; err != nil {
			return nil, fmt.Errorf("failed to retrieve data: %w", err)
		}
		return f.DataQueryNoPageCSVCustom(allData, filterRoot, customGetter)
	} else {
		// Large table: use database filtering with custom CSV export
		return f.GormNoPaginationCSVCustom(db, filterRoot, customGetter)
	}
}

// HybridCSVCustomWithPreset is a convenience method that combines preset conditions with HybridCSVCustom.
// It applies preset conditions to the database query before intelligent strategy selection and CSV export.
//
// Parameters:
//   - db: GORM database instance
//   - presetConditions: struct or map with preset WHERE conditions to apply before filtering
//   - threshold: row count threshold for strategy selection
//   - filterRoot: filter configuration defining conditions, logic, and sorting
//   - customGetter: callback function that defines custom CSV field mapping
//
// Example usage:
//
//	type OrganizationFilter struct {
//	    OrganizationID uint
//	}
//
//	presetConditions := &OrganizationFilter{OrganizationID: user.OrganizationID}
//
//	csvData, err := handler.HybridCSVCustomWithPreset(db, presetConditions, 10000, filterRoot, func(user *User) map[string]any {
//	    return map[string]any{
//	        "ID": user.ID,
//	        "Name": user.Name,
//	        "Email": user.Email,
//	    }
//	})
func (f *Handler[T]) HybridCSVCustomWithPreset(
	db *gorm.DB,
	presetConditions any,
	threshold int,
	filterRoot Root,
	customGetter func(*T) map[string]any,
) ([]byte, error) {
	// Apply preset conditions to db
	if presetConditions != nil {
		db = db.Where(presetConditions)
	}

	// Call HybridCSVCustom with the modified db
	return f.HybridCSVCustom(db, threshold, filterRoot, customGetter)
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
