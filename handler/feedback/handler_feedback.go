package handler_feedback

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
	"github.com/google/uuid"
)

type FeedbackHandler interface {
	GetAll(c *gin.Context)

	GetAllFeedbackQuestions(c *gin.Context)

	GetAllOwnStudent(c *gin.Context)

	GetAllOwnTeacher(c *gin.Context)

	GetDetailByID(c *gin.Context)

	CreateOne(c *gin.Context)

	CreateMultiple(c *gin.Context)

	CreateOneFeedbackQuestion(c *gin.Context)

	EditOne(c *gin.Context)

	EditFeedbackQuestion(c *gin.Context)

	EditFeedback(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) FeedbackHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) CreateOne(c *gin.Context) {
	var request rq.FeedbackRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Feedback().CreateOne(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateMultiple(c *gin.Context) {
	var requests []*rq.FeedbackRequest
	err := c.ShouldBindJSON(&requests)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	for _, request := range requests {
		err = h.s.Feedback().CreateOne(c, request)
		if err != nil {
			rs.ErrorResponse(c, err)
			return
		}
	}
	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateOneFeedbackQuestion(c *gin.Context) {
	var request rq.FeedbackQuestionRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Feedback().CreateOneFeedbackQuestion(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) EditOne(c *gin.Context) {
	id := c.Param("id")

	var request rq.FeedbackRequest = rq.FeedbackRequest{ID: &id}
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Feedback().EditOne(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditFeedbackQuestion(c *gin.Context) {
	id := c.Param("id")

	var request rq.FeedbackQuestionRequest = rq.FeedbackQuestionRequest{ID: &id}
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Feedback().EditFeedbackQuestion(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditFeedback(c *gin.Context) {

	id := c.Param("id")

	var request rq.FeedbackRequest = rq.FeedbackRequest{ID: &id}
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Feedback().EditFeedback(c, &request)
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

	searchAcademicYearID := c.DefaultQuery("academic_year_id", "")
	parsedAcademicYearID, _ := uuid.Parse(searchAcademicYearID)

	searchStudentID := c.DefaultQuery("student_id", "")
	parsedStudentID, _ := uuid.Parse(searchStudentID)

	searchTeacherID := c.DefaultQuery("teacher_id", "")
	parsedTeacherID, _ := uuid.Parse(searchTeacherID)

	filters := model.Feedback{
		AcademicYearID: parsedAcademicYearID,
		StudentID:      parsedStudentID,
		TeacherID:      parsedTeacherID,
		FeedbackScores: []model.FeedbackScore{},
	}

	if len(searchStudentID) >= 2 {
		parsedStudentID, _ := uuid.Parse(searchStudentID)
		filters.StudentID = parsedStudentID
	}

	if len(searchTeacherID) >= 2 {
		parsedTeacherID, _ := uuid.Parse(searchTeacherID)
		filters.TeacherID = parsedTeacherID
	}

	if len(searchAcademicYearID) >= 2 {
		parsedAcademicYearID, _ := uuid.Parse(searchAcademicYearID)
		filters.AcademicYearID = parsedAcademicYearID
	}

	params := rq.PaginationParams[model.Feedback]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data:      filters,
	}

	if len(searchAcademicYearID) >= 2 {
		parsedAcademicYearID, _ := uuid.Parse(searchAcademicYearID)
		params.Data.AcademicYearID = parsedAcademicYearID
	}

	res, err := h.s.Feedback().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllFeedbackQuestions(c *gin.Context) {
	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchQuestion := c.DefaultQuery("question", "")

	params := rq.PaginationParams[model.FeedbackQuestion]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.FeedbackQuestion{
			Question: searchQuestion,
		},
	}

	res, err := h.s.Feedback().GetAllFeedbackQuestions(c, &params)
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

	searchAcademicYearID := c.DefaultQuery("academic_year_id", "")
	parsedAcademicYearID, _ := uuid.Parse(searchAcademicYearID)

	searchTeacherID := c.DefaultQuery("teacher_id", "")
	parsedTeacherID, _ := uuid.Parse(searchTeacherID)

	filters := model.Feedback{
		AcademicYearID: parsedAcademicYearID,
		TeacherID:      parsedTeacherID,
		FeedbackScores: []model.FeedbackScore{},
	}

	if len(searchTeacherID) >= 2 {
		parsedTeacherID, _ := uuid.Parse(searchTeacherID)
		filters.TeacherID = parsedTeacherID
	}

	if len(searchAcademicYearID) >= 2 {
		parsedAcademicYearID, _ := uuid.Parse(searchAcademicYearID)
		filters.AcademicYearID = parsedAcademicYearID
	}

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	params := rq.PaginationParams[model.Feedback]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data:      filters,
	}

	res, err := h.s.Feedback().GetAllByStudentID(c, gotAccount.Student.ID.String(), &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllOwnTeacher(c *gin.Context) {
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

	searchAcademicYearID := c.DefaultQuery("academic_year_id", "")
	parsedAcademicYearID, _ := uuid.Parse(searchAcademicYearID)

	searchStudentID := c.DefaultQuery("student_id", "")
	parsedStudentID, _ := uuid.Parse(searchStudentID)

	filters := model.Feedback{
		AcademicYearID: parsedAcademicYearID,
		StudentID:      parsedStudentID,
		FeedbackScores: []model.FeedbackScore{},
	}

	if len(searchStudentID) >= 2 {
		parsedStudentID, _ := uuid.Parse(searchStudentID)
		filters.StudentID = parsedStudentID
	}

	if len(searchAcademicYearID) >= 2 {
		parsedAcademicYearID, _ := uuid.Parse(searchAcademicYearID)
		filters.AcademicYearID = parsedAcademicYearID
	}

	params := rq.PaginationParams[model.Feedback]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data:      filters,
	}

	res, err := h.s.Feedback().GetAllByTeacherID(c, gotAccount.Teacher.ID.String(), &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetDetailByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.s.Feedback().GetDetailByID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}
