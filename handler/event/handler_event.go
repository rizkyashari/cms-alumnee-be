package handler_event

import (
	"net/http"
	"time"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
)

type EventHandler interface {
	GetDetailByID(c *gin.Context)
	GetAllByClassroomID(c *gin.Context)
	GetAllByTeacherID(c *gin.Context)
	GetAllByStudentID(c *gin.Context)
	GetAllByOwnTeacherID(c *gin.Context)
	GetAllByOwnStudentID(c *gin.Context)
	CreateOne(c *gin.Context)
	CreateRepeated(c *gin.Context)
	CreateManyRepeated(c *gin.Context)
	EditOne(c *gin.Context)
	DeleteOne(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) EventHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) GetDetailByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.s.Event().GetDetailByID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllByClassroomID(c *gin.Context) {
	classroomID := c.Param("classroom_id")

	begin := c.DefaultQuery("begin", "")
	end := c.DefaultQuery("end", "")

	beginTime, err := time.Parse(time.RFC3339Nano, begin)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	endTime, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	beginTime = util.ForceCorrectTimeZone(beginTime)
	endTime = util.ForceCorrectTimeZone(endTime)

	params := rq.EventParams{
		Begin: beginTime,
		End:   endTime,
	}

	res, err := h.s.Event().GetAllByClassroomID(c, classroomID, params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllByTeacherID(c *gin.Context) {
	teacherID := c.Param("teacher_id")

	begin := c.DefaultQuery("begin", "")
	end := c.DefaultQuery("end", "")

	beginTime, err := time.Parse(time.RFC3339Nano, begin)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	endTime, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	beginTime = util.ForceCorrectTimeZone(beginTime)
	endTime = util.ForceCorrectTimeZone(endTime)

	params := rq.EventParams{
		Begin: beginTime,
		End:   endTime,
	}

	res, err := h.s.Event().GetAllByTeacherID(c, teacherID, params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllByStudentID(c *gin.Context) {
	studentID := c.Param("student_id")

	begin := c.DefaultQuery("begin", "")
	end := c.DefaultQuery("end", "")

	beginTime, err := time.Parse(time.RFC3339Nano, begin)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	endTime, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	beginTime = util.ForceCorrectTimeZone(beginTime)
	endTime = util.ForceCorrectTimeZone(endTime)

	params := rq.EventParams{
		Begin: beginTime,
		End:   endTime,
	}

	res, err := h.s.Event().GetAllByStudentID(c, studentID, params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllByOwnTeacherID(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	begin := c.DefaultQuery("begin", "")
	end := c.DefaultQuery("end", "")

	beginTime, err := time.Parse(time.RFC3339Nano, begin)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	endTime, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	beginTime = util.ForceCorrectTimeZone(beginTime)
	endTime = util.ForceCorrectTimeZone(endTime)

	params := rq.EventParams{
		Begin: beginTime,
		End:   endTime,
	}

	res, err := h.s.Event().GetAllByTeacherID(c, gotAccount.Teacher.ID.String(), params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllByOwnStudentID(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	begin := c.DefaultQuery("begin", "")
	end := c.DefaultQuery("end", "")

	beginTime, err := time.Parse(time.RFC3339Nano, begin)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	endTime, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	beginTime = util.ForceCorrectTimeZone(beginTime)
	endTime = util.ForceCorrectTimeZone(endTime)

	params := rq.EventParams{
		Begin: beginTime,
		End:   endTime,
	}

	res, err := h.s.Event().GetAllByStudentID(c, gotAccount.Student.ID.String(), params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) CreateOne(c *gin.Context) {
	var request rq.CreateSingleEventRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Event().CreateOne(c, request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateRepeated(c *gin.Context) {
	var request rq.CreateRepeatWeeklyEventRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Event().CreateRepeated(c, request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateManyRepeated(c *gin.Context) {
	var request rq.CreateManyRepeatEventRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Event().CreateManyRepeated(c, request.Events)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) EditOne(c *gin.Context) {
	eventID := c.Param("id")

	var request rq.EditEventRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	request.ID = eventID

	err = h.s.Event().EditOne(c, request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) DeleteOne(c *gin.Context) {
	eventID := c.Param("id")

	err := h.s.Event().DeleteOne(c, eventID)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusOK)
}
