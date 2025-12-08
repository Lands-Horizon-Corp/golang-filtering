package query

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
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

// Helper to log SQL query
func (p *Pagination[T]) log(db *gorm.DB, msg string) {
	if p.Verbose && db != nil {
		db = db.Debug() // GORM will print SQL to stdout
		if db.Statement != nil {
			fmt.Println("=============================")

			fmt.Println(db.Statement.SQL)
			fmt.Println("=============================")
			p.logger.Info(msg,
				zap.String("sql", db.Statement.SQL.String()),
				zap.Any("vars", db.Statement.Vars),
			)
		}
	}
}
