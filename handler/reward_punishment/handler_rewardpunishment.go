package handler_rewardpunishment

import (
	"net/http"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
)

type RewardPunishmentHandler interface {
	GetAll(c *gin.Context)

	GetAllOwnStudent(c *gin.Context)

	GetAllRewardPunishmentForTeacher(c *gin.Context)

	CreateOne(c *gin.Context)

	EditOne(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) RewardPunishmentHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) CreateOne(c *gin.Context) {
	var request rq.RewardPunishmentRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.RewardPunishment().CreateOne(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) EditOne(c *gin.Context) {
	id := c.Param("id")

	var request rq.RewardPunishmentRequest = rq.RewardPunishmentRequest{ID: &id}
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.RewardPunishment().EditOne(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) GetAll(c *gin.Context) {
	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchDescription := c.DefaultQuery("description", "")

	params := rq.PaginationParams[model.RewardPunishment]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.RewardPunishment{
			Description: &searchDescription,
		},
	}

	res, err := h.s.RewardPunishment().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllOwnStudent(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchDescription := c.DefaultQuery("description", "")

	params := rq.PaginationParams[model.RewardPunishment]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.RewardPunishment{
			Description: &searchDescription,
		},
	}

	res, err := h.s.RewardPunishment().GetAllByStudentID(c, gotAccount.Student.ID.String(), &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllRewardPunishmentForTeacher(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchDescription := c.DefaultQuery("description", "")

	params := rq.PaginationParams[model.RewardPunishment]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.RewardPunishment{
			Description: &searchDescription,
		},
	}

	rewardPunishments, err := h.s.RewardPunishment().GetAllRewardPunishmentForTeacher(c, gotAccount.Teacher.ID.String(), &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, rewardPunishments, http.StatusOK)
}
