package model

import (
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	Base
	Description *string
	Color       string `gorm:"default:'#049AE3'"`
	Start       time.Time
	End         time.Time
	ClassroomID uuid.UUID
	SubjectID   uuid.UUID
}
