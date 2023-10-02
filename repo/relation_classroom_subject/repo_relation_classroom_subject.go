package repo_relationclassroomsubject

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RelationClassroomSubjectRepo interface {
	GetAll(params *rq.PaginationParams[model.RelationClassroomSubject]) (*[]model.RelationClassroomSubject, int, int, error)
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

func (r *impRepo) GetAll(params *rq.PaginationParams[model.RelationClassroomSubject]) (*[]model.RelationClassroomSubject, int, int, error) {
	var relations []model.RelationClassroomSubject

	chain := r.db.Preload("Classroom").Preload("Subject")

	if params.Data.Subject.TeacherID != uuid.Nil {
		chain = chain.Where(r.db.Where("subject_id IN (?)",
			r.db.Debug().Model(&model.Subject{}).Where("teacher_id = ?", params.Data.Subject.TeacherID).Select("id")))
	}

	if params.Data.Classroom.AcademicYearID != uuid.Nil {
		chain = chain.Where(r.db.Where("classroom_id IN (?)",
			r.db.Debug().Model(&model.Classroom{}).Where("academic_year_id = ?", params.Data.Classroom.AcademicYearID).Select("id")))
	}

	if params.Data.ClassroomID != uuid.Nil {
		chain = chain.Where(r.db.Where("classroom_id = ?", params.Data.ClassroomID))
	}

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&relations), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"name",
	}

	result := chain.Scopes(
		util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&relations)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &relations, maxPage, rowCount, nil
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
