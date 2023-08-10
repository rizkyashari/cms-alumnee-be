package model

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	AccountID uint `gorm:"uniqueIndex"`
}
