package rq

import (
	"github.com/google/uuid"
)

type GetAllAttendanceParams struct {
	StudentID   *uuid.UUID
	EventID     *uuid.UUID
	SubjectID   *uuid.UUID
	ClassroomID *uuid.UUID
}

type CreateAttendanceStatusRequest struct {
	StudentID uuid.UUID `json:"student_id"`
	EventID   uuid.UUID `json:"event_id"`
	Status    int       `json:"status"`
	Remarks   *string   `json:"remarks,omitempty"`
}

type CreateAttendanceRequest struct {
	EventID  string                          `json:"event_id"`
	Statuses []CreateAttendanceStatusRequest `json:"statuses"`
}

type EditAttendanceRequest struct {
	ID      string  `json:"id"`
	Status  int     `json:"status"`
	Remarks *string `json:"remarks,omitempty"`
}
