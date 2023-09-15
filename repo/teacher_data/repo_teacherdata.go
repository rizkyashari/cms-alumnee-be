package repo_teacherdata

import (
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeacherDataRepo interface {
	GetTeacherDataByTeacherID(id uuid.UUID) (*model.TeacherData, error)
	CreateOne(tx *gorm.DB, newTeacherData *model.TeacherData) error
	UpdateOne(tx *gorm.DB, teacherID uuid.UUID, newTeacherData *model.TeacherData) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) TeacherDataRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetTeacherDataByTeacherID(id uuid.UUID) (*model.TeacherData, error) {
	var teacherData model.TeacherData
	if err := r.db.Where("teacher_id = ?", id).First(&teacherData).Error; err != nil {
		return nil, err
	}

	return &teacherData, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newTeacherData *model.TeacherData) error {
	if err := tx.Create(newTeacherData).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, teacherID uuid.UUID, newTeacherData *model.TeacherData) error {
	if newTeacherData == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher Data"}
	}

	result := tx.Model(newTeacherData).Where("teacher_id = ?", teacherID).Updates(newTeacherData)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
