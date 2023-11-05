package model

import "github.com/google/uuid"

type Student struct {
	Base
	AccountID         uuid.UUID `gorm:"uniqueIndex"`
	StudentData       StudentData
	ClassroomID       *uuid.UUID
	Classroom         Classroom `gorm:"foreignKey:ClassroomID;references:ID"`
	Scores            []Score
	RewardPunishments []RewardPunishment
	Attendances       []Attendance
}
