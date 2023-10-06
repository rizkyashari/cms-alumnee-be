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
	serviceutil "github.com/fadhln/lms-be/util/service_util"
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

	gotYears, maxPage, rowCount, err := s.repo.AcademicYear().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.AcademicYearResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotYears == nil {
		return &rs.PaginationResponse[any, rs.AcademicYearResponse]{
			Data: []rs.AcademicYearResponse{},
		}, nil
	}

	if len(*gotYears) < 1 {
		return &rs.PaginationResponse[any, rs.AcademicYearResponse]{
			Data: []rs.AcademicYearResponse{},
		}, nil
	}

	var datares []rs.AcademicYearResponse
	err = copier.Copy(&datares, gotYears)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res := rs.PaginationResponse[any, rs.AcademicYearResponse]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            datares,
	}

	return &res, nil
}

func (s *impService) GetDetailByID(c context.Context, id string) (*rs.AcademicYearResponse, error) {
	parsedAcademicYearID, err := serviceutil.GetUUIDFromStringWithValidation("Academic Year ID", &id)
	if err != nil {
		return nil, err
	}

	gotYear, err := s.repo.AcademicYear().GetDetailByID(*parsedAcademicYearID)
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
	if newYear.Year == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Year"}
	}

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
	if newYear.Year == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Year"}
	}

	if len(*newYear.Year) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Year"}
	}

	parsedAcademicYearID, err := serviceutil.GetUUIDFromStringWithValidation("Academic Year ID", newYear.ID)
	if err != nil {
		return err
	}

	year := model.AcademicYear{
		Base: model.Base{ID: *parsedAcademicYearID},
		Year: *newYear.Year,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.AcademicYear().UpdateOne(tx, *parsedAcademicYearID, &year); err != nil {
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
	parsedAcademicYearID, err := serviceutil.GetUUIDFromStringWithValidation("Academic Year ID", newYear.ID)
	if err != nil {
		return err
	}

	year := model.AcademicYear{
		Base:     model.Base{ID: *parsedAcademicYearID},
		IsActive: *newYear.IsActive,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.AcademicYear().UpdateStatus(tx, *parsedAcademicYearID, &year); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
