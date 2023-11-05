package handler_score

import (
	"net/http"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	"github.com/gin-gonic/gin"
)

type ScoreHandler interface {
	GetTotalScoreForSubjectIDAndStudentID(c *gin.Context)
	GetTotalScoreForSubjectIDAndOwnStudent(c *gin.Context)
	GetByStudentIDAndComponentID(c *gin.Context)
	SaveForStudentID(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) ScoreHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) GetTotalScoreForSubjectIDAndStudentID(c *gin.Context) {
	studentID := c.DefaultQuery("student_id", "")
	subjectID := c.DefaultQuery("subject_id", "")

	res, err := h.s.Score().GetTotalScoreForSubjectIDAndStudentID(subjectID, studentID)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetTotalScoreForSubjectIDAndOwnStudent(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}
	subjectID := c.DefaultQuery("subject_id", "")

	res, err := h.s.Score().GetTotalScoreForSubjectIDAndStudentID(subjectID, gotAccount.Student.ID.String())
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)

}

func (h *impHandler) GetByStudentIDAndComponentID(c *gin.Context) {
	studentID := c.DefaultQuery("student_id", "")
	componentID := c.DefaultQuery("component_id", "")

	res, err := h.s.Score().GetByStudentIDAndComponentID(studentID, componentID)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) SaveForStudentID(c *gin.Context) {
	var request rq.StudentScoreRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Score().SaveForStudentID(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}
