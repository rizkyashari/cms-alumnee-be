package repo_student

import (
	"github.com/fadhln/lms-be/model"
	"gorm.io/gorm"
)

type StudentRepo interface {
	CreateOne(tx *gorm.DB, newStudent *model.Student) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) StudentRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) CreateOne(tx *gorm.DB, newStudent *model.Student) error {
	if err := tx.Create(newStudent).Error; err != nil {
		return err
	}

	return nil
}
