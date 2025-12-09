package query

import (
	"fmt"

	"gorm.io/gorm"
)

func (f *Pagination[T]) Pagination(
	db *gorm.DB,
	filterRoot StructuredFilter,
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
	if f.Verbose {
		db = db.Debug()
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
