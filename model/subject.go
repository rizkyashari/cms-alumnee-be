package model

import (
	"github.com/google/uuid"
)

type Subject struct {
	Base
	Name              string
	TeacherID         uuid.UUID
	Schedules         []Schedule
	SubjectComponents []SubjectComponent
}

type SubjectComponent struct {
	Base
	Name            string
	SubjectID       uuid.UUID
	ScoreComponents []ScoreComponent
}
