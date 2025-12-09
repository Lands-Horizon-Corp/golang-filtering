package query

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (f *Pagination[T]) NormalPagination(
	db *gorm.DB,
	filter *T,
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
	query := db.Where(filter)
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

func (p *Pagination[T]) NormalCount(
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

func (p *Pagination[T]) NormalFind(
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
func (r *Pagination[T]) NormalFindLock(
	db *gorm.DB,
	filter T,
	preloads ...string,
) ([]*T, error) {
	var entities []*T
	db = db.Where(&filter)
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find entities with lock: %w", err)
	}
	return entities, nil
}
func (p *Pagination[T]) NormalFindOne(
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

func (p *Pagination[T]) NormalFindOneWithLock(
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

func (p *Pagination[T]) NormalExists(
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

func (p *Pagination[T]) NormalExistsByID(
	db *gorm.DB,
	id any,
) (bool, error) {
	var dummy int
	query := db.Where(fmt.Sprintf("%s = ?", p.columnDefaultID), id)
	err := query.Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check existence by ID: %w", err)
	}
	return dummy == 1, nil
}

func (p *Pagination[T]) NormalExistsIncludingDeleted(
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

func (p *Pagination[T]) NormalGetMax(
	db *gorm.DB,
	field string,
	filter T,
) (any, error) {
	var result any
	db = db.Where(&filter)
	row := db.Select(fmt.Sprintf("MAX(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get max of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) NormalGetMin(
	db *gorm.DB,
	field string,
	filter T,
) (any, error) {
	var result any
	db = db.Where(&filter)
	row := db.Select(fmt.Sprintf("MIN(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get min of %s: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) NormalGetMaxLock(
	tx *gorm.DB,
	field string,
	filter T,
) (any, error) {
	var result any
	tx = tx.Where(&filter).Clauses(clause.Locking{Strength: "UPDATE"})
	row := tx.Select(fmt.Sprintf("MAX(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get max of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) NormalGetMinLock(
	tx *gorm.DB,
	field string,
	filter T,
) (any, error) {
	var result any
	tx = tx.Where(&filter).Clauses(clause.Locking{Strength: "UPDATE"})
	row := tx.Select(fmt.Sprintf("MIN(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get min of %s with lock: %w", field, err)
	}
	return result, nil
}

func (p *Pagination[T]) NormalGetByID(
	db *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	err := db.First(&entity, fmt.Sprintf("%s = ?", p.columnDefaultID), id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get entity by ID %v: %w", id, err)
	}
	return &entity, nil
}

func (p *Pagination[T]) NormalGetByIDLock(
	tx *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	var entity T
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.First(&entity, fmt.Sprintf("%s = ?", p.columnDefaultID), id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get entity by ID %v with lock: %w", id, err)
	}
	return &entity, nil
}
func (p *Pagination[T]) NormalGetByIDIncludingDeleted(
	db *gorm.DB,
	id any,
	preloads ...string,
) (*T, error) {
	db = db.Unscoped()
	for _, preload := range preloads {
		db = db.Preload(preload)
	}

	var entity T
	err := db.First(&entity, fmt.Sprintf("%s = ?", p.columnDefaultID), id).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get entity by ID %v including deleted: %w", id, err)
	}

	return &entity, nil
}

func (f *Pagination[T]) NormalTabular(
	db *gorm.DB,
	filter T,
	getter func(data *T) map[string]any,
	preloads ...string,
) ([]byte, error) {
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	data, err := f.NormalFind(db, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %w", err)
	}
	return csvCreation(data, getter)
}
