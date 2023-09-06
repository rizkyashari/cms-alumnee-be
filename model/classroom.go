package model

import "github.com/google/uuid"

type Classroom struct {
	Base
	Name           string
	Code           string `gorm:"uniqueIndex"`
	TeacherID      uuid.UUID
	SchoolID       uuid.UUID
	AcademicYearID uuid.UUID
	AcademicYear   AcademicYear `gorm:"foreignKey:AcademicYearID;references:ID"`
	Teacher        Teacher      `gorm:"foreignKey:TeacherID;references:ID"`
	Schedules      []Schedule
}

func GetClassroomHeader() []string {
	return []string{
		"Name",
		"TeacherID",
		"SchoolID",
		"AcademicYearID",
	}
}

func GetClassroomRow(classroom *Classroom) []string {
	return []string{
		classroom.Name,
		classroom.TeacherID.String(),
		classroom.SchoolID.String(),
		classroom.AcademicYearID.String(),
	}
}
