package rs

import (
	"time"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/google/uuid"
)

type EventDraftResponse struct {
	ID             uuid.UUID                     `json:"id"`
	CreatedAt      time.Time                     `json:"created_at"`
	UpdatedAt      time.Time                     `json:"updated_at"`
	SchoolID       uuid.UUID                     `json:"school_id"`
	AcademicYearID uuid.UUID                     `json:"academic_year_id"`
	StudyHourUnit  int                           `json:"study_hour_unit"`
	Draft          []rq.CreateSingleEventRequest `json:"draft"`
}
