package repo_masscreate

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MassCreateRepo interface {
	GetAll(params *rq.PaginationParams[any]) (*[]model.MassCreate, int, int, error)
	GetDetailByID(id uuid.UUID) (*model.MassCreate, error)
	CreateOne(tx *gorm.DB, newCreation *model.MassCreate) error
	UpdateOne(tx *gorm.DB, id uuid.UUID, newCreation *model.MassCreate) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) MassCreateRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetAll(params *rq.PaginationParams[any]) (*[]model.MassCreate, int, int, error) {
	var entries []model.MassCreate

	chain := r.db

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&entries), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&entries)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &entries, maxPage, rowCount, nil
}

func (r *impRepo) GetDetailByID(id uuid.UUID) (*model.MassCreate, error) {
	var masscreate model.MassCreate
	if err := r.db.Where("id = ?", id).First(&masscreate).Error; err != nil {
		return nil, err
	}

	return &masscreate, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newCreation *model.MassCreate) error {
	if err := tx.Create(newCreation).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, id uuid.UUID, newCreation *model.MassCreate) error {
	if newCreation == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Mass Creation"}
	}

	result := tx.Model(newCreation).Where("id = ?", id).Updates(newCreation)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
