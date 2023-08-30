package service_academicyear

import (
	"context"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
)

type AcademicYearService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.AcademicYear]) (*rs.PaginationResponse[any, rs.AcademicYearResponse], error)
	GetDetailByID(c context.Context, id string) (*rs.AcademicYearResponse, error)

	CreateOne(c context.Context, newSchool *rq.AcademicYearResponse) error

	EditOne(c context.Context, newSchool *rq.AcademicYearResponse) error
}

// TODO: Create Implementation
