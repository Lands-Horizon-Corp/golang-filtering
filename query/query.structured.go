package query

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
