package model

import (
	"database/sql"

	"gorm.io/gorm"
)

const (
	Adm int = iota // Admin
	Stu            // Student
	Tch            // Tch
)

func IsValidAccountType(number int) bool {
	return number >= 0 && number <= int(Tch)
}

type Account struct {
	gorm.Model
	Email       string `gorm:"uniqueIndex"`
	Username    *string
	Password    string
	AccountType int
	ActivatedAt sql.NullTime
	Admin       *Admin
	Student     *Student
	Teacher     *Teacher
}
