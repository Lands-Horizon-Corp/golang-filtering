package query

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (f *Pagination[T]) ArrPagination(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	pageIndex int,
	pageSize int,
	preloads ...string,
) (*PaginationResult[T], error) {
	result := PaginationResult[T]{
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}
	if result.PageIndex < 0 {
		result.PageIndex = 0
	}
	if result.PageSize <= 0 {
		result.PageSize = 30
	}
	if f.verbose {
		db = db.Debug()
	}
	query := f.arrQuery(db, filters, sorts)
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}
	result.TotalSize = int(totalCount)
	result.TotalPage = (result.TotalSize + result.PageSize - 1) / result.PageSize
	offset := result.PageIndex * result.PageSize
	query = query.Offset(offset).Limit(result.PageSize)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	var data []*T
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}
	result.Data = data
	return &result, nil
}

func (p *Pagination[T]) ArrFind(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	preloads ...string,
) ([]*T, error) {
	var entities []*T
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = p.applySort(db, sorts)
	} else {
		db = db.Order(p.columnDefaultSort)
	}
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities with %d filters: %w", len(filters), err)
	}
	return entities, nil
}

func (p Pagination[T]) ArrCount(
	db *gorm.DB,
	filters []ArrFilterSQL,
) (int64, error) {
	var count int64
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count entities with %d filters: %w", len(filters), err)
	}
	return count, nil
}

func (p *Pagination[T]) ArrFindLock(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	preloads ...string,
) ([]*T, error) {
	var entities []*T
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = p.applySort(db, sorts)
	} else {
		db = db.Order(p.columnDefaultSort)
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities with %d filters and lock:: %w", len(filters), err)
	}
	return entities, nil
}

func (p *Pagination[T]) ArrFindOne(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	preloads ...string,
) (*T, error) {
	var entity T
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = p.applySort(db, sorts)
	} else {
		db = db.Order(p.columnDefaultSort)
	}
	err := db.First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find entity with %d filters: %w", len(filters), err)
	}
	return &entity, nil
}

func (p *Pagination[T]) ArrFindOneWithLock(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	preloads ...string,
) (*T, error) {
	var entity T
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = p.applySort(db, sorts)
	} else {
		db = db.Order(p.columnDefaultSort)
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	err := db.First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find entity with lock and %d filters: %w", len(filters), err)
	}
	return &entity, nil
}

func (p *Pagination[T]) ArrExists(
	db *gorm.DB,
	filters []ArrFilterSQL,
) (bool, error) {
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	var dummy int
	err := db.Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) ArrExistsIncludingDeleted(
	db *gorm.DB,
	filters []ArrFilterSQL,
) (bool, error) {
	db = db.Unscoped()
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	var dummy int
	err := db.Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence including deleted: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) ArrGetMax(
	db *gorm.DB,
	field string,
	filters []ArrFilterSQL,
) (any, error) {
	var result any
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	row := db.Select(fmt.Sprintf("MAX(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get max of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) ArrGetMin(
	db *gorm.DB,
	field string,
	filters []ArrFilterSQL,
) (any, error) {
	var result any
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	row := db.Select(fmt.Sprintf("MIN(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get min of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) ArrGetMaxLock(
	tx *gorm.DB,
	field string,
	filters []ArrFilterSQL,
) (any, error) {
	var result any
	tx = p.applyJoinsForFilters(tx, filters)
	tx = p.applySQLFilters(tx, filters)
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	row := tx.Select(fmt.Sprintf("MAX(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get max of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) ArrGetMinLock(
	tx *gorm.DB,
	field string,
	filters []ArrFilterSQL,
) (any, error) {
	var result any
	tx = p.applyJoinsForFilters(tx, filters)
	tx = p.applySQLFilters(tx, filters)
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	row := tx.Select(fmt.Sprintf("MIN(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get min of %s with lock: %w", field, err)
	}
	return result, nil
}

func (f *Pagination[T]) ArrTabular(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	getter func(data *T) map[string]any,
	preloads ...string,
) ([]byte, error) {
	data, err := f.ArrFind(db, filters, sorts, preloads...)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	return csvCreation(data, getter)
}

func (f *Pagination[T]) ArrRequestTabular(
	db *gorm.DB,
	ctx echo.Context,
	extraFilters []ArrFilterSQL,
	extraSorts []ArrFilterSortSQL,
	getter func(data *T) map[string]any,
	preloads ...string,
) ([]byte, error) {
	filterRoot, _, _, err := parseQuery(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}
	filters := make([]ArrFilterSQL, len(filterRoot.FieldFilters))
	for i, ff := range filterRoot.FieldFilters {
		filters[i] = ArrFilterSQL{
			Field: ff.Field,
			Value: ff.Value,
			Op:    ff.Mode,
		}
	}
	filters = append(filters, extraFilters...)
	sorts := make([]ArrFilterSortSQL, len(filterRoot.SortFields))
	for i, sf := range filterRoot.SortFields {
		sorts[i] = ArrFilterSortSQL{
			Field: sf.Field,
			Order: sf.Order,
		}
	}
	sorts = append(sorts, extraSorts...)
	data, err := f.ArrFind(db, filters, sorts, preloads...)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	return csvCreation(data, getter)
}

func (f *Pagination[T]) ArrStringTabular(
	db *gorm.DB,
	filterValue string,
	extraFilters []ArrFilterSQL,
	extraSorts []ArrFilterSortSQL,
	getter func(data *T) map[string]any,
	preloads ...string,
) ([]byte, error) {
	filterRoot, _, _, err := strParseQuery(filterValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query string: %w", err)
	}
	filters := make([]ArrFilterSQL, len(filterRoot.FieldFilters))
	for i, ff := range filterRoot.FieldFilters {
		filters[i] = ArrFilterSQL{
			Field: ff.Field,
			Value: ff.Value,
			Op:    ff.Mode,
		}
	}
	filters = append(filters, extraFilters...)
	sorts := make([]ArrFilterSortSQL, len(filterRoot.SortFields))
	for i, sf := range filterRoot.SortFields {
		sorts[i] = ArrFilterSortSQL{
			Field: sf.Field,
			Order: sf.Order,
		}
	}
	sorts = append(sorts, extraSorts...)
	data, err := f.ArrFind(db, filters, sorts, preloads...)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	return csvCreation(data, getter)
}

func (p *Pagination[T]) ArrFindIncludeDeleted(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	preloads ...string,
) ([]*T, error) {
	var entities []*T
	db = db.Unscoped()
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = p.applySort(db, sorts)
	} else {
		db = db.Order(p.columnDefaultSort)
	}
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities including deleted with %d filters: %w", len(filters), err)
	}
	return entities, nil
}

func (p *Pagination[T]) ArrFindLockIncludeDeleted(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	preloads ...string,
) ([]*T, error) {
	var entities []*T
	db = db.Unscoped()
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = p.applySort(db, sorts)
	} else {
		db = db.Order(p.columnDefaultSort)
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities including deleted with %d filters and lock: %w", len(filters), err)
	}
	return entities, nil
}
