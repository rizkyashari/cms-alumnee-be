package rs

type TransactionResponse struct {
	Token             string `json:"token"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	PaymentType       string `json:"payment_type"`
	TransactionStatus string `json:"transaction_status"`
	TransactionTime   string `json:"transaction_time"`
	SettlementTime    string `json:"settlement_time"`
	ExpiryTime        string `json:"expiry_time"`
	BillID            string `json:"bill_id"`
}
