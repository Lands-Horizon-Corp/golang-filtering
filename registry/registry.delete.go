package registry

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) Delete(
	context context.Context,
	id any,
) error {
	return r.client.WithContext(context).Transaction(func(tx *gorm.DB) error {
		return r.DeleteWithTx(context, tx, id)
	})
}

func (r *Registry[TData, TResponse, TRequest]) DeleteWithTx(
	context context.Context,
	tx *gorm.DB,
	id any,
) error {
	var entity TData
	if err := tx.Where(fmt.Sprintf("%s = ?", r.columnDefaultID), id).First(&entity).WithContext(context).Error; err != nil {
		return fmt.Errorf("failed to find entity for delete: %w", err)
	}

	if err := tx.Delete(&entity).Error; err != nil {
		return fmt.Errorf("failed to delete entity with transaction: %w", err)
	}

	r.OnDelete(context, &entity)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) BulkDelete(
	context context.Context,
	ids []any,
) error {
	if len(ids) == 0 {
		return nil
	}
	return r.client.WithContext(context).Transaction(func(tx *gorm.DB) error {
		return r.BulkDeleteWithTx(context, tx, ids)
	})
}

func (r *Registry[TData, TResponse, TRequest]) BulkDeleteWithTx(
	context context.Context,
	tx *gorm.DB,
	ids []any,
) error {
	if len(ids) == 0 {
		return nil
	}
	var entities []TData
	if err := tx.Where(fmt.Sprintf("%s IN ?", r.columnDefaultID), ids).Find(&entities).Error; err != nil {
		return fmt.Errorf("failed to find entities for bulk delete: %w", err)
	}

	if len(entities) != len(ids) {
		return fmt.Errorf("some entities not found for bulk delete")
	}
	if err := tx.Where(fmt.Sprintf("%s IN ?", r.columnDefaultID), ids).Delete(&entities).Error; err != nil {
		return fmt.Errorf("failed to bulk delete entities: %w", err)
	}
	for _, data := range entities {
		r.OnDelete(context, &data)
	}

	return nil
}
