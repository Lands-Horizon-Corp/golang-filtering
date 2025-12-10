package registry

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry[TData, TResponse, TRequest]) UpdateByID(
	context context.Context,
	id any,
	fields *TData,
	preloads ...string,
) error {
	if preloads == nil {
		preloads = r.preloads
	}
	if err := r.client.WithContext(context).Where(fmt.Sprintf("%s = ?", r.columnDefaultID), id).Updates(fields).Error; err != nil {
		return fmt.Errorf("failed to update fields for entity %v: %w", id, err)
	}
	reloadDb := r.client.WithContext(context).Where(fmt.Sprintf("%s = ?", r.columnDefaultID), id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return fmt.Errorf("failed to reload entity %v after field update: %w", id, err)
	}
	r.OnUpdate(context, fields)
	return nil
}

func (r *Registry[TData, TResponse, TRequest]) UpdateByIDWithTx(
	context context.Context,
	tx *gorm.DB,
	id any,
	fields *TData,
	preloads ...string,
) error {
	if preloads == nil {
		preloads = r.preloads
	}
	if err := tx.Where(fmt.Sprintf("%s = ?", r.columnDefaultID), id).Updates(fields).Error; err != nil {
		return fmt.Errorf("failed to update fields for entity %v in transaction: %w", id, err)
	}
	reloadDb := tx.Where(fmt.Sprintf("%s = ?", r.columnDefaultID), id)
	for _, preload := range preloads {
		reloadDb = reloadDb.Preload(preload)
	}
	if err := reloadDb.First(fields).Error; err != nil {
		return fmt.Errorf("failed to reload entity %v after field update in transaction: %w", id, err)
	}
	r.OnUpdate(context, fields)
	return nil
}
