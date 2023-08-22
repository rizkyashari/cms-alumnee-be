package rs

type TeacherResponse struct {
	Name     string            `json:"name"`
	NIK      *string           `json:"nik,omitempty"`
	Subjects []SubjectResponse `json:"subject"`
}
