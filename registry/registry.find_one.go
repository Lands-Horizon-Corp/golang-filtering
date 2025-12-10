package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) FindOne(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.NormalFindOne(r.client.WithContext(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) FindOneWithLock(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.NormalFindOneWithLock(r.client.WithContext(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOne(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.ArrFindOne(r.client.WithContext(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOneWithLock(
	context context.Context,
	tx *gorm.DB,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.ArrFindOneWithLock(tx, filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOne(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.StructuredFindOne(r.client.WithContext(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOneWithLock(
	context context.Context,
	tx *gorm.DB,
	filter query.StructuredFilter,
	preloads ...string,
) (*TData, error) {
	entity, err := r.pagination.StructuredFindOneWithLock(tx, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) FindOneRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.NormalFindOne(r.client.WithContext(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) FindOneWithLockRaw(
	context context.Context,
	tx *gorm.DB,
	fields *TData,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.NormalFindOneWithLock(tx, *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOneRaw(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.ArrFindOne(r.client.WithContext(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOneWithLockRaw(
	context context.Context,
	tx *gorm.DB,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.ArrFindOneWithLock(tx, filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOneRaw(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.StructuredFindOne(r.client.WithContext(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOneWithLockRaw(
	context context.Context,
	tx *gorm.DB,
	filter query.StructuredFilter,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.StructuredFindOneWithLock(tx, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}
