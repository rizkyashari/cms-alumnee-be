package repo_teacher

import (
	"github.com/fadhln/lms-be/model"
	"gorm.io/gorm"
)

type TeacherRepo interface {
	CreateOne(tx *gorm.DB, newTeacher *model.Teacher) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) TeacherRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) CreateOne(tx *gorm.DB, newTeacher *model.Teacher) error {
	if err := tx.Create(newTeacher).Error; err != nil {
		return err
	}

	return nil
}
