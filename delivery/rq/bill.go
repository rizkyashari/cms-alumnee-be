package rq

type BillRequest struct {
	ID             *string  `json:"id,omitempty"`
	SchoolID       *string  `json:"school_id"`
	AccountID      *string  `json:"account_id"`
	GrossAmount    int64    `json:"gross_amount"`
	Environment    int64    `json:"environment"`
	EnablePayments []string `json:"enable_payments"`
	AdminFee       int64    `json:"admin_fee"`
	Deadline       string   `json:"deadline"`
	Description    string   `json:"description,omitempty"`
}
