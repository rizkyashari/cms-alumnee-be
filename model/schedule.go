package model

import "gorm.io/gorm"

type Schedule struct {
	gorm.Model
	SubjectID   uint
	ClassroomID uint
}
