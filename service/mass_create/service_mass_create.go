package service_masscreate

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type MassCreateService interface {
	GetAll(c context.Context, params *rq.PaginationParams[any]) (*rs.PaginationResponse[any, rs.MassCreateResponse], error)
	GetDetailByID(c context.Context, id string) (*rs.MassCreateResponse, error)
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) MassCreateService {
	return &impService{
		repo: r,
	}
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[any]) (*rs.PaginationResponse[any, rs.MassCreateResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotRecords, maxPage, err := s.repo.MassCreate().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.MassCreateResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotRecords == nil {
		return &rs.PaginationResponse[any, rs.MassCreateResponse]{
			Data: []rs.MassCreateResponse{},
		}, nil
	}

	if len(*gotRecords) < 1 {
		return &rs.PaginationResponse[any, rs.MassCreateResponse]{
			Data: []rs.MassCreateResponse{},
		}, nil
	}

	var response []rs.MassCreateResponse

	for _, record := range *gotRecords {
		var tempResponse rs.MassCreateResponse

		err = copier.Copy(&tempResponse, record)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		errMsgsStr := record.ErrorMessages
		var errMsg []model.ErrorMsg

		json.Unmarshal([]byte(errMsgsStr), &errMsg)

		tempResponse.ErrorMessages = errMsg
		response = append(response, tempResponse)
	}

	res := rs.PaginationResponse[any, rs.MassCreateResponse]{
		MaxPage:         maxPage,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) GetDetailByID(c context.Context, id string) (*rs.MassCreateResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	gotRecords, err := s.repo.MassCreate().GetDetailByID(parsedID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.MassCreateResponse
	err = copier.Copy(&res, gotRecords)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}
