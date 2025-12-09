package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
)

func (r *Registry[TData, TResponse, TRequest]) GetMax(ctx context.Context, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMax(r.Client(ctx), field, *filter)
}

func (r *Registry[TData, TResponse, TRequest]) GetMin(ctx context.Context, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMin(r.Client(ctx), field, *filter)
}

func (r *Registry[TData, TResponse, TRequest]) GetMaxLock(ctx context.Context, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMaxLock(r.Client(ctx), field, *filter)
}

func (r *Registry[TData, TResponse, TRequest]) GetMinLock(ctx context.Context, field string, filter *TData) (any, error) {
	return r.pagination.NormalGetMinLock(r.Client(ctx), field, *filter)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMax(ctx context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMax(r.Client(ctx), field, filter)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMin(ctx context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMin(r.Client(ctx), field, filter)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMaxLock(ctx context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMaxLock(r.Client(ctx), field, filter)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredGetMinLock(ctx context.Context, field string, filter query.StructuredFilter) (any, error) {
	return r.pagination.StructuredGetMinLock(r.Client(ctx), field, filter)
}
