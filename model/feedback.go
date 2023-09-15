package model

import "github.com/google/uuid"

type Feedback struct {
	Base
	AcademicYearID uuid.UUID
	StudentID      uuid.UUID
	TeacherID      uuid.UUID
	FeedbackScores []FeedbackScore
}

type FeedbackQuestion struct {
	Base
	Question string
}

type FeedbackScore struct {
	Base
	FeedbackID         uuid.UUID
	FeedbackQuestionID uuid.UUID
	FeedbackQuestion   FeedbackQuestion
	Value              int
}
