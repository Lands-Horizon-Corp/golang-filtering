package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/labstack/echo/v4"
)

func (r *Registry[TData, TResponse, TRequest]) NoPagination(
	ctx context.Context,
	echoCtx echo.Context,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPagination(r.Client(ctx), echoCtx, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}

	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationNormal(
	ctx context.Context,
	echoCtx echo.Context,
	filter *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationNormal(r.Client(ctx), ctx, echoCtx, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}

	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationArray(
	ctx context.Context,
	echoCtx echo.Context,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationArray(r.Client(ctx), ctx, echoCtx, filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}

	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationStructured(
	ctx context.Context,
	echoCtx echo.Context,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationStructured(r.Client(ctx), ctx, echoCtx, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}

	return r.ToModels(data), nil
}
