package model

import (
	"github.com/google/uuid"
)

type Subject struct {
	Base
	TeacherID         uuid.UUID
	Schedules         []Schedule
	SubjectComponents []SubjectComponent
}

type SubjectComponent struct {
	Base
	SubjectID       uuid.UUID
	ScoreComponents []ScoreComponent
}
