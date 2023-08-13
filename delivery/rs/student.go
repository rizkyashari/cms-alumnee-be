package rs

type StudentResponse struct {
	Name      string            `json:"name"`
	Classroom ClassroomResponse `json:"classroom"`
	Scores    []ScoreResponse   `json:"score"`
}
