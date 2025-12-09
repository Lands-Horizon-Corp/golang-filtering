package registry

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
)

func (r *Registry[TData, TResponse, TRequest]) NoPaginationStr(
	ctx context.Context,
	filterValue string,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationStr(r.Client(ctx), filterValue, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get no pagination data: %w", err)
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationNormalStr(
	ctx context.Context,
	filterValue string,
	filter *TData,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationNormalStr(r.Client(ctx), filterValue, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get no pagination normal data: %w", err)
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationArrayStr(
	ctx context.Context,
	filterValue string,
	filters []query.ArrFilterSQL,
	sorts []query.ArrFilterSortSQL,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationArrayStr(r.Client(ctx), filterValue, filters, sorts, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get no pagination array data: %w", err)
	}
	return r.ToModels(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) NoPaginationStructuredStr(
	ctx context.Context,
	filterValue string,
	filter query.StructuredFilter,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.pagination.NoPaginationStructuredStr(r.Client(ctx), filterValue, filter, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get no pagination structured data: %w", err)
	}
	return r.ToModels(data), nil
}
