package rq

import (
	"time"

	"github.com/fadhln/lms-be/model"
)

type PaginationParams[T interface{}] struct {
	Limit     int
	Page      int
	SortBy    string
	SortOrder string
	Data      T
}

type ScheduleParams struct {
	Start time.Time
	End   time.Time
	Data  *model.Schedule
}
