package registry

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Pagination(
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*query.PaginationResult[TResponse], error) {
	filterRoot, pageIndex, pageSize, err := parseQuery(ctx)
	if err != nil {
		return &query.PaginationResult[TResponse]{}, fmt.Errorf("failed to parse query: %w", err)
	}
}

func (r *Registry[TData, TResponse, TRequest]) NormalPagination(
	context context.Context,
	fields *TData,
	pageIndex int,
	pageSize int,
	preloads ...string,
) (*query.PaginationResult[TData], error) {
	return r.pagination.Pagination(r.Client(context), fields, pageIndex, pageSize, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) ArrPagination(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	pageIndex int,
	pageSize int,
	preloads ...string,
) (*query.PaginationResult[TData], error) {
	return r.pagination.ArrPagination(r.Client(context), filters, sorts, pageIndex, pageSize, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredPagination(
	context context.Context,
	db *gorm.DB,
	filterRoot query.StructuredFilter,
	pageIndex int,
	pageSize int,
	preloads ...string,
) (*query.PaginationResult[TData], error) {
	return r.pagination.StructuredPagination(db, filterRoot, pageIndex, pageSize, r.preload(preloads...)...)
}
