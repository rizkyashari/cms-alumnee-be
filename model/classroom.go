package model

import "github.com/google/uuid"

type Classroom struct {
	Base
	TeacherID      uuid.UUID
	SchoolID       uuid.UUID
	AcademicYearID uuid.UUID
	Teacher        Teacher `gorm:"foreignKey:TeacherID;references:ID"`
}
