package rewardpunishment

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RewardPunishmentRepo interface {
	GetAll(params *rq.PaginationParams[model.RewardPunishment]) (*[]model.RewardPunishment, int, int, error)
	GetRewardPunishmentsByStudentID(studentID uuid.UUID, params *rq.PaginationParams[model.RewardPunishment]) ([]model.RewardPunishment, int, int, error)
	CreateOne(tx *gorm.DB, newRewardPunishment *model.RewardPunishment) error
	UpdateOne(tx *gorm.DB, newRewardPunishment *model.RewardPunishment) error
	DeleteOne(tx *gorm.DB, id uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) RewardPunishmentRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetAll(params *rq.PaginationParams[model.RewardPunishment]) (*[]model.RewardPunishment, int, int, error) {
	var rewardPunishments []model.RewardPunishment

	chain := r.db

	if params.Data.StudentID != uuid.Nil {
		chain = chain.Where(r.db.Where("student_id = ?", params.Data.StudentID.String()))
	}

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&rewardPunishments), params.Limit)

	validColumDescription := []string{
		"description",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumDescription)).Find(&rewardPunishments)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &rewardPunishments, maxPage, rowCount, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newRewardPunishment *model.RewardPunishment) error {
	if err := tx.Create(newRewardPunishment).Error; err != nil {
		return err
	}
	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, newRewardPunishment *model.RewardPunishment) error {
	if newRewardPunishment == nil {
		return &errmsg.ErrIsEmpty{FieldName: "RewardPunishment"}
	}

	result := tx.Model(newRewardPunishment).Updates(newRewardPunishment)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, id uuid.UUID) error {
	if id == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.RewardPunishment{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) GetRewardPunishmentsByStudentID(studentID uuid.UUID, params *rq.PaginationParams[model.RewardPunishment]) ([]model.RewardPunishment, int, int, error) {
	var rewardPunishments []model.RewardPunishment

	chain := r.db.Where("student_id = ?", studentID.String())

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&rewardPunishments), params.Limit)

	validColumnDescription := []string{
		"description",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnDescription)).Find(&rewardPunishments)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return rewardPunishments, maxPage, rowCount, nil
}
