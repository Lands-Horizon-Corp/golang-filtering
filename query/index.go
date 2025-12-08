package query

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Pagination[T any] struct{}

func NewPagination[T any]() *Pagination[T] {
	return &Pagination[T]{}
}

func (f *Pagination[T]) StructuredPagination(
	db *gorm.DB,
	filterRoot StructuredFilter,
	pageIndex int,
	pageSize int,
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
	query := f.structuredQuery(db, filterRoot)
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}
	result.TotalSize = int(totalCount)
	result.TotalPage = (result.TotalSize + result.PageSize - 1) / result.PageSize
	offset := result.PageIndex * result.PageSize
	query = query.Offset(int(offset)).Limit(int(result.PageSize))
	var data []*T
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}
	result.Data = data
	return &result, nil
}

func (f *Pagination[T]) StructuredCount(
	db *gorm.DB,
	filterRoot StructuredFilter,
) (int64, error) {
	query := f.structuredQuery(db, filterRoot)
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}
	return totalCount, nil
}

func (f *Pagination[T]) StructuredFind(
	db *gorm.DB,
	filterRoot StructuredFilter,
) ([]*T, error) {
	query := f.structuredQuery(db, filterRoot)
	var data []*T
	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}
	return data, nil
}

func (p *Pagination[T]) StructuredFindLock(
	db *gorm.DB,
	filterRoot StructuredFilter,
	preloads ...string,
) ([]*T, error) {
	var entities []*T
	db = p.structuredQuery(db, filterRoot)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities with lock: %w", err)
	}
	return entities, nil
}

func (p *Pagination[T]) StructuredFindOne(
	db *gorm.DB,
	filterRoot StructuredFilter,
	preloads ...string,
) (*T, error) {
	var entity T
	db = p.structuredQuery(db, filterRoot)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	err := db.First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find entity: %w", err)
	}
	return &entity, nil
}

func (p *Pagination[T]) StructuredFindOneWithLock(
	db *gorm.DB,
	filterRoot StructuredFilter,
	preloads ...string,
) (*T, error) {
	var entity T
	db = p.structuredQuery(db, filterRoot)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	err := db.First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find entity with lock: %w", err)
	}
	return &entity, nil
}

func (p *Pagination[T]) StructuredExists(
	db *gorm.DB,
	filterRoot StructuredFilter,
) (bool, error) {
	db = p.structuredQuery(db, filterRoot)
	var dummy int
	err := db.Model(new(T)).Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) StructuredExistsByID(
	db *gorm.DB,
	id any,
) (bool, error) {
	var dummy int
	subQuery := db.Model(new(T)).Select("1").Where("id = ?", id).Limit(1)
	err := db.Raw("SELECT EXISTS (?)", subQuery).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence by ID: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) StructuredExistsIncludingDeleted(
	db *gorm.DB,
	filterRoot StructuredFilter,
) (bool, error) {
	db = db.Unscoped()
	db = p.structuredQuery(db, filterRoot)
	var dummy int
	err := db.Model(new(T)).Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence including deleted: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) StructuredGetMax(
	db *gorm.DB,
	field string,
	filterRoot StructuredFilter,
) (any, error) {
	var result any
	db = p.structuredQuery(db, filterRoot)
	err := db.Model(new(T)).Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get max of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) StructuredGetMin(
	db *gorm.DB,
	field string,
	filterRoot StructuredFilter,
) (any, error) {
	var result any
	db = p.structuredQuery(db, filterRoot)
	err := db.Model(new(T)).Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get min of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) StructuredGetMaxLock(
	tx *gorm.DB,
	field string,
	filterRoot StructuredFilter,
) (any, error) {
	var result any
	tx = p.structuredQuery(tx, filterRoot)
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.Model(new(T)).Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get max of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) StructuredGetMinLock(
	tx *gorm.DB,
	field string,
	filterRoot StructuredFilter,
) (any, error) {
	var result any
	tx = p.structuredQuery(tx, filterRoot)
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.Model(new(T)).Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get min of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) StructuredGetByID(
	db *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entity by ID %v: %w", id, err)
	}
	return &entity, nil
}

func (p *Pagination[T]) StructuredGetByIDLock(
	tx *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entity by ID %v with lock: %w", id, err)
	}
	return &entity, nil
}

func (f *Pagination[T]) StructuredTabular(
	db *gorm.DB,
	filterRoot StructuredFilter,
	getter func(data *T) map[string]any,
) ([]byte, error) {
	data, err := f.StructuredFind(db, filterRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	return csvCreation(data, getter)
}

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
	db = f.arrQuery(db, filters, sorts)
	var totalCount int64
	if err := db.Model(new(T)).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}
	result.TotalSize = int(totalCount)
	result.TotalPage = (result.TotalSize + result.PageSize - 1) / result.PageSize
	offset := result.PageIndex * result.PageSize
	db = db.Offset(offset).Limit(result.PageSize)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	var data []*T
	if err := db.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}
	result.Data = data
	return &result, nil
}

func (p Pagination[T]) ArrCount(
	db *gorm.DB,
	filters []ArrFilterSQL,
) (int64, error) {
	var count int64
	db = p.applyJoinsForFilters(db, filters)
	db = p.applySQLFilters(db, filters)
	if err := db.Model(new(T)).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count entities with %d filters: %w", len(filters), err)
	}
	return count, nil
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
		db = db.Order("updated_at DESC")
	}
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities with %d filters: %w", len(filters), err)
	}
	return entities, nil
}

func (r *Pagination[T]) ArrFindLock(
	db *gorm.DB,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	preloads ...string,
) ([]*T, error) {
	var entities []*T
	db = r.applyJoinsForFilters(db, filters)
	db = r.applySQLFilters(db, filters)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if len(sorts) > 0 {
		db = r.applySort(db, sorts)
	} else {
		db = db.Order("updated_at DESC")
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
		db = db.Order("updated_at DESC")
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
		db = db.Order("updated_at DESC")
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
	err := db.Model(new(T)).Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) ArrExistsByID(
	db *gorm.DB,
	id any,
) (bool, error) {
	var dummy int
	subQuery := db.Model(new(T)).Select("1").Where("id = ?", id).Limit(1)
	err := db.Raw("SELECT EXISTS (?)", subQuery).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence by ID: %w", err)
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
	err := db.Model(new(T)).Select("1").Limit(1).Scan(&dummy).Error
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
	err := db.Model(new(T)).Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	if err != nil {
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
	err := db.Model(new(T)).Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	if err != nil {
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
	err := tx.Model(new(T)).Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	if err != nil {
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
	err := tx.Model(new(T)).Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get min of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) ArrGetByID(
	db *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entity by ID %v: %w", id, err)
	}

	return &entity, nil
}

func (p *Pagination[T]) ArrGetByIDLock(
	tx *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entity by ID %v with lock: %w", id, err)
	}

	return &entity, nil
}

func (f *Pagination[T]) ArrTabular(
	db *gorm.DB,
	getter func(data *T) map[string]any,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	preloads ...string,
) ([]byte, error) {
	data, err := f.ArrFind(db, filters, sorts, preloads...)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	return csvCreation(data, getter)
}

func (f *Pagination[T]) FilterPagination(
	db *gorm.DB,
	filter T,
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
	query := db.Where(&filter)
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

func (p *Pagination[T]) Count(
	db *gorm.DB,
	filter T,
) (int64, error) {
	var count int64
	db = db.Where(&filter)
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}
	return count, nil
}

func (p *Pagination[T]) Find(
	db *gorm.DB,
	filter T,
	preloads ...string,
) ([]*T, error) {
	db = db.Where(&filter)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	var data []*T
	if err := db.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities: %w", err)
	}
	return data, nil
}
func (p *Pagination[T]) FilterFindOne(
	db *gorm.DB,
	filter T,
	preloads ...string,
) (*T, error) {
	db = db.Where(&filter)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	var entity T
	err := db.First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find entity: %w", err)
	}
	return &entity, nil
}

func (p *Pagination[T]) FilterFindOneWithLock(
	db *gorm.DB,
	filter T,
	preloads ...string,
) (*T, error) {
	db = db.Where(&filter)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	var entity T
	err := db.First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find entity with lock: %w", err)
	}
	return &entity, nil
}

func (p *Pagination[T]) Exists(
	db *gorm.DB,
	filter T,
) (bool, error) {
	db = db.Where(&filter)
	var dummy int
	err := db.Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) ExistsByID(
	db *gorm.DB,
	id any,
) (bool, error) {
	var dummy int
	subQuery := db.Where("id = ?", id).Select("1").Limit(1)
	err := db.Raw("SELECT EXISTS (?)", subQuery).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence by ID: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) ExistsIncludingDeleted(
	db *gorm.DB,
	filter T,
) (bool, error) {
	db = db.Unscoped().Where(&filter)
	var dummy int
	err := db.Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence including deleted: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) GetMax(
	db *gorm.DB,
	field string,
	filter T,
) (any, error) {
	var result any
	db = db.Where(&filter)
	err := db.Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get max of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) GetMin(
	db *gorm.DB,
	field string,
	filter T,
) (any, error) {
	var result any
	db = db.Where(&filter)
	err := db.Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get min of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) GetMaxLock(
	tx *gorm.DB,
	field string,
	filter T,
) (any, error) {
	var result any
	tx = tx.Where(&filter).Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.Select(fmt.Sprintf("MAX(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get max of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) GetMinLock(
	tx *gorm.DB,
	field string,
	filter T,
) (any, error) {
	var result any
	tx = tx.Where(&filter).Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.Select(fmt.Sprintf("MIN(%s)", field)).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get min of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) GetByID(
	db *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	err := db.First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entity by ID %v: %w", id, err)
	}
	return &entity, nil
}

func (p *Pagination[T]) GetByIDLock(
	tx *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.First(&entity, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entity by ID %v with lock: %w", id, err)
	}
	return &entity, nil
}

func (f *Pagination[T]) Tabular(
	db *gorm.DB,
	filter T,
	getter func(data *T) map[string]any,
) ([]byte, error) {
	data, err := f.Find(db, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	return csvCreation(data, getter)
}
