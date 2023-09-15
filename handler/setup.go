package handler

import (
	academicyear "github.com/fadhln/lms-be/handler/academic_year"
	auth "github.com/fadhln/lms-be/handler/auth"
	classroom "github.com/fadhln/lms-be/handler/classroom"
	masscreate "github.com/fadhln/lms-be/handler/mass_create"
	school "github.com/fadhln/lms-be/handler/school"
	student "github.com/fadhln/lms-be/handler/student"
	subject "github.com/fadhln/lms-be/handler/subject"
	teacher "github.com/fadhln/lms-be/handler/teacher"
	"github.com/fadhln/lms-be/service"
)

type Handler struct {
	s service.Service

	Auth         auth.AuthHandler
	AcademicYear academicyear.AcademicYearHandler
	Classroom    classroom.ClassroomHandler
	MassCreate   masscreate.MassCreateHandler
	School       school.SchoolHandler
	Student      student.StudentHandler
	Subject      subject.SubjectHandler
	Teacher      teacher.TeacherHandler
}

func SetupHandler(s service.Service) *Handler {
	return &Handler{
		s:            s,
		Auth:         auth.Init(s),
		AcademicYear: academicyear.Init(s),
		Classroom:    classroom.Init(s),
		MassCreate:   masscreate.Init(s),
		School:       school.Init(s),
		Student:      student.Init(s),
		Subject:      subject.Init(s),
		Teacher:      teacher.Init(s),
	}
}
