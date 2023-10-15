package rs

import (
	"time"

	"github.com/google/uuid"
)

type BillResponse struct {
	ID              uuid.UUID `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	AccountID       *string   `json:"account_id"`
	GrossAmount     int64     `json:"gross_amount"`
	PurchasedAmount int64     `json:"purchased_amount"`
	RemainingAmount int64     `json:"remaining_amount"`
	AdminFee        int64     `json:"admin_fee"`
	Deadline        time.Time `json:"deadline"`
	Description     string    `json:"description,omitempty"`
}
