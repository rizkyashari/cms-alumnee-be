package model

import "gorm.io/gorm"

type Classroom struct {
	gorm.Model
	AcademicYear      string
	HomeroomTeacherID uint
	SchoolID          uint
	Students          []Student
	Schedules         []Schedule
}
