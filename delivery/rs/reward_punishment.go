package rs

import (
	"time"

	"github.com/google/uuid"
)

type RewardPunishmentResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	StudentID   string    `json:"student_id"`
	Type        int       `json:"type"`
	Point       int       `json:"point"`
	Description string    `json:"description"`
}
