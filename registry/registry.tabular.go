package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
)

func (r *Registry[TData, TResponse, TRequest]) NormalTabular(
	context context.Context,
	filter TData,
	getter func(data *TData) map[string]any,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.NormalTabular(r.Client(context), filter, getter, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) ArrTabular(
	context context.Context,
	getter func(data *TData) map[string]any,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.ArrTabular(r.Client(context), getter, filters, sorts, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredTabular(
	context context.Context,
	filter query.StructuredFilter,
	getter func(data *TData) map[string]any,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.StructuredTabular(r.Client(context), filter, getter, r.preload(preloads...)...)
}
