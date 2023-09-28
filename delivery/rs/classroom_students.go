package rs

type ClassroomStudentsResponse struct {
	Classrooms []ClassroomResponse `json:"classrooms"`
	Students   []StudentResponse   `json:"students"`
}
