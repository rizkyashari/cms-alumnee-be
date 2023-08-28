package rs

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type AccountResponse struct {
	ID          uuid.UUID        `json:"id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Email       string           `json:"email"`
	Username    string           `json:"username"`
	Name        string           `json:"name"`
	Avatar      string           `json:"avatar"`
	AccountType int              `json:"account_type"`
	ActivatedAt sql.NullTime     `json:"activated_at"`
	StudentData *StudentResponse `json:"student_data,omitempty"`
	TeacherData *TeacherResponse `json:"teacher_data,omitempty"`
}
