package handler_eventdraft

import (
	"net/http"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventDraftHandler interface {
	GetAll(c *gin.Context)
	CreateOrSaveOne(c *gin.Context)
	DeleteOne(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) EventDraftHandler {
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

	searchAcademicYearID := c.DefaultQuery("academic_year_id", "")
	searchSchoolID := c.DefaultQuery("school_id", "")
	filters := model.EventDraft{}

	if len(searchAcademicYearID) >= 2 {
		parsedAcademicYearID, _ := uuid.Parse(searchAcademicYearID)
		filters.AcademicYearID = parsedAcademicYearID
	}

	if len(searchSchoolID) >= 2 {
		parsedSchoolID, _ := uuid.Parse(searchSchoolID)
		filters.SchoolID = parsedSchoolID
	}

	params := rq.PaginationParams[model.EventDraft]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data:      filters,
	}

	res, err := h.s.EventDraft().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) CreateOrSaveOne(c *gin.Context) {
	var request rq.EventDraftRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.EventDraft().CreateOrSaveOne(c,
		request.ID,
		request.SchoolID,
		request.AcademicYearID,
		request.StudyHourUnit,
		request.Events,
	)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) DeleteOne(c *gin.Context) {
	draftID := c.Param("id")

	err := h.s.EventDraft().DeleteOne(c, draftID)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusOK)
}
