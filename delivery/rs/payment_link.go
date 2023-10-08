package rs

import "time"

type CustomerDetailsResponse struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type PaymentLinkResponse struct {
	ID              uint                    `gorm:"primary_key"`
	Title           string                  `json:"title"`
	PaymentLinkURL  string                  `json:"payment_link_url"`
	CustomerDetails CustomerDetailsResponse `json:"customer_details"`
	Purchases       []PurchaseResponse      `json:"purchases"`
	Usage           uint                    `json:"usage"`
	UsageLimit      uint                    `json:"usage_limit"`
	DynamicAmount   DynamicAmountResponse   `json:"dynamic_amount"`
	CreatedAt       time.Time               `json:"createdAt"`
	UpdatedAt       time.Time               `json:"updatedAt"`
}

type PurchaseResponse struct {
	SnapToken     string    `json:"snap_token"`
	OrderID       string    `json:"order_id"`
	PaymentStatus string    `json:"payment_status"`
	PaymentMethod string    `json:"payment_method"`
	AmountValue   uint      `json:"amount_value"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	PaymentLinkID uint
}

type DynamicAmountResponse struct {
	PaymentLinkID uint `json:"payment_link_id"`
	PresetAmount  uint `json:"preset_amount"`
}
