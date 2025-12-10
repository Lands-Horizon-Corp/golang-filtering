package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) GetMax(context context.Context, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMax(r.client.WithContext(context), field, *filter)
}
func (r *Registry[TData, TResponse, TRequest]) GetMaxInt(context context.Context, field string, filter *TData) (int, error) {
	result, err := r.pagination.NormalGetMax(r.client.WithContext(context), field, *filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) GetMin(context context.Context, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMin(r.client.WithContext(context), field, *filter)
}
func (r *Registry[TData, TResponse, TRequest]) GetMinInt(context context.Context, field string, filter *TData) (int, error) {
	result, err := r.pagination.NormalGetMin(r.client.WithContext(context), field, *filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) GetMaxLock(context context.Context, tx *gorm.DB, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMaxLock(tx, field, *filter)
}

func (r *Registry[TData, TResponse, TRequest]) GetMaxLockInt(context context.Context, tx *gorm.DB, field string, filter *TData) (int, error) {
	result, err := r.pagination.NormalGetMaxLock(tx, field, *filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) GetMinLock(context context.Context, tx *gorm.DB, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMinLock(tx, field, *filter)
}

func (r *Registry[TData, TResponse, TRequest]) GetMinLockInt(context context.Context, field string, filter *TData) (int, error) {
	result, err := r.pagination.NormalGetMinLock(r.client.WithContext(context), field, *filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMax(context context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMax(r.client.WithContext(context), field, filter)
}
func (r *Registry[TData, TResponse, TRequest]) StructuredGetMaxInt(context context.Context, field string, filter query.StructuredFilter) (int, error) {
	result, err := r.pagination.StructuredGetMax(r.client.WithContext(context), field, filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMin(context context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMin(r.client.WithContext(context), field, filter)
}
func (r *Registry[TData, TResponse, TRequest]) StructuredGetMinInt(context context.Context, field string, filter query.StructuredFilter) (int, error) {
	result, err := r.pagination.StructuredGetMin(r.client.WithContext(context), field, filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMaxLock(context context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMaxLock(r.client.WithContext(context), field, filter)
}
func (r *Registry[TData, TResponse, TRequest]) StructuredGetMaxLockInt(context context.Context, field string, filter query.StructuredFilter) (int, error) {
	result, err := r.pagination.StructuredGetMaxLock(r.client.WithContext(context), field, filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMinLock(context context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMinLock(r.client.WithContext(context), field, filter)
}
func (r *Registry[TData, TResponse, TRequest]) StructuredGetMinLockInt(context context.Context, field string, filter query.StructuredFilter) (int, error) {
	result, err := r.pagination.StructuredGetMinLock(r.client.WithContext(context), field, filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMax(context context.Context, field string, filters []query.ArrFilterSQL) (any, error) {
	return r.pagination.ArrGetMax(r.client.WithContext(context), field, filters)
}
func (r *Registry[TData, TResponse, TRequest]) ArrGetMaxInt(context context.Context, field string, filters []query.ArrFilterSQL) (int, error) {
	result, err := r.pagination.ArrGetMax(r.client.WithContext(context), field, filters)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMin(context context.Context, field string, filters []query.ArrFilterSQL,
) (any, error) {
	return r.pagination.ArrGetMin(r.client.WithContext(context), field, filters)
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMinInt(context context.Context, field string, filters []query.ArrFilterSQL) (int, error) {
	result, err := r.pagination.ArrGetMin(r.client.WithContext(context), field, filters)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMaxLock(context context.Context, field string, filters []query.ArrFilterSQL) (any, error) {
	return r.pagination.ArrGetMaxLock(r.client.WithContext(context), field, filters)
}
func (r *Registry[TData, TResponse, TRequest]) ArrGetMaxLockInt(context context.Context, field string, filters []query.ArrFilterSQL) (int, error) {
	result, err := r.pagination.ArrGetMaxLock(r.client.WithContext(context), field, filters)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMinLock(context context.Context, field string, filters []query.ArrFilterSQL,
) (any, error) {
	return r.pagination.ArrGetMinLock(r.client.WithContext(context), field, filters)
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMinLockInt(context context.Context, field string, filters []query.ArrFilterSQL) (int, error) {
	result, err := r.pagination.ArrGetMinLock(r.client.WithContext(context), field, filters)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}
