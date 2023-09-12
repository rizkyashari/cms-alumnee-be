package service_school

import (
	"context"
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

type SchoolService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.School]) (*rs.PaginationResponse[any, rs.SchoolResponse], error)
	GetDetailByID(c context.Context, id string) (*rs.SchoolResponse, error)

	CreateOne(c context.Context, newSchool *rq.SchoolRequest) error

	EditOne(c context.Context, newSchool *rq.SchoolRequest) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) SchoolService {
	return &impService{
		repo: r,
	}
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[model.School]) (*rs.PaginationResponse[any, rs.SchoolResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotSchools, maxPage, err := s.repo.School().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.SchoolResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotSchools == nil {
		return &rs.PaginationResponse[any, rs.SchoolResponse]{
			Data: []rs.SchoolResponse{},
		}, nil
	}

	if len(*gotSchools) < 1 {
		return &rs.PaginationResponse[any, rs.SchoolResponse]{
			Data: []rs.SchoolResponse{},
		}, nil
	}

	var datares []rs.SchoolResponse
	err = copier.Copy(&datares, gotSchools)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res := rs.PaginationResponse[any, rs.SchoolResponse]{
		MaxPage:         maxPage,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            datares,
	}

	return &res, nil
}

func (s *impService) GetDetailByID(c context.Context, id string) (*rs.SchoolResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	gotSchool, err := s.repo.School().GetDetailByID(parsedID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.SchoolResponse
	err = copier.Copy(&res, gotSchool)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}

func (s *impService) CreateOne(c context.Context, newSchool *rq.SchoolRequest) error {
	if len(newSchool.Name) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
	}

	school := model.School{
		Name: newSchool.Name,
	}

	err := s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.School().CreateOne(tx, &school); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) EditOne(c context.Context, newSchool *rq.SchoolRequest) error {
	if len(newSchool.Name) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
	}

	if newSchool.ID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "ID"}
	}

	id, err := uuid.Parse(*newSchool.ID)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	school := model.School{
		Base: model.Base{ID: id},
		Name: newSchool.Name,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.School().UpdateOne(tx, &school); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
