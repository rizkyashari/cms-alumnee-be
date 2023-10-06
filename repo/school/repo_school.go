package repo_school

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SchoolRepo interface {
	GetDetailByID(id uuid.UUID) (*model.School, error)
	GetAll(params *rq.PaginationParams[model.School]) (*[]model.School, int, int, error)
	CreateOne(tx *gorm.DB, newSchool *model.School) error
	UpdateOne(tx *gorm.DB, newSchool *model.School) error
	DeleteOne(tx *gorm.DB, id uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) SchoolRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetDetailByID(id uuid.UUID) (*model.School, error) {
	var school model.School
	if err := r.db.Where("id = ?", id).Find(&school).Error; err != nil {
		return nil, err
	}

	return &school, nil
}

func (r *impRepo) GetAll(params *rq.PaginationParams[model.School]) (*[]model.School, int, int, error) {
	var schools []model.School

	chain := r.db

	if len(params.Data.Name) >= 2 {
		chain = chain.Where(r.db.Where("name ILIKE " + `'%` + params.Data.Name + `%'`))
	}

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&schools), params.Limit)

	validColumnName := []string{
		"name",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).Find(&schools)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &schools, maxPage, rowCount, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newSchool *model.School) error {
	if err := tx.Create(newSchool).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, newSchool *model.School) error {
	if newSchool == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Article"}
	}

	result := tx.Model(newSchool).Updates(newSchool)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, id uuid.UUID) error {
	if id == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.School{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
