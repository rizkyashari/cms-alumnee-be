package rq

type StudentScore struct {
	Value       float64 `json:"value"`
	Percentage  int     `json:"percentage"`
	Description string  `json:"description"`
}

type StudentScoreRequest struct {
	StudentID          string         `json:"student_id"`
	SubjectComponentID string         `json:"subject_component_id"`
	Scores             []StudentScore `json:"scores"`
}
