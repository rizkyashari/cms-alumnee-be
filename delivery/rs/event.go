package rs

import (
	"time"

	"github.com/google/uuid"
)

type EventResponse struct {
	ID          uuid.UUID         `json:"id"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	BeginDate   time.Time         `json:"begin_date"`
	EndDate     time.Time         `json:"end_date"`
	Type        string            `json:"type"`
	Classroom   ClassroomResponse `json:"classroom"`
	Title       *string           `json:"title,omitempty"`
	Description *string           `json:"description,omitempty"`
	Subject     *SubjectResponse  `json:"subject,omitempty"`
}

type ManyEventReponse struct {
	Begin  time.Time       `json:"begin"`
	End    time.Time       `json:"end"`
	Events []EventResponse `json:"events"`
}
