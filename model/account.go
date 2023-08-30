package model

import (
	"database/sql"
)

const (
	ADMIN int = iota
	STUDENT
	TEACHER
)

func IsValidAccountType(number int) bool {
	return number >= 0 && number <= int(TEACHER)
}

type Account struct {
	Base
	Email       string `gorm:"uniqueIndex"`
	Name        *string
	Avatar      *string
	Password    string `json:"-"`
	AccountType int
	ActivatedAt sql.NullTime
	Admin       *Admin
	Student     *Student
	Teacher     *Teacher
}
