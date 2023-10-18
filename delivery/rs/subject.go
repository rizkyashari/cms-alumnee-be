package rs

import (
	"time"

	"github.com/google/uuid"
)

type SubjectResponse struct {
	ID                uuid.UUID                  `json:"id"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
	Name              string                     `json:"name"`
	TeacherID         uuid.UUID                  `json:"teacher_id"`
	Teacher           AccountResponse            `json:"teacher"`
	SubjectComponents []SubjectComponentResponse `json:"subject_component"`
}

type SubjectComponentResponse struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Name       string    `json:"name"`
	SubjectID  uuid.UUID `json:"subject_id"`
	Percentage int       `json:"percentage"`
}
