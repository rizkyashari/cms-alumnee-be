package rq

type PaginationParams[T interface{}] struct {
	Limit     int
	Page      int
	SortBy    string
	SortOrder string
	Data      T
}
