package model

import "github.com/google/uuid"

type EventDraft struct {
	Base
	SchoolID       uuid.UUID
	AcademicYearID uuid.UUID
	StudyHourUnit  int
	Draft          string
}
