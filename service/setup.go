package service

import (
	"github.com/fadhln/lms-be/repo"
	auth "github.com/fadhln/lms-be/service/auth"
)

type Service interface {
	Auth() auth.AuthService
}

type impService struct {
	repo repo.Repository
}

func SetupService(r repo.Repository) Service {
	return &impService{
		repo: r,
	}
}

func (s *impService) Auth() auth.AuthService {
	return auth.Init(s.repo)
}
