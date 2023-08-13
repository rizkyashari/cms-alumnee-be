package model

type AcademicYear struct {
	Base
	Year       string
	Classrooms []Classroom
}
