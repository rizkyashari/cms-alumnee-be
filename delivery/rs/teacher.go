package rs

type TeacherResponse struct {
	NIK      *string           `json:"nik,omitempty"`
	Subjects []SubjectResponse `json:"subject"`
}
