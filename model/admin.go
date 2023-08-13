package model

import (
	"github.com/google/uuid"
)

type Admin struct {
	Base
	AccountID uuid.UUID `gorm:"uniqueIndex"`
}
