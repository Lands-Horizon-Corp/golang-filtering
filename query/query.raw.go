package query

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (f *Pagination[T]) RawPagination(
	db *gorm.DB,
	pageIndex int,
	pageSize int,
	preloads ...string,
) (*PaginationResult[T], error) {
	result := PaginationResult[T]{PageIndex: pageIndex, PageSize: pageSize}

	if result.PageIndex < 0 {
		result.PageIndex = 0
	}
	if result.PageSize <= 0 {
		result.PageSize = 30
	}

	if f.verbose {
		db = db.Debug()
	}

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
func (f *Pagination[T]) RawFind(db *gorm.DB, preloads ...string) ([]*T, error) {
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	var data []*T
	if err := db.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch raw records: %w", err)
	}
	return data, nil
}

func (f *Pagination[T]) RawCount(db *gorm.DB) (int64, error) {
	var count int64
	if err := db.Model(new(T)).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count raw records: %w", err)
	}
	return count, nil
}

func (f *Pagination[T]) RawFindLock(db *gorm.DB, preloads ...string) ([]*T, error) {
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	var data []*T
	if err := db.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch raw records with lock: %w", err)
	}
	return data, nil
}

func (f *Pagination[T]) RawFindOne(db *gorm.DB, preloads ...string) (*T, error) {
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	var entity T
	err := db.First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch raw first record: %w", err)
	}
	return &entity, nil
}

func (f *Pagination[T]) RawFindOneWithLock(db *gorm.DB, preloads ...string) (*T, error) {
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
		return nil, fmt.Errorf("failed to fetch raw first record with lock: %w", err)
	}
	return &entity, nil
}

func (f *Pagination[T]) RawExists(db *gorm.DB) (bool, error) {
	var dummy int
	err := db.Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check raw existence: %w", err)
	}
	return dummy == 1, nil
}

func (f *Pagination[T]) RawExistsIncludingDeleted(db *gorm.DB) (bool, error) {
	var dummy int
	db = db.Unscoped()
	err := db.Select("1").Limit(1).Scan(&dummy).Error
	if err != nil {
		return false, fmt.Errorf("failed to check raw existence including deleted: %w", err)
	}
	return dummy == 1, nil
}

func (f *Pagination[T]) RawGetMax(db *gorm.DB, field string) (any, error) {
	var result any
	row := db.Model(new(T)).Select(fmt.Sprintf("MAX(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get max of %s: %w", field, err)
	}
	return result, nil
}

func (f *Pagination[T]) RawGetMin(db *gorm.DB, field string) (any, error) {
	var result any
	row := db.Model(new(T)).Select(fmt.Sprintf("MIN(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get min of %s: %w", field, err)
	}
	return result, nil
}

func (f *Pagination[T]) RawGetMaxLock(db *gorm.DB, field string) (any, error) {
	var result any
	row := db.Model(new(T)).Clauses(clause.Locking{Strength: "UPDATE"}).Select(fmt.Sprintf("MAX(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get max of %s with lock: %w", field, err)
	}
	return result, nil
}

func (f *Pagination[T]) RawGetMinLock(db *gorm.DB, field string) (any, error) {
	var result any
	row := db.Model(new(T)).Clauses(clause.Locking{Strength: "UPDATE"}).Select(fmt.Sprintf("MIN(%s)", field)).Row()
	if err := row.Scan(&result); err != nil {
		return nil, fmt.Errorf("failed to get min of %s with lock: %w", field, err)
	}
	return result, nil
}

func (f *Pagination[T]) RawTabular(db *gorm.DB, getter func(data *T) map[string]any, preloads ...string) ([]byte, error) {
	data, err := f.RawFind(db, preloads...)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw data for tabular: %w", err)
	}
	return csvCreation(data, getter)
}

func (f *Pagination[T]) RawRequestTabular(db *gorm.DB, ctx echo.Context, getter func(data *T) map[string]any, preloads ...string) ([]byte, error) {
	data, err := f.RawFind(db, preloads...)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw data for request tabular: %w", err)
	}
	return csvCreation(data, getter)
}

func (f *Pagination[T]) RawStringTabular(db *gorm.DB, str string, getter func(data *T) map[string]any, preloads ...string) ([]byte, error) {
	data, err := f.RawFind(db, preloads...)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw data for string tabular: %w", err)
	}
	return csvCreation(data, getter)
}

func (f *Pagination[T]) RawFindIncludeDeleted(db *gorm.DB, preloads ...string) ([]*T, error) {
	db = db.Unscoped()
	return f.RawFind(db, preloads...)
}

func (f *Pagination[T]) RawFindLockIncludeDeleted(db *gorm.DB, preloads ...string) ([]*T, error) {
	db = db.Unscoped()
	return f.RawFindLock(db, preloads...)
}
