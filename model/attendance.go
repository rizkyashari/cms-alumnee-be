package model

import "github.com/google/uuid"

type Attendance struct {
	Base
	StudentID uuid.UUID
	EventID   uuid.UUID
	Status    int
	Remarks   *string
}
