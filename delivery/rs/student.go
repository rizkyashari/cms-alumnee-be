package rs

import (
	"time"

	"github.com/google/uuid"
)

type StudentResponse struct {
	ID          uuid.UUID            `json:"id"`
	AccountID   uuid.UUID            `json:"account_id"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	ClassroomID *uuid.UUID           `json:"classroom_id"`
	StudentData *StudentDataResponse `json:"student_data,omitempty"`
}
