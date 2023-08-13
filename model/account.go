package model

import (
	"database/sql"
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
	Base
	Email       string `gorm:"uniqueIndex"`
	Username    *string
	Password    string
	AccountType int
	ActivatedAt sql.NullTime
	Admin       *Admin
	Student     *Student
	Teacher     *Teacher
}
