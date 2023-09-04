package rq

type TeacherRegisterRequest struct {
	Account BaseRegisterRequest `json:"account"`
	Data    TeacherRequest      `json:"data"`
}

type TeacherRegisterWithDetailRequest struct {
	Account BaseRegisterRequest  `json:"account"`
	Data    TeacherUpdateRequest `json:"data"`
}

type TeacherRequest struct {
	SchoolID string `json:"school_id"`
}

type TeacherUpdateRequest struct {
	TeacherID        string  `json:"teacher_id"`
	SchoolID         *string `json:"school_id,omitempty"`
	Gender           *int    `json:"gender,omitempty"`
	NIK              *string `json:"nik,omitempty"`
	NUPTK            *string `json:"nuptk,omitempty"`
	NIP              *string `json:"nip,omitempty"`
	EmploymentStatus *int    `json:"employment_status,omitempty"`
	BirthPlace       *string `json:"birth_place,omitempty"`
	BirthDate        *string `json:"birth_date,omitempty"`
	PhoneNumber      *string `json:"phone_number,omitempty"`
}
