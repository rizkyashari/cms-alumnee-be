package service_auth

import (
	"strings"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/auth"
	"github.com/fadhln/lms-be/util/errmsg"
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
