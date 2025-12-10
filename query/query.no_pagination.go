package query

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (f *Pagination[T]) NoPagination(
	db *gorm.DB,
	ctx echo.Context,
	preloads ...string,
) ([]*T, error) {
	filterRoot, err := parseQueryNoPagination(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}
	return f.StructuredFind(db, filterRoot, preloads...)
}

func (f *Pagination[T]) NoPaginationStructured(
	db *gorm.DB,

	context context.Context,
	ctx echo.Context,

	filter StructuredFilter,

	preloads ...string,
) ([]*T, error) {
	filterRoot, err := parseQueryNoPagination(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}
	filterRoot.FieldFilters = append(filterRoot.FieldFilters, filter.FieldFilters...)
	filterRoot.Logic = LogicAnd
	if len(filterRoot.SortFields) == 0 && len(filter.SortFields) > 0 {
		filterRoot.SortFields = filter.SortFields
	}
	filterRoot.Preload = append(filterRoot.Preload, filter.Preload...)
	return f.StructuredFind(db, filterRoot, preloads...)
}

func (f *Pagination[T]) NoPaginationArray(
	db *gorm.DB,

	context context.Context,
	ctx echo.Context,

	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,

	preloads ...string,
) ([]*T, error) {
	filterRoot, err := parseQueryNoPagination(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}
	for _, f := range filters {
		filterRoot.FieldFilters = append(filterRoot.FieldFilters, FieldFilter{
			Field:    f.Field,
			Value:    f.Value,
			Mode:     f.Op,
			DataType: DetectDataType(f.Value),
		})
	}
	filterRoot.Logic = LogicAnd
	if len(filterRoot.SortFields) == 0 && len(sorts) > 0 {
		for _, s := range sorts {
			filterRoot.SortFields = append(filterRoot.SortFields, SortField(s))
		}
	}
	filterRoot.Preload = append(filterRoot.Preload, preloads...)
	return f.StructuredFind(db, filterRoot, preloads...)
}

func (f *Pagination[T]) NoPaginationNormal(
	db *gorm.DB,

	context context.Context,
	ctx echo.Context,

	filter *T,

	preloads ...string,
) ([]*T, error) {
	filterRoot, err := parseQueryNoPagination(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}
	if filter != nil {
		db = db.Where(filter)
	}
	return f.StructuredFind(db, filterRoot, preloads...)
}
