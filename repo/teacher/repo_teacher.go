package repo_teacher

import (
	"github.com/fadhln/lms-be/model"
	"gorm.io/gorm"
)

type TeacherRepo interface {
	CreateOne(newTeacher *model.Teacher) (*model.Teacher, error)
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) TeacherRepo {
	return &impRepo{
		db: db,
	}
}
