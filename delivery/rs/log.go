package rs

import (
	"time"

	"github.com/google/uuid"
)

type LogDataResponse struct {
	ID                        uuid.UUID `json:"id"`
	CreatedAt                 time.Time `json:"created_at"`
	UserRole                  int       `json:"user_role"`
	UserName                  string    `json:"user_name"`
	UserEmail                 string    `json:"user_email"`
	RequestMethod             string    `json:"request_method"`
	RequestURL                string    `json:"request_url"`
	RequestBody               string    `json:"request_body"`
	ResponseData              string    `json:"response_data"`
	ResponseStatus            int       `json:"response_status"`
	ResponseStatusDescription string    `json:"response_status_description"`
}
