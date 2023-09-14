package service_subject

import (
	"context"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
)

type SubjectService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.School]) (*rs.PaginationResponse[any, rs.SchoolResponse], error)
	GetDetailByID(c context.Context, id string) (*rs.SchoolResponse, error)

	CreateOne(c context.Context, newSchool *rq.SchoolRequest) error

	EditOne(c context.Context, newSchool *rq.SchoolRequest) error
}
