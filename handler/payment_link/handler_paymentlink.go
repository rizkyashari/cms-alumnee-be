package handler

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

type PaymentLinkHandler interface {
	GetPaymentLinksByEmail(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) PaymentLinkHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) GetPaymentLinksByEmail(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}
	email := c.DefaultQuery("email", "")

	filters := model.PaymentLink{
		CustomerDetails: model.CustomerDetails{
			Email: email,
		},
	}

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	params := rq.PaginationParams[model.PaymentLink]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data:      filters,
	}

	res, err := h.s.PaymentLink().GetPaymentLinksByEmailFromAPI(c, gotAccount.Email, &params)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}
