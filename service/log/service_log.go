package service_log

import (
	"context"
	"errors"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type LogService interface {
	LogActivity(c context.Context, logData model.LogData) error
	UpdateLogActivity(c context.Context, logData model.LogData) error
	GetActivityLogs(c context.Context, params *rq.PaginationParams[model.LogData]) (*rs.PaginationResponse[any, rs.LogDataResponse], error)
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) LogService {
	return &impService{
		repo: r,
	}
}

func (s *impService) LogActivity(c context.Context, logData model.LogData) error {
	return s.repo.Log().InsertLog(logData)
}

func (s *impService) UpdateLogActivity(c context.Context, logData model.LogData) error {
	return s.repo.Log().UpdateLogActivity(logData)
}

func (s *impService) GetActivityLogs(c context.Context, params *rq.PaginationParams[model.LogData]) (*rs.PaginationResponse[any, rs.LogDataResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotLogs, maxPage, rowCount, err := s.repo.Log().GetActivityLogs(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.LogDataResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotLogs == nil {
		return &rs.PaginationResponse[any, rs.LogDataResponse]{
			Data: []rs.LogDataResponse{},
		}, nil
	}

	if len(*gotLogs) < 1 {
		return &rs.PaginationResponse[any, rs.LogDataResponse]{
			Data: []rs.LogDataResponse{},
		}, nil
	}

	var datares []rs.LogDataResponse
	err = copier.Copy(&datares, gotLogs)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res := rs.PaginationResponse[any, rs.LogDataResponse]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            datares,
	}

	return &res, nil
}
