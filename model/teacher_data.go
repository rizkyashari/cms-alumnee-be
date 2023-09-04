package model

import (
	"time"

	"github.com/google/uuid"
)

type TeacherData struct {
	Base
	TeacherID        uuid.UUID
	Gender           *int
	NIK              *string
	NUPTK            *string
	NIP              *string
	EmploymentStatus *int
	BirthPlace       *string
	BirthDate        *time.Time
	PhoneNumber      *string
}
