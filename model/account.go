package model

import (
	"database/sql"

	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Username    string
	Email       *string
	Password    string
	ActivatedAt sql.NullTime
	Admin       *Admin
	Student     *Student
	Teacher     *Teacher
}
