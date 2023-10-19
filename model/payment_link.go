package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type CustomerDetails struct {
	Base
	FullName string
	Email    string
}

func (cd CustomerDetails) Value() (driver.Value, error) {
	// Implement how CustomerDetails is stored in the database
	return fmt.Sprintf("%s,%s", cd.FullName, cd.Email), nil
}

func (cd *CustomerDetails) Scan(value interface{}) error {
	// Implement how CustomerDetails is retrieved from the database
	if str, ok := value.(string); ok {
		parts := strings.Split(str, ",")
		if len(parts) == 2 {
			cd.FullName = parts[0]
			cd.Email = parts[1]
		}
	}
	return nil
}

type PaymentLink struct {
	Base
	Title           string
	PaymentLinkURL  string
	CustomerDetails CustomerDetails `gorm:"foreignKey:CustomerDetailsID;references:ID"`
	Purchases       []Purchase
	DynamicAmount   DynamicAmount
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Purchase struct {
	Base
	SnapToken     string
	OrderID       string
	PaymentStatus string
	PaymentMethod string
	AmountValue   uint
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	PaymentLinkID uint
}

type DynamicAmount struct {
	Base
	PaymentLinkID uint
	PresetAmount  uint `json:"preset_amount"`
}
