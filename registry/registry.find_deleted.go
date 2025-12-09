package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
)

func (r *Registry[TData, TResponse, TRequest]) FindIncludeDeleted(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.NormalFindIncludeDeleted(r.Client(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) FindWithLockIncludeDeleted(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.NormalFindLockIncludeDeleted(r.Client(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindIncludeDeleted(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.ArrFindIncludeDeleted(r.Client(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindWithLockIncludeDeleted(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.ArrFindLockIncludeDeleted(r.Client(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindIncludeDeleted(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.StructuredFindIncludeDeleted(r.Client(context), filter)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindWithLockIncludeDeleted(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.StructuredFindLockIncludeDeleted(r.Client(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ===== RAW versions =====

func (r *Registry[TData, TResponse, TRequest]) FindIncludeDeletedRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.FindIncludeDeleted(context, fields, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) FindWithLockIncludeDeletedRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.FindWithLockIncludeDeleted(context, fields, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindIncludeDeletedRaw(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.ArrFindIncludeDeleted(context, filters, sorts, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindWithLockIncludeDeletedRaw(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.ArrFindWithLockIncludeDeleted(context, filters, sorts, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindIncludeDeletedRaw(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.StructuredFindIncludeDeleted(context, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindWithLockIncludeDeletedRaw(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.StructuredFindWithLockIncludeDeleted(context, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}
