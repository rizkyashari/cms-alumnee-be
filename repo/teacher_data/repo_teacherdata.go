package repo_teacherdata

import (
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeacherDataRepo interface {
	GetTeacherDataByTeacherID(id uuid.UUID) (*model.TeacherData, error)
	UpdateOne(tx *gorm.DB, newTeacherData *model.TeacherData) error
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
	if err := r.db.Where("teacher_id = ?", id).Find(&teacherData).Error; err != nil {
		return nil, err
	}

	return &teacherData, nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, newTeacherData *model.TeacherData) error {
	if newTeacherData == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher Data"}
	}

	result := tx.Save(newTeacherData)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
