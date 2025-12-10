package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"gorm.io/gorm"
)

// FilterSQL is an alias for the query.ArrFilterSQL type used in array-style filters.
type FilterSQL = query.ArrFilterSQL

func (r *Registry[TData, TResponse, TRequest]) Exists(
	context context.Context,
	fields *TData,
) (bool, error) {
	return r.pagination.NormalExists(r.client.WithContext(context), *fields)
}

func (r *Registry[TData, TResponse, TRequest]) ArrExists(
	context context.Context,
	filters []query.ArrFilterSQL,
) (bool, error) {
	return r.pagination.ArrExists(r.client.WithContext(context), filters)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredExists(
	context context.Context,
	db *gorm.DB,
	filterRoot query.StructuredFilter,
) (bool, error) {
	return r.pagination.StructuredExists(db, filterRoot)
}

func (r *Registry[TData, TResponse, TRequest]) ExistsByID(
	context context.Context,
	id any,
) (bool, error) {
	return r.pagination.NormalExistsByID(r.client.WithContext(context), id)
}

func (r *Registry[TData, TResponse, TRequest]) ExistsByIDWithTx(
	context context.Context,
	tx *gorm.DB,
	id any,
) (bool, error) {
	return r.pagination.NormalExistsByID(tx, id)
}

func (r *Registry[TData, TResponse, TRequest]) ExistsIncludingDeleted(
	context context.Context,
	filters []FilterSQL,
) (bool, error) {
	return r.pagination.ArrExistsIncludingDeleted(r.client.WithContext(context), filters)
}

func (r *Registry[TData, TResponse, TRequest]) ExistsIncludingDeletedWithTx(
	context context.Context,
	tx *gorm.DB,
	filters []FilterSQL,
) (bool, error) {
	return r.pagination.ArrExistsIncludingDeleted(tx, filters)
}
