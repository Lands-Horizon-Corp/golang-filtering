package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Pagination(
	context context.Context,
	fields *TData,
	pageIndex int,
	pageSize int,
) (*query.PaginationResult[TData], error) {
	return r.pagination.Pagination(r.Client(context), fields, pageIndex, pageSize, r.preloads...)
}

func (r *Registry[TData, TResponse, TRequest]) ArrPagination(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	pageIndex int,
	pageSize int,
) (*query.PaginationResult[TData], error) {
	return r.pagination.ArrPagination(r.Client(context), filters, sorts, pageIndex, pageSize, r.preloads...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredPagination(
	context context.Context,
	db *gorm.DB,
	filterRoot query.StructuredFilter,
	pageIndex int,
	pageSize int,
) (*query.PaginationResult[TData], error) {
	return r.pagination.StructuredPagination(db, filterRoot, pageIndex, pageSize, r.preloads...)
}
