package repo_student

import (
	"github.com/fadhln/lms-be/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepo interface {
	GetDetailByAccountID(id uuid.UUID) (*model.Student, error)

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

func (r *impRepo) GetDetailByAccountID(id uuid.UUID) (*model.Student, error) {
	var student model.Student
	if err := r.db.Where("account_id = ?", id).Find(&student).Error; err != nil {
		return nil, err
	}

	return &student, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newStudent *model.Student) error {
	if err := tx.Create(newStudent).Error; err != nil {
		return err
	}

	return nil
}
