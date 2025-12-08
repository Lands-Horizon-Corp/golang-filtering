package registry

import (
	"context"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Topics []string
type RegistryEvent interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	Publish(topic string, payload any) error
	Dispatch(topics Topics, payload any) error
}
type RegistryParams[TData any, TResponse any, TRequest any] struct {
	Database  *gorm.DB
	Event     RegistryEvent
	Validator *validator.Validate
	Created   func(*TData) Topics
	Updated   func(*TData) Topics
	Deleted   func(*TData) Topics
	Resource  func(*TData) *TResponse
	tabular   func(data *TData) map[string]any
	Preloads  []string
}

type Registry[TData any, TResponse any, TRequest any] struct {
	database   *gorm.DB
	event      RegistryEvent
	validator  *validator.Validate
	preloads   []string
	resource   func(*TData) *TResponse
	created    func(*TData) Topics
	updated    func(*TData) Topics
	deleted    func(*TData) Topics
	tabular    func(data *TData) map[string]any
	pagination query.Pagination[TData]
}

func NewRegistry[TData any, TResponse any, TRequest any](
	params RegistryParams[TData, TResponse, TRequest],
) *Registry[TData, TResponse, TRequest] {
	return &Registry[TData, TResponse, TRequest]{
		database:   params.Database,
		event:      params.Event,
		preloads:   params.Preloads,
		resource:   params.Resource,
		created:    params.Created,
		updated:    params.Updated,
		deleted:    params.Deleted,
		tabular:    params.tabular,
		validator:  params.Validator,
		pagination: query.Pagination[TData]{},
	}
}

func (r *Registry[TData, TResponse, TRequest]) Client(context context.Context) *gorm.DB {
	if r.database == nil {
		return nil
	}
	return r.database.WithContext(context).Model(new(TData))
}
