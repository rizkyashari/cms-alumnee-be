package model

import (
	"database/sql"

	"gorm.io/gorm"
)

type AccountType int

const (
	Adm AccountType = iota // Admin
	Stu                    // Student
	Tch                    // Tch
)

type Account struct {
	gorm.Model
	Username    string
	Email       *string
	Password    string
	AccountType int
	ActivatedAt sql.NullTime
	Admin       *Admin
	Student     *Student
	Teacher     *Teacher
}
