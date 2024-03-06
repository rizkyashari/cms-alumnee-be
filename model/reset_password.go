package model

import (
	"time"

	"github.com/google/uuid"
)

type ResetPassword struct {
	AccountID            uuid.UUID `json:"account_id"`
	Account              Account   `gorm:"foreignKey:AccountID;references:ID"`
	ResetPasswordToken   string
	ResetPasswordExpires time.Time
}
