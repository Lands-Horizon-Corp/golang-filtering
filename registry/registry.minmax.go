package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) GetMax(ctx context.Context, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMax(r.Client(ctx), field, *filter)
}
func (r *Registry[TData, TResponse, TRequest]) GetMaxInt(ctx context.Context, field string, filter *TData) (int, error) {
	result, err := r.pagination.NormalGetMax(r.Client(ctx), field, *filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) GetMin(ctx context.Context, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMin(r.Client(ctx), field, *filter)
}
func (r *Registry[TData, TResponse, TRequest]) GetMinInt(ctx context.Context, field string, filter *TData) (int, error) {
	result, err := r.pagination.NormalGetMin(r.Client(ctx), field, *filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) GetMaxLock(ctx context.Context, tx *gorm.DB, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMaxLock(tx, field, *filter)
}

func (r *Registry[TData, TResponse, TRequest]) GetMaxLockInt(ctx context.Context, tx *gorm.DB, field string, filter *TData) (int, error) {
	result, err := r.pagination.NormalGetMaxLock(tx, field, *filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) GetMinLock(ctx context.Context, tx *gorm.DB, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMinLock(tx, field, *filter)
}

func (r *Registry[TData, TResponse, TRequest]) GetMinLockInt(ctx context.Context, field string, filter *TData) (int, error) {
	result, err := r.pagination.NormalGetMinLock(r.Client(ctx), field, *filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMax(ctx context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMax(r.Client(ctx), field, filter)
}
func (r *Registry[TData, TResponse, TRequest]) StructuredGetMaxInt(ctx context.Context, field string, filter query.StructuredFilter) (int, error) {
	result, err := r.pagination.StructuredGetMax(r.Client(ctx), field, filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMin(ctx context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMin(r.Client(ctx), field, filter)
}
func (r *Registry[TData, TResponse, TRequest]) StructuredGetMinInt(ctx context.Context, field string, filter query.StructuredFilter) (int, error) {
	result, err := r.pagination.StructuredGetMin(r.Client(ctx), field, filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMaxLock(ctx context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMaxLock(r.Client(ctx), field, filter)
}
func (r *Registry[TData, TResponse, TRequest]) StructuredGetMaxLockInt(ctx context.Context, field string, filter query.StructuredFilter) (int, error) {
	result, err := r.pagination.StructuredGetMaxLock(r.Client(ctx), field, filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMinLock(ctx context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMinLock(r.Client(ctx), field, filter)
}
func (r *Registry[TData, TResponse, TRequest]) StructuredGetMinLockInt(ctx context.Context, field string, filter query.StructuredFilter) (int, error) {
	result, err := r.pagination.StructuredGetMinLock(r.Client(ctx), field, filter)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMax(ctx context.Context, field string, filters []query.ArrFilterSQL) (any, error) {
	return r.pagination.ArrGetMax(r.Client(ctx), field, filters)
}
func (r *Registry[TData, TResponse, TRequest]) ArrGetMaxInt(ctx context.Context, field string, filters []query.ArrFilterSQL) (int, error) {
	result, err := r.pagination.ArrGetMax(r.Client(ctx), field, filters)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMin(ctx context.Context, field string, filters []query.ArrFilterSQL,
) (any, error) {
	return r.pagination.ArrGetMin(r.Client(ctx), field, filters)
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMinInt(ctx context.Context, field string, filters []query.ArrFilterSQL) (int, error) {
	result, err := r.pagination.ArrGetMin(r.Client(ctx), field, filters)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMaxLock(ctx context.Context, field string, filters []query.ArrFilterSQL) (any, error) {
	return r.pagination.ArrGetMaxLock(r.Client(ctx), field, filters)
}
func (r *Registry[TData, TResponse, TRequest]) ArrGetMaxLockInt(ctx context.Context, field string, filters []query.ArrFilterSQL) (int, error) {
	result, err := r.pagination.ArrGetMaxLock(r.Client(ctx), field, filters)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMinLock(ctx context.Context, field string, filters []query.ArrFilterSQL,
) (any, error) {
	return r.pagination.ArrGetMinLock(r.Client(ctx), field, filters)
}

func (r *Registry[TData, TResponse, TRequest]) ArrGetMinLockInt(ctx context.Context, field string, filters []query.ArrFilterSQL) (int, error) {
	result, err := r.pagination.ArrGetMinLock(r.Client(ctx), field, filters)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}
