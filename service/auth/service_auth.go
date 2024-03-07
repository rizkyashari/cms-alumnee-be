package service_auth

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type AuthService interface {
	CheckEmailExist(context.Context, *rq.EmailOnlyRequest) (*rs.IsExistResponsse, error)
	Login(context.Context, *rq.LoginRequest) (*rs.TokenResponse, error)
	Register(context.Context, *rq.RegisterRequest) (*rs.StatusResponse, error)
	GetOwnAccountDetail(*gin.Context) (*rs.AccountResponse, error)
	ForgotPassword(context.Context, *rq.EmailOnlyRequest) (*rs.StatusResponse, error)
	ResetPassword(context.Context, *rq.ResetPasswordRequest) (*rs.StatusResponse, error)
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

func (s *impService) ForgotPassword(c context.Context, body *rq.EmailOnlyRequest) (*rs.StatusResponse, error) {
	if len(body.Email) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Email"}
	}

	email := strings.ToLower(body.Email)

	if !(util.IsEmailValid(email)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Email"}
	}

	account, err := s.repo.Account().ReadOneByEmail(email)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if account == nil {
		return nil, &errmsg.ErrNotFound{FieldName: "User"}
	}

	// Generate reset password token and set expiry time
	token := util.RandString(32)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Store the reset password request
	_, err = s.repo.ResetPassword().CreateResetPassword(account.ID, token, expiresAt)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	// Send reset password email to the user with the token
	err = s.sendResetPasswordEmail(email, token)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &rs.StatusResponse{Status: "Reset password email sent successfully"}, nil
}

func (s *impService) ResetPassword(c context.Context, body *rq.ResetPasswordRequest) (*rs.StatusResponse, error) {

	if len(body.Token) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Token"}
	}
	if len(body.NewPassword) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "New Password"}
	}
	if len(body.ConfirmPassword) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Confirm Password"}
	}
	if body.NewPassword != body.ConfirmPassword {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Confirm Password"}
	}

	// Retrieve reset password request by token
	resetPassword, err := s.repo.ResetPassword().GetResetPasswordByToken(body.Token)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if resetPassword == nil {
		return nil, &errmsg.ErrNotFound{FieldName: "Reset Password Request"}
	}

	// Check if the token has expired
	if time.Now().After(resetPassword.ResetPasswordExpires) {
		return nil, &errmsg.ErrFieldIsExpired{FieldName: "Reset Password Token"}
	}

	// Update user's password
	hashedPassword, err := auth.HashAndSalt(body.NewPassword)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	accountID := resetPassword.AccountID
	// Update the password for the account with the specified accountID
	if err := s.repo.Account().UpdatePassword(c, accountID, hashedPassword); err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	// Delete the reset password request from the database
	if err := s.repo.ResetPassword().DeleteResetPassword(body.Token); err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &rs.StatusResponse{Status: "Password reset successful"}, nil
}

func (s *impService) sendResetPasswordEmail(email, token string) error {

	SMTP_EMAIL := os.Getenv("SMTP_EMAIL")
	SMTP_SERVER := os.Getenv("SMTP_SERVER")
	SMTP_PASSWORD := os.Getenv("SMTP_PASSWORD")
	SMTP_PORT, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}
	RESET_PASSWORD_HOST_URL := os.Getenv("RESET_PASSWORD_HOST_URL")

	// Compose the email
	m := gomail.NewMessage()
	m.SetHeader("From", SMTP_EMAIL)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Reset Password Akun SIMA (Sistem Informasi Al-Muddatsiriyah)")

	// Create the reset password link
	resetLink := fmt.Sprintf(RESET_PASSWORD_HOST_URL+"reset-password?token=%s", token)

	// Compose the email body
	body := fmt.Sprintf("Klik link di bawah ini untuk melakukan reset password:\n\n%s\n\nPERHATIAN! Token akan kadaluarsa setelah 24 jam.", resetLink)
	m.SetBody("text/plain", body)

	// Create a new SMTP client
	d := gomail.NewDialer(SMTP_SERVER, SMTP_PORT, SMTP_EMAIL, SMTP_PASSWORD)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
