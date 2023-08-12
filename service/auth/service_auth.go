package service_auth

import (
	"strings"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/auth"
	"github.com/fadhln/lms-be/util/errmsg"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(*rq.LoginRequest) (*rs.TokenResponse, error)
	Register(*rq.RegisterRequest) (*rs.StatusResponse, error)
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) AuthService {
	return &impService{
		repo: r,
	}
}

func (s *impService) Login(body *rq.LoginRequest) (*rs.TokenResponse, error) {
	if len(body.Email) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Email"}
	}
	if len(body.Password) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Password"}
	}

	email := strings.ToLower(body.Email)
	password := body.Password

	if !(util.IsEmailValid(email)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Email"}
	}

	gotAccount, err := s.repo.Account().ReadOneByEmail(email)
	if gotAccount == nil {
		return nil, &errmsg.ErrNotFound{FieldName: "User"}
	}
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if !(auth.ComparePassword(gotAccount.Password, password)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Password"}
	}

	jwt, err := auth.GenerateJWT(email)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &rs.TokenResponse{Jwt: jwt.Token, ExpiresAt: jwt.ExpiresAt}, nil
}

func (s *impService) Register(body *rq.RegisterRequest) (*rs.StatusResponse, error) {
	if len(body.Email) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Email"}
	}
	if len(body.Password) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Password"}
	}

	if body.Password != body.ConfirmPassword {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Confirm Password"}
	}

	if !(model.IsValidAccountType(body.AccountType)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Account Type"}
	}

	email := strings.ToLower(body.Email)
	if !(util.IsEmailValid(email)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Email"}
	}

	hashedPassword, err := auth.HashAndSalt(body.Password)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	newAccount := model.Account{
		Email:       email,
		Password:    hashedPassword,
		AccountType: body.AccountType,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Account().CreateOne(tx, &newAccount); err != nil {
			return err
		}

		gotAcc, err := s.repo.Account().ReadOneByEmail(newAccount.Email)
		if err != nil {
			return err
		}

		switch gotAcc.AccountType {
		case model.Adm:
			newAdmin := model.Admin{
				AccountID: gotAcc.ID,
			}
			if err := s.repo.Admin().CreateOne(tx, &newAdmin); err != nil {
				return err
			}
			break

		case model.Stu:
			newStudent := model.Student{
				AccountID: gotAcc.ID,
				Name:      body.Name,
				Scores:    []model.Score{},
			}
			if err := s.repo.Student().CreateOne(tx, &newStudent); err != nil {
				return err
			}
			break

		case model.Tch:
			newTeacher := model.Teacher{
				AccountID: gotAcc.ID,
				Name:      body.Name,
				NIP:       "",
				Subjects:  []model.Subject{},
			}
			if err := s.repo.Teacher().CreateOne(tx, &newTeacher); err != nil {
				return err
			}
			break

		default:
			return &errmsg.ErrFieldIsWrong{FieldName: "Account Type"}
		}

		return nil
	})

	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &rs.StatusResponse{Status: "User registered successfully"}, nil
}
