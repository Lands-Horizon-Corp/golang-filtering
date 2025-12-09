package registry

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Create(
	ctx context.Context,
	data *TData,
) error {
	db := r.Client(ctx)
	if err := db.Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entity: %w", err)
	}
	r.OnCreate(ctx, data)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) CreateWithTx(
	ctx context.Context,
	tx *gorm.DB,
	data *TData,
) error {
	if err := tx.Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entity with transaction: %w", err)
	}
	r.OnCreate(ctx, data)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) CreateMany(
	ctx context.Context,
	data []*TData,
) error {
	db := r.Client(ctx)
	if err := db.Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entities: %w", err)
	}
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) CreateManyWithTx(
	ctx context.Context,
	tx *gorm.DB,
	data []*TData,
) error {
	if err := tx.Create(data).Error; err != nil {
		return fmt.Errorf("failed to create entities with transaction: %w", err)
	}
	for _, entity := range data {
		r.OnCreate(ctx, entity)
	}
	return nil
}
