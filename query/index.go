package query

type Pagination[T any] struct{}

func NewPagination[T any]() *Pagination[T] {
	return &Pagination[T]{}
}
