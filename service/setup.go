package service

import (
	"github.com/fadhln/lms-be/repo"
	account "github.com/fadhln/lms-be/service/account"
	auth "github.com/fadhln/lms-be/service/auth"
)

type Service interface {
	Auth() auth.AuthService
	Account() account.AccountService
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

func (s *impService) Account() account.AccountService {
	return account.Init(s.repo)
}
