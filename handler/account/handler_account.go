package handler_account

import (
	"net/http"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
)

type AccountHandler interface {
	UploadAvatar(c *gin.Context)
	UploadOwnAvatar(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) AccountHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) UploadAvatar(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	accountID := c.Param("account_id")

	var imagefile rq.ImageUploadRequest
	if err := c.ShouldBind(&imagefile); err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, errmsg.ErrRequestFileInvalid)
		return
	}

	if imagefile.ImageFile == nil {
		util.SaveResponseBody(c, errmsg.ErrRequestFileInvalid, nil)
		rs.ErrorResponse(c, errmsg.ErrRequestFileInvalid)
		return
	}

	err := h.s.Account().UploadAvatar(c, accountID, imagefile.ImageFile)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, http.StatusCreated)
}

func (h *impHandler) UploadOwnAvatar(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, 0)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	var imagefile rq.ImageUploadRequest
	if err := c.ShouldBind(&imagefile); err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestFileInvalid)
		return
	}

	if imagefile.ImageFile == nil {
		rs.ErrorResponse(c, errmsg.ErrRequestFileInvalid)
		return
	}

	err = h.s.Account().UploadAvatar(c, gotAccount.ID.String(), imagefile.ImageFile)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, http.StatusCreated)
}
