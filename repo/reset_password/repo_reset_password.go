package repo_reset_password

import (
	"time"

	"github.com/fadhln/lms-be/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResetPasswordRepo interface {
	CreateResetPassword(accountID uuid.UUID, token string, expiresAt time.Time) (*model.ResetPassword, error)
	GetResetPasswordByToken(token string) (*model.ResetPassword, error)
	DeleteResetPassword(token string) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) ResetPasswordRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) CreateResetPassword(accountID uuid.UUID, token string, expiresAt time.Time) (*model.ResetPassword, error) {
	resetPassword := &model.ResetPassword{
		AccountID:            accountID,
		ResetPasswordToken:   token,
		ResetPasswordExpires: expiresAt,
	}
	if err := r.db.Create(resetPassword).Error; err != nil {
		return nil, err
	}
	return resetPassword, nil
}

func (r *impRepo) GetResetPasswordByToken(token string) (*model.ResetPassword, error) {
	resetPassword := &model.ResetPassword{}
	if err := r.db.Where("reset_password_token = ?", token).First(resetPassword).Error; err != nil {
		return nil, err
	}
	return resetPassword, nil
}

func (r *impRepo) DeleteResetPassword(token string) error {
	resetPassword := &model.ResetPassword{}
	if err := r.db.Where("reset_password_token = ?", token).Delete(resetPassword).Error; err != nil {
		return err
	}
	return nil
}
