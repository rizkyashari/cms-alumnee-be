package rq

type RewardPunishmentRequest struct {
	ID          *string `json:"id,omitempty"`
	StudentID   *string `json:"student_id,omitempty"`
	Type        *int    `json:"type,omitempty"`
	Point       *int    `json:"point,omitempty"`
	Description *string `json:"description,omitempty"`
}
