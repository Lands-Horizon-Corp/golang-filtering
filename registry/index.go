package registry

import (
	"context"
	"fmt"

	"github.com/Lands-Horizon-Corp/golang-filtering/query"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Topics []string

type RegistryParams[TData any, TResponse any, TRequest any] struct {
	ColumnDefaultID   string
	ColumnDefaultSort string
	Database          *gorm.DB
	Dispatch          func(topics Topics, payload any) error
	Validator         *validator.Validate
	Created           func(*TData) Topics
	Updated           func(*TData) Topics
	Deleted           func(*TData) Topics
	Resource          func(*TData) *TResponse
	Tabular           func(data *TData) map[string]any
	Preloads          []string
}

type Registry[TData any, TResponse any, TRequest any] struct {
	columnDefaultID   string
	columnDefaultSort string
	database          *gorm.DB
	dispatch          func(topics Topics, payload any) error
	validator         *validator.Validate
	preloads          []string
	resource          func(*TData) *TResponse
	created           func(*TData) Topics
	updated           func(*TData) Topics
	deleted           func(*TData) Topics
	tabular           func(data *TData) map[string]any
	pagination        query.Pagination[TData]
}

func NewRegistry[TData any, TResponse any, TRequest any](
	params RegistryParams[TData, TResponse, TRequest],
) *Registry[TData, TResponse, TRequest] {
	if params.ColumnDefaultID == "" {
		params.ColumnDefaultID = "id"
	}
	if params.ColumnDefaultSort == "" {
		params.ColumnDefaultSort = "created_at DESC"
	}
	fmt.Println("Initializing Registry with ColumnDefaultID:", params)
	return &Registry[TData, TResponse, TRequest]{
		columnDefaultID:   params.ColumnDefaultID,
		columnDefaultSort: params.ColumnDefaultSort,
		database:          params.Database,
		dispatch:          params.Dispatch,
		preloads:          params.Preloads,
		resource:          params.Resource,
		created:           params.Created,
		updated:           params.Updated,
		deleted:           params.Deleted,
		tabular:           params.Tabular,
		validator:         params.Validator,
		pagination: *query.NewPagination[TData](query.PaginationConfig{
			Verbose:           true,
			ColumnDefaultSort: params.ColumnDefaultSort,
			ColumnDefaultID:   params.ColumnDefaultID,
		}),
	}
}

func (r *Registry[TData, TResponse, TRequest]) Client(context context.Context) *gorm.DB {
	if r.database == nil {
		return nil
	}
	return r.database.WithContext(context).Model(new(TData))
}

func (r *Registry[TData, TResponse, TRequest]) preload(preloads ...string) []string {
	if len(preloads) > 0 {
		if preloads[0] == "" {
			return []string{}
		}
		return preloads
	}
	return r.preloads
}
