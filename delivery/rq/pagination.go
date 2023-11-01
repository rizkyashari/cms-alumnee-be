package rq

import (
	"time"

	"github.com/google/uuid"
)

type PaginationParams[T interface{}] struct {
	Limit     int
	Page      int
	SortBy    string
	SortOrder string
	Data      T
}

type EventParamsData struct {
	ClassroomID *uuid.UUID
	TeacherID   *uuid.UUID
	StudentID   *uuid.UUID
}

type EventParams struct {
	Begin time.Time
	End   time.Time
	Data  *EventParamsData
}
