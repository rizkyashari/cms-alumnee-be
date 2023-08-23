package handler_school

import (
	"net/http"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
)

type SchoolHandler interface {
	GetAll(c *gin.Context)
	GetDetailByID(c *gin.Context)

	CreateOne(c *gin.Context)

	EditOne(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) SchoolHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) GetAll(c *gin.Context) {
	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchName := c.DefaultQuery("name", "")

	params := rq.PaginationParams[model.School]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.School{
			Name: searchName,
		},
	}

	res, err := h.s.School().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetDetailByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.s.School().GetDetailByID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) CreateOne(c *gin.Context) {
	var request rq.SchoolRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.School().CreateOne(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) EditOne(c *gin.Context) {
	id := c.Param("id")

	var request rq.SchoolRequest = rq.SchoolRequest{ID: &id}
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.School().EditOne(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}
