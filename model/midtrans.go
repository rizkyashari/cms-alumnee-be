package model

import "github.com/google/uuid"

type TransactionDetail struct {
	Base
	OrderID     string
	GrossAmount string
}

type CustomerDetail struct {
	Base
	FirstName string
	Email     string
}

type Transaction struct {
	Base
	OrderID           string
	GrossAmount       int64
	BillFee           int64
	AdminFee          int64
	Email             string
	TransactionID     string
	PaymentType       string
	TransactionStatus string
	TransactionTime   string
	SettlementTime    string
	ExpiryTime        string
	Token             string
	BillID            uuid.UUID
	Bill              Bill `gorm:"foreignKey:BillID;references:ID"`
}

type MidtransCredentials struct {
	Base
	ServerKey      string    `json:"server_key"`
	ClientKey      string    `json:"client_key"`
	Environment    int       `json:"environment"`
	TransactionAPI string    `json:"transaction_api"`
	SnapJSUrl      string    `json:"snapjs_url"`
	SchoolID       uuid.UUID `json:"school_id"`
	School         School    `gorm:"foreignKey:SchoolID;references:ID"`
}
