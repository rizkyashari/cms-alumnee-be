package model

import (
	"database/sql"
)

const (
	ADMIN   int = iota // Admin
	STUDENT            // Student
	TEACHER            // Tch
)

func IsValidAccountType(number int) bool {
	return number >= 0 && number <= int(TEACHER)
}

type Account struct {
	Base
	Email       string `gorm:"uniqueIndex"`
	Username    *string
	Password    string `json:"-"`
	AccountType int
	ActivatedAt sql.NullTime
	Admin       *Admin
	Student     *Student
	Teacher     *Teacher
}
