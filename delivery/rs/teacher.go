package rs

import (
	"time"

	"github.com/google/uuid"
)

type TeacherResponse struct {
	ID          uuid.UUID            `json:"id"`
	AccountID   uuid.UUID            `json:"account_id"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	SchoolID    uuid.UUID            `json:"school_id"`
	TeacherData *TeacherDataResponse `json:"teacher_data,omitempty"`
}

type ClassroomSubject struct {
	ID        uuid.UUID         `json:"id"`
	AccountID uuid.UUID         `json:"account_id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Classroom ClassroomResponse `json:"classroom"`
	Subject   SubjectResponse   `json:"subject"`
}
