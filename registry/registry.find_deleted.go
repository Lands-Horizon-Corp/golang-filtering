package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) FindIncludeDeleted(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.NormalFindIncludeDeleted(r.client.WithContext(context), *fields, r.preload(preloads...)...)
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
	data, err := r.pagination.NormalFindLockIncludeDeleted(r.client.WithContext(context), *fields, r.preload(preloads...)...)
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
	data, err := r.pagination.ArrFindIncludeDeleted(r.client.WithContext(context), filters, sorts, r.preload(preloads...)...)
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
	data, err := r.pagination.ArrFindLockIncludeDeleted(r.client.WithContext(context), filters, sorts, r.preload(preloads...)...)
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
	data, err := r.pagination.StructuredFindIncludeDeleted(r.client.WithContext(context), filter)
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
	data, err := r.pagination.StructuredFindLockIncludeDeleted(r.client.WithContext(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

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

func (r *Registry[TData, TResponse, TRequest]) RawFindIncludeDeleted(
	context context.Context,
	filter *gorm.DB,
	preloads ...string,
) ([]*TData, error) {
	var db *gorm.DB
	if filter != nil {
		db = filter.Model(new(TData))
	} else {
		db = r.client.WithContext(context)
	}
	return r.pagination.RawFindIncludeDeleted(db, preloads...)
}

func (r *Registry[TData, TResponse, TRequest]) RawFindLockIncludeDeleted(
	context context.Context,
	filter *gorm.DB,
	preloads ...string,
) ([]*TData, error) {
	var db *gorm.DB
	if filter != nil {
		db = filter.Model(new(TData))
	} else {
		db = r.client.WithContext(context)
	}
	return r.pagination.RawFindLockIncludeDeleted(db, preloads...)
}

func (r *Registry[TData, TResponse, TRequest]) RawFindIncludeDeletedRaw(
	ctx context.Context,
	filter *gorm.DB,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.RawFindIncludeDeleted(ctx, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) RawFindLockIncludeDeletedRaw(
	ctx context.Context,
	filter *gorm.DB,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.RawFindLockIncludeDeleted(ctx, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}
