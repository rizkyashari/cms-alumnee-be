package service

import "github.com/fadhln/lms-be/repo"

type Service interface{}

type impService struct {
	repo repo.Repository
}

func SetupService(r repo.Repository) Service {
	return &impService{
		repo: r,
	}
}
