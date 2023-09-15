package model

import "github.com/google/uuid"

type Score struct {
	Base
	StudentID          uuid.UUID
	SubjectComponentID uuid.UUID
	Value              float64
	SubjectComponent   SubjectComponent
}
