package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Find(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.NormalFind(r.client.WithContext(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) FindWithLock(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.NormalFindLock(r.client.WithContext(context), *fields, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFind(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.ArrFind(r.client.WithContext(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindWithLock(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.ArrFindLock(r.client.WithContext(context), filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFind(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.StructuredFind(r.client.WithContext(context), filter)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindWithLock(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TData, error) {
	data, err := r.pagination.StructuredFindLock(r.client.WithContext(context), filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) FindRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.Find(context, fields, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) FindWithLockRaw(
	context context.Context,
	fields *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.FindWithLock(context, fields, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindRaw(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.ArrFind(context, filters, sorts, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrFindWithLockRaw(
	context context.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.ArrFindWithLock(context, filters, sorts, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindRaw(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.StructuredFind(context, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredFindWithLockRaw(
	context context.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.StructuredFindWithLock(context, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) RawFind(
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

	data, err := r.pagination.RawFind(db, preloads...)
	if err != nil {
		return nil, err
	}

	return data, nil
}
func (r *Registry[TData, TResponse, TRequest]) RawFindRaw(
	context context.Context,
	filter *gorm.DB,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.RawFind(context, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) RawFindWithLock(
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

	data, err := r.pagination.RawFindLock(db, preloads...)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *Registry[TData, TResponse, TRequest]) RawFindWithLockRaw(
	context context.Context,
	filter *gorm.DB,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.RawFindWithLock(context, filter, preloads...)
	if err != nil {
		return nil, err
	}
	return r.ToModels(data), nil
}
