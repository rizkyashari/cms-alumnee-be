package model

type AcademicYear struct {
	Base
	Year     string
	IsActive bool `gorm:"default:false"`
}
