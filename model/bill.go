package model

import (
	"time"

	"github.com/google/uuid"
)

type Bill struct {
	Base
	AccountID       uuid.UUID `json:"account_id"`
	Account         Account   `gorm:"foreignKey:AccountID;references:ID"`
	GrossAmount     int64     `json:"gross_amount"`
	PurchasedAmount int64     `json:"purchased_amount"`
	RemainingAmount int64     `json:"remaining_amount"`
	EnablePayments  string    `json:"enable_payments"`
	AdminFee        int64     `json:"admin_fee"`
	Deadline        time.Time `json:"deadline"`
	Description     *string   `json:"description"`
}
