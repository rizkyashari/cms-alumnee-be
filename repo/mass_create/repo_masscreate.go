package repo_masscreate

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"gorm.io/gorm"
)

type MassCreateRepo interface {
	GetAll(params *rq.PaginationParams[model.MassCreate]) (*[]model.MassCreate, int, error)
	CreateOne(tx *gorm.DB, newCreation *model.MassCreate) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) MassCreateRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetAll(params *rq.PaginationParams[model.MassCreate]) (*[]model.MassCreate, int, error) {
	var entries []model.MassCreate

	chain := r.db

	maxPage := util.GetMaxPage(chain.Find(&entries), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&entries)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return &entries, maxPage, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newCreation *model.MassCreate) error {
	if err := tx.Create(newCreation).Error; err != nil {
		return err
	}

	return nil
}
