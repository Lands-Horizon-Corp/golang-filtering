package query

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (f *Pagination[T]) StructuredPaginationRaw(
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
	if err := db.Count(&totalCount).Error; err != nil {
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

func (f *Pagination[T]) PaginationRaw(
	db *gorm.DB,
	ctx echo.Context,
	rawQuery func(*gorm.DB) *gorm.DB,
	preloads ...string,
) (*PaginationResult[T], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &PaginationResult[T]{}, fmt.Errorf("failed to parse query: %w", err)
	}
	dbQuery := rawQuery(db)
	dbQuery = f.structuredQuery(dbQuery, filterRoot)
	return f.StructuredPaginationRaw(dbQuery, pageIndex, pageSize, preloads...)
}
