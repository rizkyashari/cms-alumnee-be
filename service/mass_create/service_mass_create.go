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
	serviceutil "github.com/fadhln/lms-be/util/service_util"
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

func (s *impService) convertToResponse(record *model.MassCreate) (*rs.MassCreateResponse, error) {
	var tempResponse rs.MassCreateResponse

	err := copier.Copy(&tempResponse, record)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	errMsgsStr := record.ErrorMessages
	var errMsg []model.ErrorMsg

	json.Unmarshal([]byte(errMsgsStr), &errMsg)

	tempResponse.ErrorMessages = errMsg

	return &tempResponse, nil
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

	gotRecords, maxPage, rowCount, err := s.repo.MassCreate().GetAll(params)
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
		tempResponse, err := s.convertToResponse(&record)
		if err != nil {
			return nil, err
		}

		response = append(response, *tempResponse)
	}

	res := rs.PaginationResponse[any, rs.MassCreateResponse]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) GetDetailByID(c context.Context, id string) (*rs.MassCreateResponse, error) {
	parsedMassCreateID, err := serviceutil.GetUUIDFromStringWithValidation("Mass Create ID", &id)
	if err != nil {
		return nil, err
	}

	gotRecords, err := s.repo.MassCreate().GetDetailByID(*parsedMassCreateID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res, err := s.convertToResponse(gotRecords)
	if err != nil {
		return nil, err
	}

	return res, nil
}
