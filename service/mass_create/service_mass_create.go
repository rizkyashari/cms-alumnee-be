package service_masscreate

import (
	"context"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
)

type MassCreateService interface {
	GetAll(c context.Context, params *rq.PaginationParams[any]) (*rs.PaginationResponse[any, rs.MassCreateResponse], error)
}

// TODO: Create implementation
