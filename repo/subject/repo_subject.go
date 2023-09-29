package repo_subject

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubjectRepo interface {
	GetDetailByID(id uuid.UUID) (*model.Subject, error)
	GetAll(params *rq.PaginationParams[model.Subject]) (*[]model.Subject, int, int, error)
	CreateOne(tx *gorm.DB, newSubject *model.Subject) error
	UpdateOne(tx *gorm.DB, id uuid.UUID, newSubject *model.Subject) error
	DeleteOne(tx *gorm.DB, id uuid.UUID) error

	GetComponentDetailByComponentID(componentId uuid.UUID) (*model.SubjectComponent, error)
	GetAllComponent(params *rq.PaginationParams[model.SubjectComponent]) (*[]model.SubjectComponent, int, int, error)
	CreateOneComponent(tx *gorm.DB, newSubjectComp *model.SubjectComponent) error
	UpdateOneComponent(tx *gorm.DB, componentId uuid.UUID, newSubjectComp *model.SubjectComponent) error
	DeleteOneComponent(tx *gorm.DB, id uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) SubjectRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetDetailByID(id uuid.UUID) (*model.Subject, error) {
	var subject model.Subject
	chain := r.db.Preload("SubjectComponents")

	if err := chain.Where("id = ?", id).Find(&subject).Error; err != nil {
		return nil, err
	}

	return &subject, nil
}

func (r *impRepo) GetAll(params *rq.PaginationParams[model.Subject]) (*[]model.Subject, int, int, error) {
	var subjects []model.Subject

	chain := r.db.Preload("SubjectComponents")

	if len(params.Data.Name) >= 2 {
		chain = chain.Where(r.db.Where("name ILIKE " + `'%` + params.Data.Name + `%'`))
	}

	if params.Data.TeacherID != uuid.Nil {
		chain = chain.Where(r.db.Where("teacher_id = ?", params.Data.TeacherID.String()))
	}

	// For future reference: index 0 is for include, index 1 is for exclude
	if len(params.Data.RelationClassroomSubjects) >= 1 {
		relationParams := params.Data.RelationClassroomSubjects[0]
		if relationParams.ClassroomID != uuid.Nil {
			chain = chain.Where(r.db.Where("id IN (?)",
				r.db.Debug().Model(&model.RelationClassroomSubject{}).Where("classroom_id = ?", relationParams.ClassroomID).Distinct("subject_id").Select("subject_id")))
		}
	}

	if len(params.Data.RelationClassroomSubjects) >= 2 {
		excludeRelationParams := params.Data.RelationClassroomSubjects[1]
		if excludeRelationParams.ClassroomID != uuid.Nil {
			chain = chain.Where(r.db.Where("id IN (?)",
				r.db.Debug().Model(&model.RelationClassroomSubject{}).Where("classroom_id != ?", excludeRelationParams.ClassroomID).Distinct("subject_id").Select("subject_id")))
		}
	}

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&subjects), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"name",
	}

	result := chain.Scopes(
		util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&subjects)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &subjects, maxPage, rowCount, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, subject *model.Subject) error {
	if err := tx.Create(subject).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, id uuid.UUID, subject *model.Subject) error {
	if subject == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Subject"}
	}

	result := tx.Model(subject).Where("id = ?", id).Updates(subject)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, id uuid.UUID) error {
	if id == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.Subject{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) GetComponentDetailByComponentID(componentId uuid.UUID) (*model.SubjectComponent, error) {
	var SubjectComponent model.SubjectComponent
	if err := r.db.Where("id = ?", componentId).Find(&SubjectComponent).Error; err != nil {
		return nil, err
	}

	return &SubjectComponent, nil
}

func (r *impRepo) GetAllComponent(params *rq.PaginationParams[model.SubjectComponent]) (*[]model.SubjectComponent, int, int, error) {
	var subjectComps []model.SubjectComponent

	chain := r.db

	if len(params.Data.Name) >= 2 {
		chain = chain.Where(r.db.Where("name ILIKE " + `'%` + params.Data.Name + `%'`))
	}

	if params.Data.SubjectID != uuid.Nil {
		chain = chain.Where(r.db.Where("subject_id = ?", params.Data.SubjectID.String()))
	}

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&subjectComps), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"name",
	}

	result := chain.Scopes(
		util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&subjectComps)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &subjectComps, maxPage, rowCount, nil
}

func (r *impRepo) CreateOneComponent(tx *gorm.DB, newSubjectComp *model.SubjectComponent) error {
	if err := tx.Create(newSubjectComp).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOneComponent(tx *gorm.DB, componentId uuid.UUID, newSubjectComp *model.SubjectComponent) error {
	if newSubjectComp == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Subject Component"}
	}

	result := tx.Model(newSubjectComp).Where("id = ?", componentId).Updates(newSubjectComp)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) DeleteOneComponent(tx *gorm.DB, id uuid.UUID) error {
	if id == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.SubjectComponent{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
