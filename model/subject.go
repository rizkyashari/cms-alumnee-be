package model

import "gorm.io/gorm"

type Subject struct {
	gorm.Model
	TeacherID         uint
	Schedules         []Schedule
	SubjectComponents []SubjectComponent
}

type SubjectComponent struct {
	gorm.Model
	SubjectID       uint
	ScoreComponents []ScoreComponent
}
