package model

type Account struct {
	Base
	Email       string `gorm:"uniqueIndex"`
	Name        *string
	Avatar      *string
	Password    string `json:"-"`
	AccountType int
	Admin       *Admin
	Student     *Student
	Teacher     *Teacher
}
