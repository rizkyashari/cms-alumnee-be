package rs

import (
	"time"

	"github.com/google/uuid"
)

type ScoreResponse struct {
	ID                 uuid.UUID `json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	StudentID          uuid.UUID `json:"student_id"`
	SubjectComponentID uuid.UUID `json:"subject_component_id"`
	Value              float64   `json:"value"`
	Percentage         int       `json:"percentage"`
	Description        string    `json:"description"`
}

type StudentScoreResponse struct {
	StudentID          uuid.UUID       `json:"student_id"`
	SubjectComponentID uuid.UUID       `json:"subject_component_id"`
	Total              float64         `json:"total"`
	Scores             []ScoreResponse `json:"scores"`
}

type TotalScoreResponse struct {
	StudentID uuid.UUID `json:"student_id"`
	SubjectID uuid.UUID `json:"subject_id"`
	Total     *float64  `json:"total,omitempty"`
}
