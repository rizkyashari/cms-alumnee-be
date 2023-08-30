package rs

import (
	"time"

	"github.com/google/uuid"
)

type AcademicYearResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Year      string    `json:"year"`
	IsActive  bool      `json:"is_active"`
}
