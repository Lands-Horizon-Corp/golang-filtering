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
	return r.pagination.NormalCount(r.client.WithContext(context), *fields)
}

func (r *Registry[TData, TResponse, TRequest]) ArrCount(
	context context.Context,
	filters []query.ArrFilterSQL,
) (int64, error) {
	return r.pagination.ArrCount(r.client.WithContext(context), filters)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredCount(
	context context.Context,
	db *gorm.DB,
	filterRoot query.StructuredFilter) (int64, error) {
	return r.pagination.StructuredCount(r.client.WithContext(context), filterRoot)
}

func (r *Registry[TData, TResponse, TRequest]) RawCount(
	context context.Context,
	filter *gorm.DB,
) (int64, error) {
	var db *gorm.DB
	if filter != nil {
		db = filter.Model(new(TData))
	} else {
		db = r.client.WithContext(context)
	}
	return r.pagination.RawCount(db)
}
