package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/labstack/echo/v4"
)

func (r *Registry[TData, TResponse, TRequest]) Tabular(
	context context.Context,
	filter TData,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.NormalTabular(r.Client(context), filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) RequestTabular(
	ctx context.Context,
	echoCtx echo.Context,
	filter *TData,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.NormalRequestTabular(r.Client(ctx), echoCtx, filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StringTabular(
	ctx context.Context,
	filterValue string,
	filter *TData,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.NormalStringTabular(r.Client(ctx), filterValue, filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) ArrTabular(
	context context.Context,

	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]byte, error) {
	return r.pagination.ArrTabular(r.Client(context), filters, sorts, r.tabular, r.preload(preloads...)...)
}
func (r *Registry[TData, TResponse, TRequest]) ArrRequestTabular(
	ctx context.Context,
	echoCtx echo.Context,
	extraFilters []query.ArrFilterSQL,
	extraSorts []query.ArrFilterSortSQL,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.ArrRequestTabular(r.Client(ctx), echoCtx, extraFilters, extraSorts, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) ArrStringTabular(
	ctx context.Context,
	filterValue string,
	extraFilters []query.ArrFilterSQL,
	extraSorts []query.ArrFilterSortSQL,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.ArrStringTabular(r.Client(ctx), filterValue, extraFilters, extraSorts, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredTabular(
	context context.Context,
	filter query.StructuredFilter,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.StructuredTabular(r.Client(context), filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredRequestTabular(
	ctx context.Context,
	echoCtx echo.Context,
	filter query.StructuredFilter,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.StructuredRequestTabular(r.Client(ctx), echoCtx, filter, r.tabular, r.preload(preloads...)...)
}

func (r *Registry[TData, TResponse, TRequest]) StructuredStringTabular(
	ctx context.Context,
	filterValue string,
	filter query.StructuredFilter,

	preloads ...string,
) ([]byte, error) {
	return r.pagination.StructuredStringTabular(r.Client(ctx), filterValue, filter, r.tabular, r.preload(preloads...)...)
}
