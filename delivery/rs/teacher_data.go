package rs

import (
	"time"

	"github.com/google/uuid"
)

type TeacherDataResponse struct {
	ID               uuid.UUID `json:"id"`
	TeacherID        uuid.UUID `json:"teacher_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Gender           int       `json:"gender"`
	NIK              string    `json:"nik"`
	NUPTK            string    `json:"nuptk"`
	NIP              string    `json:"nip"`
	EmploymentStatus int       `json:"employment_status"`
	BirthPlace       string    `json:"birth_place"`
	BirthDate        time.Time `json:"birth_date"`
	PhoneNumber      string    `json:"phone_number"`
}
