package rs

type TeacherResponse struct {
	Name     string            `json:"name"`
	NIP      string            `json:"nip"`
	Subjects []SubjectResponse `json:"subject"`
}
