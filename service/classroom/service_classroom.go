package service_classroom

import (
	"context"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
)

type ClassroomService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.Classroom]) (*rs.PaginationResponse[any, rs.ClassroomResponse], error)
	GetDetailByID(c context.Context, id string) (*rs.ClassroomResponse, error)

	CreateOne(c context.Context, newClassroom *rq.ClassroomRequest) error
	CreateMass(c context.Context, requestFileUrl string, newClassrooms *[]rq.ClassroomRequest) (*rs.MassCreateResponse, error)

	EditOne(c context.Context, newClassroom *rq.ClassroomRequest) error

	DeleteOne(c context.Context, idReq *rq.IDOnlyRequest) error
}

// TODO: Create Implementation
