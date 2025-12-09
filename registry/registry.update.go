package registry

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) UpdateByID(
	ctx context.Context,
	id any,
	fields *TData,
	preloads ...string,
) error {
	if preloads == nil {
		preloads = r.preloads
	}
	if err := r.Client(ctx).Model(new(TData)).Where("id = ?", id).Updates(fields).Error; err != nil {
		return fmt.Errorf("failed to update fields for entity %v: %w", id, err)
	}
	reloadDb := r.Client(ctx).Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return fmt.Errorf("failed to reload entity %v after field update: %w", id, err)
	}
	r.OnUpdate(ctx, fields)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) UpdateByIDWithTx(
	ctx context.Context,
	tx *gorm.DB,
	id any,
	fields *TData,
	preloads ...string,
) error {
	if preloads == nil {
		preloads = r.preloads
	}
	if err := tx.Model(new(TData)).Where("id = ?", id).Updates(fields).Error; err != nil {
		return fmt.Errorf("failed to update fields for entity %v in transaction: %w", id, err)
	}
	reloadDb := tx.Where("id = ?", id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return fmt.Errorf("failed to reload entity %v after field update in transaction: %w", id, err)
	}
	r.OnUpdate(ctx, fields)
	return nil
}
