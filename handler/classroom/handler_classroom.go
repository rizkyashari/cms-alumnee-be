package handler_classroom

import (
	"net/http"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
)

type ClassroomHandler interface {
	CreateMass(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) ClassroomHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) CreateMass(c *gin.Context) {
	var csvfile rq.CSVFileUploadRequest
	if err := c.ShouldBind(&csvfile); err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestFileInvalid)
		return
	}

	if csvfile.CSVFile == nil {
		rs.ErrorResponse(c, errmsg.ErrRequestFileInvalid)
		return
	}

	res, err := h.s.Classroom().CreateMass(c, csvfile.CSVFile)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusCreated)
}
