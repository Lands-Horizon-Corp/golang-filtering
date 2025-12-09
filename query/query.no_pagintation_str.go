package query

import (
	"fmt"

	"gorm.io/gorm"
)

func (f *Pagination[T]) NoPaginationStr(
	db *gorm.DB,
	filterValue string,
	preloads ...string,
) ([]*T, error) {
	filterRoot, _, _, err := strParseQuery(filterValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query string: %w", err)
	}
	return f.StructuredFind(db, filterRoot, preloads...)
}

func (f *Pagination[T]) NoPaginationStructuredStr(
	db *gorm.DB,
	filterValue string,
	filter StructuredFilter,
	preloads ...string,
) ([]*T, error) {
	filterRoot, _, _, err := strParseQuery(filterValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query string: %w", err)
	}
	filterRoot.FieldFilters = append(filterRoot.FieldFilters, filter.FieldFilters...)
	filterRoot.Logic = LogicAnd
	if len(filterRoot.SortFields) == 0 && len(filter.SortFields) > 0 {
		filterRoot.SortFields = filter.SortFields
	}
	filterRoot.Preload = append(filterRoot.Preload, filter.Preload...)
	return f.StructuredFind(db, filterRoot, preloads...)
}

func (f *Pagination[T]) NoPaginationArrayStr(
	db *gorm.DB,
	filterValue string,
	filters []ArrFilterSQL,
	sorts []ArrFilterSortSQL,
	preloads ...string,
) ([]*T, error) {
	filterRoot, _, _, err := strParseQuery(filterValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query string: %w", err)
	}
	for _, f := range filters {
		filterRoot.FieldFilters = append(filterRoot.FieldFilters, FieldFilter{
			Field:    f.Field,
			Value:    f.Value,
			Mode:     f.Op,
			DataType: DataTypeText,
		})
	}
	filterRoot.Logic = LogicAnd
	if len(filterRoot.SortFields) == 0 && len(sorts) > 0 {
		for _, s := range sorts {
			filterRoot.SortFields = append(filterRoot.SortFields, SortField(s))
		}
	}
	filterRoot.Preload = append(filterRoot.Preload, preloads...)
	return f.StructuredFind(db, filterRoot, preloads...)
}

func (f *Pagination[T]) NoPaginationNormalStr(
	db *gorm.DB,
	filterValue string,
	filter *T,
	preloads ...string,
) ([]*T, error) {
	filterRoot, _, _, err := strParseQuery(filterValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query string: %w", err)
	}
	if filter != nil {
		db = db.Where(filter)
	}
	return f.StructuredFind(db, filterRoot, preloads...)
}
