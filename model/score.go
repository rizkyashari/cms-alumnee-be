package model

import "github.com/google/uuid"

type Score struct {
	Base
	StudentID       uuid.UUID
	ScoreComponents []ScoreComponent
}

type ScoreComponent struct {
	Base
	ScoreID            uuid.UUID
	SubjectComponentID uuid.UUID
}
