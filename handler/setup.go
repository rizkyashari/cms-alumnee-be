package handler

import (
	auth "github.com/fadhln/lms-be/handler/auth"
	"github.com/fadhln/lms-be/service"
)

type Handler struct {
	s service.Service

	Auth auth.AuthHandler
}

func SetupHandler(s service.Service) *Handler {
	return &Handler{
		s:    s,
		Auth: auth.Init(s),
	}
}
