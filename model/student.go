package model

import "gorm.io/gorm"

type Student struct {
	gorm.Model
	AccountID   uint `gorm:"uniqueIndex"`
	ClassroomID uint
	Scores      []Score
}
