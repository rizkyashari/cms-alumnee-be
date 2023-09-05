package seed

import (
	"context"
	"os"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/database"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/service"
)

func SeedingProcess(server *database.RepoServer) error {
	repository := repo.SetupRepo(server)
	s := service.SetupService(repository)

	// Seed Admin
	adminName := "Admin"
	newAdminRequest := rq.RegisterRequest{
		Email:           "admin@gmail.com",
		Password:        os.Getenv("DEFAULT_PASSWORD"),
		ConfirmPassword: os.Getenv("DEFAULT_PASSWORD"),
		AccountType:     constants.ACCOUNT_ADMIN,
		Name:            &adminName,
	}

	_, err := s.Auth().Register(context.Background(), &newAdminRequest)
	if err != nil {
		return err
	}

	// Seed School
	err = s.School().CreateOne(context.Background(), &rq.SchoolRequest{
		Name: "Madrasah Ibtidaiyah",
	})
	if err != nil {
		return err
	}

	err = s.School().CreateOne(context.Background(), &rq.SchoolRequest{
		Name: "Madrasah Tsanawiyah",
	})
	if err != nil {
		return err
	}

	err = s.School().CreateOne(context.Background(), &rq.SchoolRequest{
		Name: "Madrasah Aliyah",
	})
	if err != nil {
		return err
	}

	return nil
}
