package service_auth

import (
	"context"
	"strings"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/auth"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService interface {
	CheckEmailExist(context.Context, *rq.EmailOnlyRequest) (*rs.IsExistResponsse, error)
	Login(context.Context, *rq.LoginRequest) (*rs.TokenResponse, error)
	Register(context.Context, *rq.RegisterRequest) (*rs.StatusResponse, error)
	GetOwnAccountDetail(*gin.Context) (*rs.AccountResponse, error)
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) AuthService {
	return &impService{
		repo: r,
	}
}

func (s *impService) CheckEmailExist(c context.Context, body *rq.EmailOnlyRequest) (*rs.IsExistResponsse, error) {
	if len(body.Email) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Email"}
	}

	email := strings.ToLower(body.Email)

	if !(util.IsEmailValid(email)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Email"}
	}

	gotAccount, _ := s.repo.Account().ReadOneByEmail(email)

	return &rs.IsExistResponsse{IsExist: gotAccount != nil}, nil
}

func (s *impService) Login(c context.Context, body *rq.LoginRequest) (*rs.TokenResponse, error) {
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

	jwt, err := auth.GenerateJWT(gotAccount.Email, gotAccount.AccountType)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &rs.TokenResponse{Jwt: jwt.Token, ExpiresAt: jwt.ExpiresAt}, nil
}

func (s *impService) Register(c context.Context, body *rq.RegisterRequest) (*rs.StatusResponse, error) {
	if len(body.Email) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Email"}
	}
	if len(body.Password) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Password"}
	}

	if body.Password != body.ConfirmPassword {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Confirm Password"}
	}

	if !(util.IsValidConstant(body.AccountType, constants.AccountTypeMap)) {
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

	newID := uuid.New()

	newAccount := model.Account{
		Base:        model.Base{ID: newID},
		Name:        body.Name,
		Email:       email,
		Password:    hashedPassword,
		AccountType: body.AccountType,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Account().CreateOne(tx, &newAccount); err != nil {
			return err
		}

		switch newAccount.AccountType {
		case constants.ACCOUNT_ADMIN:
			newAdmin := model.Admin{
				AccountID: newID,
			}
			if err := s.repo.Admin().CreateOne(tx, &newAdmin); err != nil {
				return err
			}

		case constants.ACCOUNT_STUDENT:
			newStudent := model.Student{
				AccountID: newID,
				Scores:    []model.Score{},
			}
			if err := s.repo.Student().CreateOne(tx, &newStudent); err != nil {
				return err
			}

		case constants.ACCOUNT_TEACHER:
			newTeacher := model.Teacher{
				AccountID: newID,
				Subjects:  []model.Subject{},
			}
			if err := s.repo.Teacher().CreateOne(tx, &newTeacher); err != nil {
				return err
			}

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

func (s *impService) GetOwnAccountDetail(c *gin.Context) (*rs.AccountResponse, error) {
	gotUser, err := util.GetAccountContext(c, 0)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return gotUser, nil
}
