package model

import "gorm.io/gorm"

type Score struct {
	gorm.Model
	StudentID       uint
	ScoreComponents []ScoreComponent
}

type ScoreComponent struct {
	gorm.Model
	ScoreID            uint
	SubjectComponentID uint
}
