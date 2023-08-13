package model

import "github.com/google/uuid"

type Schedule struct {
	Base
	ClassroomID uuid.UUID
	SubjectID   uuid.UUID
	Classroom   Classroom `gorm:"foreignKey:ClassroomID;references:ID"`
}
