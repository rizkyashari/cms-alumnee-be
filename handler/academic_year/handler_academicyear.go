package handler_academicyear

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

type AcademicYearHandler interface {
	GetAll(c *gin.Context)
	GetDetailByID(c *gin.Context)

	CreateOne(c *gin.Context)

	EditYear(c *gin.Context)
	EditStatus(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) AcademicYearHandler {
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

	searchYear := c.DefaultQuery("year", "")

	params := rq.PaginationParams[model.AcademicYear]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.AcademicYear{
			Year: searchYear,
		},
	}

	res, err := h.s.AcademicYear().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetDetailByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.s.AcademicYear().GetDetailByID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) CreateOne(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	var request rq.AcademicYearRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.AcademicYear().CreateOne(c, &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) EditYear(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	id := c.Param("id")

	var request rq.AcademicYearRequest
	err := c.ShouldBindJSON(&request)
	request.ID = &id

	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.AcademicYear().EditYear(c, &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditStatus(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	id := c.Param("id")

	var request rq.AcademicYearRequest
	err := c.ShouldBindJSON(&request)
	request.ID = &id

	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.AcademicYear().EditStatus(c, &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}
