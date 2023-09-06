package repo_academicyear

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AcademicYearRepo interface {
	GetDetailByID(id uuid.UUID) (*model.AcademicYear, error)
	GetAll(params *rq.PaginationParams[model.AcademicYear]) (*[]model.AcademicYear, int, error)
	CreateOne(tx *gorm.DB, newAcademicYear *model.AcademicYear) error
	UpdateOne(tx *gorm.DB, id uuid.UUID, newAcademicYear *model.AcademicYear) error
	UpdateStatus(tx *gorm.DB, id uuid.UUID, newAcademicYear *model.AcademicYear) error
	DeleteOne(tx *gorm.DB, id uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) AcademicYearRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetDetailByID(id uuid.UUID) (*model.AcademicYear, error) {
	var year model.AcademicYear
	if err := r.db.Where("id = ?", id).Find(&year).Error; err != nil {
		return nil, err
	}

	return &year, nil
}

func (r *impRepo) GetAll(params *rq.PaginationParams[model.AcademicYear]) (*[]model.AcademicYear, int, error) {
	var year []model.AcademicYear

	chain := r.db

	if len(params.Data.Year) >= 2 {
		chain = chain.Where(r.db.Where("year ILIKE " + `'%` + params.Data.Year + `%'`))
	}

	maxPage := util.GetMaxPage(chain.Find(&year), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"year",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).Find(&year)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return &year, maxPage, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, year *model.AcademicYear) error {
	if err := tx.Create(year).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, id uuid.UUID, year *model.AcademicYear) error {
	if year == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Academic Year"}
	}

	result := tx.Model(year).Where("id = ?", id).Updates(year)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) UpdateStatus(tx *gorm.DB, id uuid.UUID, year *model.AcademicYear) error {
	if year == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Academic Year"}
	}

	result := tx.Model(year).Where("id = ?", id).Select("IsActive").Updates(year)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, id uuid.UUID) error {
	if id == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.AcademicYear{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
