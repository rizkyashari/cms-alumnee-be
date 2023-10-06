package rq

type SubjectRequest struct {
	ID        *string `json:"id,omitempty"`
	Name      *string `json:"name,omitempty"`
	TeacherID *string `json:"teacher_id,omitempty"`
}

type SubjectRequestWithTeacherEmail struct {
	Name         string
	TeacherEmail string
}

type SubjectComponentRequest struct {
	ID         *string `json:"id,omitempty"`
	Name       *string `json:"name,omitempty"`
	Percentage *int    `json:"percentage,omitempty"`
	SubjectID  *string `json:"subject_id,omitempty"`
}
