package model

import "github.com/google/uuid"

type Attendance struct {
	Base
	StudentID   uuid.UUID
	ScheduleID  uuid.UUID
	Status      int
	Description *string
}
