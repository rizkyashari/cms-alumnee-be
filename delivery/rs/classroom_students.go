package rs

type ClassroomStudentsResponse struct {
	Classrooms []ClassroomResponse `json:"classrooms"`
	Students   []AccountResponse   `json:"students"`
}
