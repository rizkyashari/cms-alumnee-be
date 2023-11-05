package repo_eventdraft

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventDraftRepo interface {
	GetAll(params *rq.PaginationParams[model.EventDraft]) (*[]model.EventDraft, int, int, error)
	CreateOrSaveOne(tx *gorm.DB, newDraft model.EventDraft) error
	DeleteOne(tx *gorm.DB, draftID uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) EventDraftRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetAll(params *rq.PaginationParams[model.EventDraft]) (*[]model.EventDraft, int, int, error) {
	var drafts []model.EventDraft

	chain := r.db
	if params.Data.SchoolID != uuid.Nil {
		chain = chain.Where(r.db.Where("school_id = ?", params.Data.SchoolID))
	}

	if params.Data.AcademicYearID != uuid.Nil {
		chain = chain.Where(r.db.Where("academic_year_id = ?", params.Data.AcademicYearID))
	}

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&drafts), params.Limit)

	validColumnName := []string{
		"updated_at",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, "updated_at", params.SortOrder, validColumnName)).Find(&drafts)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &drafts, maxPage, rowCount, nil
}

func (r *impRepo) CreateOrSaveOne(tx *gorm.DB, newDraft model.EventDraft) error {
	result := tx.Save(&newDraft)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, draftID uuid.UUID) error {
	if draftID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.EventDraft{}, draftID)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
