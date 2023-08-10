package model

import "gorm.io/gorm"

type Teacher struct {
	gorm.Model
	AccountID uint `gorm:"uniqueIndex"`
	Name      string
	NIP       string
	Subjects  []Subject
	Homeroom  Classroom `gorm:"foreignKey:HomeroomTeacherID"`
}
