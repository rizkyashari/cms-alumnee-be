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
	GetAll(params *rq.PaginationParams[model.Subject]) (*[]model.Subject, int, error)
	CreateOne(tx *gorm.DB, newSubject *model.Subject) error
	UpdateOne(tx *gorm.DB, newSubject *model.Subject) error
	DeleteOne(tx *gorm.DB, id uuid.UUID) error
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
	if err := r.db.Where("id = ?", id).Find(&subject).Error; err != nil {
		return nil, err
	}

	return &subject, nil
}

func (r *impRepo) GetAll(params *rq.PaginationParams[model.Subject]) (*[]model.Subject, int, error) {
	var subjects []model.Subject

	chain := r.db

	if len(params.Data.Name) >= 2 {
		chain = chain.Where(r.db.Where("name ILIKE " + `'%` + params.Data.Name + `%'`))
	}

	if params.Data.TeacherID != uuid.Nil {
		chain = chain.Where(r.db.Where("teacher_id = " + params.Data.TeacherID.String()))
	}

	maxPage := util.GetMaxPage(chain.Find(&subjects), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"name",
	}

	result := chain.Scopes(
		util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&subjects)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return &subjects, maxPage, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, subject *model.Subject) error {
	if err := tx.Create(subject).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, subject *model.Subject) error {
	if subject == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Subject"}
	}

	result := tx.Model(subject).Updates(subject)
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
