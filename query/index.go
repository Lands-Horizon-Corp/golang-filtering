package query

import (
	"go.uber.org/zap"
)

type Pagination[T any] struct {
	Verbose bool `json:"verbose"`
	logger  *zap.Logger
}

func NewPagination[T any](verbose bool) *Pagination[T] {
	var logger *zap.Logger
	if verbose {
		var err error
		logger, err = zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
	} else {
		logger = zap.NewNop()
	}
	return &Pagination[T]{Verbose: verbose, logger: logger}
}
