package handler_attendance

import (
	"net/http"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	serviceutil "github.com/fadhln/lms-be/util/service_util"
	"github.com/gin-gonic/gin"
)

type AttendanceHandler interface {
	GetAllForEventID(c *gin.Context)
	GetForOwnStudentWithClassroomIDAndSubjectID(c *gin.Context)
	CreateMany(c *gin.Context)
	EditOne(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) AttendanceHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) GetAllForEventID(c *gin.Context) {
	event_id := c.Param("id")
	parsedEventID, _ := serviceutil.GetUUIDFromStringWithValidation("Event ID", &event_id)

	res, err := h.s.Attendance().GetAll(c, &rq.GetAllAttendanceParams{
		EventID: parsedEventID,
	})

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetForOwnStudentWithClassroomIDAndSubjectID(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	classroomID := c.DefaultQuery("name", "")
	subjectID := c.DefaultQuery("teacher_id", "")

	parsedClassroomID, _ := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", &classroomID)
	parsedSubjectID, _ := serviceutil.GetUUIDFromStringWithValidation("Subject ID", &subjectID)

	res, err := h.s.Attendance().GetAll(c, &rq.GetAllAttendanceParams{
		StudentID:   &gotAccount.Student.ID,
		ClassroomID: parsedClassroomID,
		SubjectID:   parsedSubjectID,
	})

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) CreateMany(c *gin.Context) {
	var request rq.CreateAttendanceRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Attendance().CreateMany(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) EditOne(c *gin.Context) {
	var request rq.EditAttendanceRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Attendance().EditOne(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}
