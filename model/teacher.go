package model

import "github.com/google/uuid"

type Teacher struct {
	Base
	AccountID   uuid.UUID `gorm:"uniqueIndex"`
	SchoolID    uuid.UUID
	Subjects    []Subject
	TeacherData TeacherData
}
