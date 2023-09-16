package repo_relationclassroomsubject

import (
	"github.com/fadhln/lms-be/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RelationClassroomSubjectRepo interface {
	CreateOne(tx *gorm.DB, newRelation *model.RelationClassroomSubject) error
	DeleteOne(tx *gorm.DB, classroomID uuid.UUID, subjectID uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) RelationClassroomSubjectRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) CreateOne(tx *gorm.DB, newRelation *model.RelationClassroomSubject) error {
	if err := tx.Create(newRelation).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, classroomID uuid.UUID, subjectID uuid.UUID) error {
	result := tx.Where("classroom_id = ?", classroomID).Where("subject_id", subjectID).Delete(&model.RelationClassroomSubject{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected < 0 {
		return gorm.ErrEmptySlice
	}

	return nil
}
