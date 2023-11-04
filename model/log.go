package model

import (
	"time"

	"github.com/google/uuid"
)

type LogData struct {
	ID                        uuid.UUID `gorm:"primarykey"`
	CreatedAt                 time.Time
	UserRole                  int
	UserName                  string
	UserEmail                 string
	RequestMethod             string
	RequestURL                string
	RequestBody               string
	ResponseData              string
	ResponseStatus            int
	ResponseStatusDescription string
}
