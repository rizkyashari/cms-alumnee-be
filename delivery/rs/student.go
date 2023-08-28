package rs

type StudentResponse struct {
	Classroom ClassroomResponse `json:"classroom"`
	Scores    []ScoreResponse   `json:"score"`
}
