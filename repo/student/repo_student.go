package repo_student

import (
	"github.com/fadhln/lms-be/model"
	"gorm.io/gorm"
)

type StudentRepo interface {
	CreateOne(newStudent *model.Student) (*model.Student, error)
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) StudentRepo {
	return &impRepo{
		db: db,
	}
}
