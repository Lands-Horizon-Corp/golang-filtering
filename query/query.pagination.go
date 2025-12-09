package query

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (f *Pagination[T]) Pagination(
	db *gorm.DB,
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*PaginationResult[T], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &PaginationResult[T]{}, fmt.Errorf("failed to parse query: %w", err)
	}
	return f.StructuredPagination(db, filterRoot, pageIndex, pageSize, preloads...)
}

func (f *Pagination[T]) PaginationStructured(
	db *gorm.DB,

	context context.Context,
	ctx echo.Context,

	filter StructuredFilter,

	preloads ...string,
) (*PaginationResult[T], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &PaginationResult[T]{}, fmt.Errorf("failed to parse query: %w", err)
	}
	filterRoot.FieldFilters = append(filterRoot.FieldFilters, filter.FieldFilters...)
	filterRoot.Logic = LogicAnd
	if len(filterRoot.SortFields) == 0 && len(filter.SortFields) > 0 {
		filterRoot.SortFields = filter.SortFields
	}
	filterRoot.Preload = append(filterRoot.Preload, filter.Preload...)
	return f.StructuredPagination(db, filterRoot, pageIndex, pageSize, preloads...)
}

func (f *Pagination[T]) PaginationArray(
	db *gorm.DB,

	context context.Context,
	ctx echo.Context,

	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,

	preloads ...string,
) (*PaginationResult[T], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &PaginationResult[T]{}, fmt.Errorf("failed to parse query: %w", err)
	}
	for _, f := range filters {
		filterRoot.FieldFilters = append(filterRoot.FieldFilters, FieldFilter{
			Field:    f.Field,
			Value:    f.Value,
			Mode:     f.Op,
			DataType: DataTypeText,
		})
	}
	filterRoot.Logic = LogicAnd
	if len(filterRoot.SortFields) == 0 && len(sorts) > 0 {
		for _, s := range sorts {
			filterRoot.SortFields = append(filterRoot.SortFields, SortField(s))
		}
	}
	filterRoot.Preload = append(filterRoot.Preload, preloads...)
	return f.StructuredPagination(db, filterRoot, pageIndex, pageSize, preloads...)
}

func (f *Pagination[T]) PaginationNormal(
	db *gorm.DB,

	context context.Context,
	ctx echo.Context,

	filter *T,

	preloads ...string,
) (*PaginationResult[T], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &PaginationResult[T]{}, fmt.Errorf("failed to parse query: %w", err)
	}
	if filter != nil {
		db = db.Where(filter)
	}
	return f.StructuredPagination(db, filterRoot, pageIndex, pageSize, preloads...)
}
