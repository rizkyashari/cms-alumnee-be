package service

import (
	"github.com/fadhln/lms-be/repo"
	account "github.com/fadhln/lms-be/service/account"
	auth "github.com/fadhln/lms-be/service/auth"
	classroom "github.com/fadhln/lms-be/service/classroom"
	school "github.com/fadhln/lms-be/service/school"
)

type Service interface {
	Auth() auth.AuthService
	Account() account.AccountService
	Classroom() classroom.ClassroomService
	School() school.SchoolService
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

func (s *impService) Classroom() classroom.ClassroomService {
	return classroom.Init(s.repo)
}

func (s *impService) School() school.SchoolService {
	return school.Init(s.repo)
}
