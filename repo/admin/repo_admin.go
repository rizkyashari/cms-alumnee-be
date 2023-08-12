package repo_admin

import (
	"github.com/fadhln/lms-be/model"
	"gorm.io/gorm"
)

type AdminRepo interface {
	CreateOne(tx *gorm.DB, newAdmin *model.Admin) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) AdminRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) CreateOne(tx *gorm.DB, newAdmin *model.Admin) error {
	if err := tx.Create(newAdmin).Error; err != nil {
		return err
	}

	return nil
}
