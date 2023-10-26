package model

import "github.com/google/uuid"

type Feedback struct {
	Base
	Title          string
	AcademicYearID uuid.UUID
	StudentID      uuid.UUID
	TeacherID      uuid.UUID
	Teacher        Teacher `gorm:"foreignKey:TeacherID;references:ID"`
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
