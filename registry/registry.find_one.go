package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
)

func (r *Registry[TData, TResponse, TRequest]) FindOne(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.NormalFindOne(r.Client(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) FindOneWithLock(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.NormalFindOneWithLock(r.Client(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOne(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.ArrFindOne(r.Client(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOneWithLock(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.ArrFindOneWithLock(r.Client(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOne(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.StructuredFindOne(r.Client(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOneWithLock(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) (*TResponse, error) {
	entity, err := r.pagination.StructuredFindOneWithLock(r.Client(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return r.ToModel(entity), nil
}

// Raw variants return both the raw data entity and the converted response model.
func (r *Registry[TData, TResponse, TRequest]) FindOneRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.NormalFindOne(r.Client(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) FindOneWithLockRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.NormalFindOneWithLock(r.Client(context), *fields, r.preload(preloads...)...)
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
	entity, err := r.pagination.ArrFindOne(r.Client(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindOneWithLockRaw(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.ArrFindOneWithLock(r.Client(context), filters, sorts, r.preload(preloads...)...)
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
	entity, err := r.pagination.StructuredFindOne(r.Client(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindOneWithLockRaw(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) (*TData, *TResponse, error) {
	entity, err := r.pagination.StructuredFindOneWithLock(r.Client(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, nil, err
	}
	return entity, r.ToModel(entity), nil
}
