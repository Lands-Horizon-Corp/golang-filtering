package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Count(
	context context.Context,
	fields *TData,
) (int64, error) {
	return r.pagination.Count(r.Client(context), *fields)
}

func (r *Registry[TData, TResponse, TRequest]) ArrCount(
	context context.Context,
	filters []query.ArrFilterSQL,
) (int64, error) {
	return r.pagination.ArrCount(r.Client(context), filters)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredCount(
	context context.Context,
	db *gorm.DB,
	filterRoot query.StructuredFilter) (int64, error) {
	return r.pagination.StructuredCount(db, filterRoot)
}
