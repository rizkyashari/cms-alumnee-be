package handler

import (
	auth "github.com/fadhln/lms-be/handler/auth"
	school "github.com/fadhln/lms-be/handler/school"
	"github.com/fadhln/lms-be/service"
)

type Handler struct {
	s service.Service

	Auth   auth.AuthHandler
	School school.SchoolHandler
}

func SetupHandler(s service.Service) *Handler {
	return &Handler{
		s:      s,
		Auth:   auth.Init(s),
		School: school.Init(s),
	}
}
