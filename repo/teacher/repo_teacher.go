package repo_teacher

import (
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeacherRepo interface {
	GetDetailByTeacherID(teacherId uuid.UUID) (*model.Teacher, error)
	GetDetailByAccountID(id uuid.UUID) (*model.Teacher, error)

	CreateOne(tx *gorm.DB, newTeacher *model.Teacher) error

	UpdateOne(tx *gorm.DB, id uuid.UUID, newTeacher *model.Teacher) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) TeacherRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetDetailByTeacherID(teacherId uuid.UUID) (*model.Teacher, error) {
	var teacher model.Teacher
	if err := r.db.Where("teacher_id = ?", teacherId).Find(&teacher).Error; err != nil {
		return nil, err
	}

	return &teacher, nil
}

func (r *impRepo) GetDetailByAccountID(id uuid.UUID) (*model.Teacher, error) {
	var teacher model.Teacher
	if err := r.db.Where("account_id = ?", id).Find(&teacher).Error; err != nil {
		return nil, err
	}

	return &teacher, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newTeacher *model.Teacher) error {
	if err := tx.Create(newTeacher).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, id uuid.UUID, newTeacher *model.Teacher) error {
	if newTeacher == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher"}
	}

	result := tx.Model(newTeacher).Where("id = ?", id).Updates(newTeacher)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
