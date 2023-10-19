package rq

type TransactionDetailsRequest struct {
	OrderID     string `json:"order_id"`
	GrossAmount string `json:"gross_amount"`
}

type CustomerDetailsRequest struct {
	// FirstName string `json:"first_name"`
	Email string `json:"email"`
}

type TransactionRequest struct {
	BillID             *string                   `json:"bill_id"`
	TransactionDetails TransactionDetailsRequest `json:"transaction_details,omitempty"`
	CustomerDetails    CustomerDetailsRequest    `json:"customer_details,omitempty"`
}
