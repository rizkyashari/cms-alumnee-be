package handler

import (
	auth "github.com/fadhln/lms-be/handler/auth"
	classroom "github.com/fadhln/lms-be/handler/classroom"
	school "github.com/fadhln/lms-be/handler/school"
	"github.com/fadhln/lms-be/service"
)

type Handler struct {
	s service.Service

	Auth      auth.AuthHandler
	Classroom classroom.ClassroomHandler
	School    school.SchoolHandler
}

func SetupHandler(s service.Service) *Handler {
	return &Handler{
		s:         s,
		Auth:      auth.Init(s),
		Classroom: classroom.Init(s),
		School:    school.Init(s),
	}
}
