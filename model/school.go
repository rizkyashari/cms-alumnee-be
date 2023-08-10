package model

import "gorm.io/gorm"

type School struct {
	gorm.Model
	Name       string
	Classrooms []Classroom
}
