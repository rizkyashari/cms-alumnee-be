package rs

import (
	"time"

	"github.com/google/uuid"
)

type MassCreateResponse struct {
	ID             uuid.UUID  `json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Destination    string     `json:"destination"`
	SuccessCount   int        `json:"success_count"`
	ErrorCount     int        `json:"error_count"`
	RequestFileUrl string     `json:"request_file_url"`
	ReportFileUrl  string     `json:"report_file_url"`
	ErrorMessages  []ErrorMsg `json:"error_messages"`
}

type ErrorMsg struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}
