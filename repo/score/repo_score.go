package repo_score

import (
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScoreRepo interface {
	GetByStudentIDAndSubjectComponentID(studentID uuid.UUID, subjectComponentID uuid.UUID) (*[]model.Score, error)
	SaveAll(tx *gorm.DB, newScores *[]model.Score) error
	DeleteAllForStudentAndComponentID(tx *gorm.DB, studentID uuid.UUID, subjectComponentID uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) ScoreRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetByStudentIDAndSubjectComponentID(studentID uuid.UUID, subjectComponentID uuid.UUID) (*[]model.Score, error) {
	var scores []model.Score

	chain := r.db

	chain = chain.Where(r.db.Where("student_id = ? ", studentID))
	chain = chain.Where(r.db.Where("subject_component_id = ? ", subjectComponentID))

	result := chain.Find(&scores)

	if result.Error != nil {
		return nil, result.Error
	}

	return &scores, nil
}

func (r *impRepo) SaveAll(tx *gorm.DB, newScores *[]model.Score) error {
	if newScores == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Scores"}
	}

	result := tx.Create(newScores)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) DeleteAllForStudentAndComponentID(tx *gorm.DB, studentID uuid.UUID, subjectComponentID uuid.UUID) error {
	result := tx.Delete(&model.Score{}, "student_id = ? AND subject_component_id = ?", studentID, subjectComponentID)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
