package model

type School struct {
	Base
	Name       string
	Teachers   []Teacher
	Classrooms []Classroom
}
