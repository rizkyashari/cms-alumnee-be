package service_academicyear

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

type AcademicYearService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.AcademicYear]) (*rs.PaginationResponse[any, rs.AcademicYearResponse], error)
	GetDetailByID(c context.Context, id string) (*rs.AcademicYearResponse, error)

	CreateOne(c context.Context, newYear *rq.AcademicYearRequest) error

	EditYear(c context.Context, newYear *rq.AcademicYearRequest) error
	EditStatus(c context.Context, newYear *rq.AcademicYearRequest) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) AcademicYearService {
	return &impService{
		repo: r,
	}
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[model.AcademicYear]) (*rs.PaginationResponse[any, rs.AcademicYearResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotYears, maxPage, err := s.repo.AcademicYear().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.AcademicYearResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	var datares []rs.AcademicYearResponse
	err = copier.Copy(&datares, gotYears)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res := rs.PaginationResponse[any, rs.AcademicYearResponse]{
		MaxPage:         maxPage,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            datares,
	}

	return &res, nil
}

func (s *impService) GetDetailByID(c context.Context, id string) (*rs.AcademicYearResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	gotYear, err := s.repo.AcademicYear().GetDetailByID(parsedID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.AcademicYearResponse
	err = copier.Copy(&res, gotYear)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}

func (s *impService) CreateOne(c context.Context, newYear *rq.AcademicYearRequest) error {
	if len(*newYear.Year) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Year"}
	}

	year := model.AcademicYear{
		Year:     *newYear.Year,
		IsActive: *newYear.IsActive,
	}

	err := s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.AcademicYear().CreateOne(tx, &year); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) EditYear(c context.Context, newYear *rq.AcademicYearRequest) error {
	if newYear.ID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "ID"}
	}

	if len(*newYear.Year) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Year"}
	}

	id, err := uuid.Parse(*newYear.ID)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	year := model.AcademicYear{
		Base: model.Base{ID: id},
		Year: *newYear.Year,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.AcademicYear().UpdateOne(tx, id, &year); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) EditStatus(c context.Context, newYear *rq.AcademicYearRequest) error {
	if newYear.ID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "ID"}
	}

	id, err := uuid.Parse(*newYear.ID)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	year := model.AcademicYear{
		Base:     model.Base{ID: id},
		IsActive: *newYear.IsActive,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.AcademicYear().UpdateStatus(tx, id, &year); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
