package model

import "github.com/google/uuid"

type Teacher struct {
	Base
	AccountID uuid.UUID `gorm:"uniqueIndex"`
	Name      *string
	NIP       string
	Subjects  []Subject
}
