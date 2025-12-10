package registry

import (
	"context"
	"fmt"
)

func (r *Registry[TData, TResponse, TRequest]) List(
	context context.Context,
	preloads ...string,
) ([]*TData, error) {
	var entities []*TData
	db := r.client.WithContext(context)
	if preloads == nil {
		preloads = r.preloads
	}
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	if err := db.Order("updated_at DESC").Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}
	return entities, nil
}

func (r *Registry[TData, TResponse, TRequest]) ListRaw(
	context context.Context,
	preloads ...string,
) ([]*TResponse, error) {
	data, err := r.List(context, preloads...)
	if err != nil {
		return nil, fmt.Errorf("failed to list data: %w", err)
	}
	return r.ToModels(data), nil
}
