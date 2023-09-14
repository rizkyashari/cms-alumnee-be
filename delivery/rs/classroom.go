package rs

import (
	"time"

	"github.com/google/uuid"
)

type ClassroomResponse struct {
	ID           uuid.UUID            `json:"id"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	Name         string               `json:"name"`
	Code         string               `json:"code"`
	SchoolID     string               `json:"school_id"`
	AcademicYear AcademicYearResponse `json:"academic_year"`
	Teacher      TeacherResponse      `json:"teacher"`
}
