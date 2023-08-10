package handler

import "github.com/fadhln/lms-be/service"

type Handler struct {
	s service.Service
}

func SetupHandler(s service.Service) *Handler {
	return &Handler{
		s: s,
	}
}
