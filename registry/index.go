package registry

import (
	"context"

	"gorm.io/gorm"
)

type Topics []string

type RegistryParams[TData any, TResponse any, TRequest any] struct {
	Database *gorm.DB
	Created  func(*TData) Topics
	Updated  func(*TData) Topics
	Deleted  func(*TData) Topics
	Resource func(*TData) *TResponse
	tabular  func(data *TData) map[string]any
	Preloads []string
}

type Registry[TData any, TResponse any, TRequest any] struct {
	database *gorm.DB
	preloads []string
	resource func(*TData) *TResponse
	created  func(*TData) Topics
	updated  func(*TData) Topics
	deleted  func(*TData) Topics
	tabular  func(data *TData) map[string]any
}

func NewRegistry[TData any, TResponse any, TRequest any](
	params RegistryParams[TData, TResponse, TRequest],
) *Registry[TData, TResponse, TRequest] {
	return &Registry[TData, TResponse, TRequest]{
		database: params.Database,
		preloads: params.Preloads,
		resource: params.Resource,
		created:  params.Created,
		updated:  params.Updated,
		deleted:  params.Deleted,
		tabular:  params.tabular,
	}
}

func (r *Registry[TData, TResponse, TRequest]) Client(context context.Context) *gorm.DB {
	if r.database == nil {
		return nil
	}
	return r.database.WithContext(context).Model(new(TData))
}
