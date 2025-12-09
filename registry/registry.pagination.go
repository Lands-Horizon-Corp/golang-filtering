package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/labstack/echo/v4"
)

func (r *Registry[TData, TResponse, TRequest]) Pagination(
	context context.Context,
	ctx echo.Context,
	preloads ...string,
) (*query.PaginationResult[TResponse], error) {
	data, err := r.pagination.Pagination(r.Client(context), context, ctx, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return &query.PaginationResult[TResponse]{
		Data:      r.ToModels(data.Data),
		TotalSize: data.TotalSize,
		TotalPage: data.TotalPage,
		PageIndex: data.PageIndex,
		PageSize:  data.PageSize,
	}, nil
}

func (r *Registry[TData, TResponse, TRequest]) NormalPagination(
	context context.Context,
	ctx echo.Context,
	filter *TData,
	preloads ...string,
) (*query.PaginationResult[TResponse], error) {
	data, err := r.pagination.PaginationNormal(r.Client(context), context, ctx, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return &query.PaginationResult[TResponse]{
		Data:      r.ToModels(data.Data),
		TotalSize: data.TotalSize,
		TotalPage: data.TotalPage,
		PageIndex: data.PageIndex,
		PageSize:  data.PageSize,
	}, nil
}

func (r *Registry[TData, TResponse, TRequest]) ArrPagination(
	context context.Context,
	ctx echo.Context,

	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,

	preloads ...string,
) (*query.PaginationResult[TResponse], error) {
	data, err := r.pagination.PaginationArray(r.Client(context), context, ctx, filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return &query.PaginationResult[TResponse]{
		Data:      r.ToModels(data.Data),
		TotalSize: data.TotalSize,
		TotalPage: data.TotalPage,
		PageIndex: data.PageIndex,
		PageSize:  data.PageSize,
	}, nil
}

func (r *Registry[TData, TResponse, TRequest]) StructuredPagination(
	context context.Context,
	ctx echo.Context,

	filter query.StructuredFilter,

	preloads ...string,
) (*query.PaginationResult[TResponse], error) {
	data, err := r.pagination.PaginationStructured(r.Client(context), context, ctx, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, err
	}
	return &query.PaginationResult[TResponse]{
		Data:      r.ToModels(data.Data),
		TotalSize: data.TotalSize,
		TotalPage: data.TotalPage,
		PageIndex: data.PageIndex,
		PageSize:  data.PageSize,
	}, nil
}
