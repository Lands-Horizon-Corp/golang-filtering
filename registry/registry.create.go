package registry

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Create(
	context context.Context,
	data *TData,
) error {
	db := r.client.WithContext(context)
	if err := db.Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}
	r.OnCreate(context, data)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) CreateWithTx(
	context context.Context,
	tx *gorm.DB,
	data *TData,
) error {
	if err := tx.Create(data).WithContext(context).Error; err != nil {
		return fmt.Errorf("failed to create entity with transaction: %w", err)
	}
	r.OnCreate(context, data)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) CreateMany(
	context context.Context,
	data []*TData,
) error {
	db := r.client.WithContext(context)
	if err := db.Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entities: %w", err)
	}
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) CreateManyWithTx(
	context context.Context,
	tx *gorm.DB,
	data []*TData,
) error {
	if err := tx.Create(data).WithContext(context).Error; err != nil {
		return fmt.Errorf("failed to create entities with transaction: %w", err)
	}
	for _, entity := range data {
		r.OnCreate(context, entity)
	}
	return nil
}
