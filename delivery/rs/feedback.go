package rs

import (
	"time"

	"github.com/google/uuid"
)

type FeedbackResponse struct {
	ID                uuid.UUID                `json:"id"`
	Title             string                   `json:"title"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
	AcademicYearID    string                   `json:"academic_year_id"`
	StudentID         string                   `json:"student_id"`
	TeacherID         string                   `json:"teacher_id"`
	IsTaughtByTeacher *bool                    `json:"is_taught_by_teacher"`
	FeedbackScores    *[]FeedbackScoreResponse `json:"feedback_score"`
}

type FeedbackQuestionResponse struct {
	ID       uuid.UUID `json:"id"`
	Question string    `json:"question"`
}

type FeedbackScoreResponse struct {
	ID                 string                    `json:"id"`
	FeedbackID         string                    `json:"feedback_id"`
	FeedbackQuestionID string                    `json:"feedback_question_id"`
	FeedbackQuestion   *FeedbackQuestionResponse `json:"feedback_question"`
	Value              int                       `json:"value"`
}
