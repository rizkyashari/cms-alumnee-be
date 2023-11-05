package repo_log

import (
	"errors"
	"fmt"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"gorm.io/gorm"
)

type LogRepo interface {
	InsertLog(logData model.LogData) error
	UpdateLogActivity(logData model.LogData) error
	GetActivityLogs(params *rq.PaginationParams[model.LogData]) (*[]model.LogData, int, int, error)
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) LogRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) InsertLog(logData model.LogData) error {
	if err := r.db.Create(&logData).Error; err != nil {
		return err
	}
	return nil
}

func (r *impRepo) UpdateLogActivity(logData model.LogData) error {
	if r.db == nil {
		return errors.New("database is not initialized")
	}

	// Update the log data in the database
	if err := r.db.Save(&logData).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) GetActivityLogs(params *rq.PaginationParams[model.LogData]) (*[]model.LogData, int, int, error) {

	var logs []model.LogData

	chain := r.db

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&logs), params.Limit)

	fmt.Println(params.Data.UserRole)
	if params.Data.UserRole != 0 {
		chain = chain.Where(r.db.Where("user_role = ?", params.Data.UserRole))
	}

	validColumnTimeStamp := []string{
		"timestamp",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnTimeStamp)).Order("created_at desc").Find(&logs)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &logs, maxPage, rowCount, nil

}
