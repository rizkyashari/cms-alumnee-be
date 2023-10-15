package model

import (
	"time"

	"github.com/google/uuid"
)

type Bill struct {
	Base
	AccountID       uuid.UUID
	Account         Account `gorm:"foreignKey:AccountID;references:ID"`
	GrossAmount     int64
	PurchasedAmount int64
	RemainingAmount int64
	EnablePayments  string `json:"enable_payments"`
	AdminFee        int64
	Deadline        time.Time
	Description     *string
}
