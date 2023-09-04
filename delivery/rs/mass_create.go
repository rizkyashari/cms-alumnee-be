package rs

import (
	"time"

	"github.com/fadhln/lms-be/model"
	"github.com/google/uuid"
)

type MassCreateResponse struct {
	ID             uuid.UUID        `json:"id"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	Status         int              `json:"status"`
	Destination    string           `json:"destination"`
	SuccessCount   int              `json:"success_count"`
	ErrorCount     int              `json:"error_count"`
	RequestFileUrl string           `json:"request_file_url"`
	ReportFileUrl  string           `json:"report_file_url"`
	ErrorMessages  []model.ErrorMsg `json:"error_messages"`
}
