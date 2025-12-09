package query

import (
	"go.uber.org/zap"
)

type Pagination[T any] struct {
	verbose           bool
	columnDefaultSort string
	columnDefaultID   string
	logger            *zap.Logger
}

type PaginationConfig struct {
	Verbose           bool   `json:"verbose"`
	ColumnDefaultSort string `json:"column_default_sort"`
	ColumnDefaultID   string `json:"column_default_id"`
}

func NewPagination[T any](config PaginationConfig) *Pagination[T] {
	var logger *zap.Logger
	if config.Verbose {
		var err error
		logger, err = zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
	} else {
		logger = zap.NewNop()
	}
	if config.ColumnDefaultID == "" {
		config.ColumnDefaultID = "id"
	}
	if config.ColumnDefaultSort == "" {
		config.ColumnDefaultSort = "created_at DESC"
	}
	return &Pagination[T]{
		verbose:           config.Verbose,
		columnDefaultSort: config.ColumnDefaultSort,
		columnDefaultID:   config.ColumnDefaultID,
		logger:            logger,
	}
}
