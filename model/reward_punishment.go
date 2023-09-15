package model

import "github.com/google/uuid"

type RewardPunishment struct {
	Base
	StudentID   uuid.UUID
	Type        int
	Point       int
	Description *string
}
