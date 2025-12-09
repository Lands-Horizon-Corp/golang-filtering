package registry

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Delete(
	ctx context.Context,
	id any,
) error {
	return r.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.DeleteWithTx(ctx, tx, id)
	})
}

func (r *Registry[TData, TResponse, TRequest]) DeleteWithTx(
	ctx context.Context,
	tx *gorm.DB,
	id any,
) error {
	var entity TData
	if err := tx.First(&entity, id).Error; err != nil {
		return fmt.Errorf("failed to find entity for delete: %w", err)
	}
	if err := tx.Delete(&entity).Error; err != nil {
		return fmt.Errorf("failed to delete entity with transaction: %w", err)
	}
	r.OnDelete(ctx, &entity)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) BulkDelete(
	ctx context.Context,
	ids []any,
) error {
	if len(ids) == 0 {
		return nil
	}
	return r.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.BulkDeleteWithTx(ctx, tx, ids)
	})
}

func (r *Registry[TData, TResponse, TRequest]) BulkDeleteWithTx(
	ctx context.Context,
	tx *gorm.DB,
	ids []any,
) error {
	if len(ids) == 0 {
		return nil
	}
	var entities []TData
	if err := tx.Find(&entities, ids).Error; err != nil {
		return fmt.Errorf("failed to find entities for bulk delete: %w", err)
	}
	if len(entities) != len(ids) {
		return fmt.Errorf("some entities not found for bulk delete")
	}
	if err := tx.Delete(&entities).Error; err != nil {
		return fmt.Errorf("failed to bulk delete entities: %w", err)
	}
	for _, data := range entities {
		r.OnDelete(ctx, &data)
	}
	return nil
}
