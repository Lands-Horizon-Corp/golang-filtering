package registry

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) GetByID(
	context context.Context,
	id any,
	preloads ...string,
) (*TData, error) {
	db := r.client.WithContext(context)
	entity, err := r.pagination.NormalGetByID(db, id, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity by ID: %w", err)
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) GetByIDRaw(
	context context.Context,
	id any,
	preloads ...string,
) (*TResponse, error) {
	data, err := r.GetByID(context, id, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw entity by ID: %w", err)
	}
	return r.ToModel(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) GetByIDIncludingDeleted(
	context context.Context,
	id any,
	preloads ...string,
) (*TData, error) {
	db := r.client.WithContext(context)
	entity, err := r.pagination.NormalGetByIDIncludingDeleted(db, id, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity by ID including deleted: %w", err)
	}
	return entity, nil
}

func (r *Registry[TData, TResponse, TRequest]) GetByIDIncludingDeletedRaw(
	context context.Context,
	id any,
	preloads ...string,
) (*TResponse, error) {
	data, err := r.GetByIDIncludingDeleted(context, id, preloads...)
	if err != nil {
		return nil, fmt.Errorf("failed to get raw entity by ID including deleted: %w", err)
	}
	return r.ToModel(data), nil
}

func (r *Registry[TData, TResponse, TRequest]) GetByIDLock(
	context context.Context,
	tx *gorm.DB,
	id any,
	preloads ...string,
) (*TData, error) {
	result, err := r.pagination.NormalGetByIDLock(context, tx, id, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity by ID with lock: %w", err)
	}
	return result, nil
}

func (r *Registry[TData, TResponse, TRequest]) GetByIDIncludingDeletedLock(
	context context.Context,
	id any,
	preloads ...string,
) (*TData, error) {
	tx := r.client.WithContext(context)
	result, err := r.pagination.NormalGetByIDIncludingDeletedLock(tx, id, r.preload(preloads...)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity by ID including deleted with lock: %w", err)
	}
	return result, nil
}
