package handler_subject

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

type SubjectHandler interface {
	GetAll(c *gin.Context)
	GetAllByClassroomID(c *gin.Context)
	GetAllNotInClassroomID(c *gin.Context)
	GetAllOwnTeacher(c *gin.Context)
	GetAllOwnStudent(c *gin.Context)
	GetDetailByID(c *gin.Context)

	GetAllSubjectComponent(c *gin.Context)
	GetAllSubjectComponentBySubjectID(c *gin.Context)
	GetSubjectComponentDetailByID(c *gin.Context)

	CreateOne(c *gin.Context)
	CreateOneWithClassroomID(c *gin.Context)
	CreateOneSubjectComponent(c *gin.Context)
	CreateOneSubjectComponentWithValidation(c *gin.Context)
	CreateMass(c *gin.Context)

	EditOne(c *gin.Context)
	EditOneWithValidation(c *gin.Context)
	EditOneSubjectComponent(c *gin.Context)
	EditOneSubjectComponentWithValidation(c *gin.Context)

	AssignClassroomsToSubject(c *gin.Context)
	RemoveClassroomsFromSubject(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) SubjectHandler {
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

	searchName := c.DefaultQuery("name", "")

	params := rq.PaginationParams[model.Subject]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Subject{
			Name: searchName,
		},
	}

	res, err := h.s.Subject().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllByClassroomID(c *gin.Context) {
	classroomID := c.Param("classroom_id")

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchName := c.DefaultQuery("name", "")

	params := rq.PaginationParams[model.Subject]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Subject{
			Name: searchName,
		},
	}

	res, err := h.s.Subject().GetAllSubjectByClassroomID(c, classroomID, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllNotInClassroomID(c *gin.Context) {
	classroomID := c.Param("classroom_id")

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchName := c.DefaultQuery("name", "")

	params := rq.PaginationParams[model.Subject]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Subject{
			Name: searchName,
		},
	}

	res, err := h.s.Subject().GetAllSubjectNotInClassroom(c, classroomID, &params)
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

	searchName := c.DefaultQuery("name", "")

	params := rq.PaginationParams[model.Subject]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Subject{
			Name: searchName,
		},
	}

	res, err := h.s.Subject().GetAllSubjectByTeacherID(c, gotAccount.Teacher.ID.String(), &params)
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

	searchName := c.DefaultQuery("name", "")

	params := rq.PaginationParams[model.Subject]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Subject{
			Name: searchName,
		},
	}

	res, err := h.s.Subject().GetAllSubjectByClassroomID(c, gotAccount.Student.ClassroomID.String(), &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetDetailByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.s.Subject().GetDetailByID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllSubjectComponent(c *gin.Context) {
	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchName := c.DefaultQuery("name", "")

	params := rq.PaginationParams[model.SubjectComponent]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.SubjectComponent{
			Name: searchName,
		},
	}

	res, err := h.s.Subject().GetAllSubjectComponent(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllSubjectComponentBySubjectID(c *gin.Context) {
	subjectId := c.Param("id")

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchName := c.DefaultQuery("name", "")

	params := rq.PaginationParams[model.SubjectComponent]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.SubjectComponent{
			Name: searchName,
		},
	}

	res, err := h.s.Subject().GetAllSubjectComponentBySubjectID(c, subjectId, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetSubjectComponentDetailByID(c *gin.Context) {
	subjectCompID := c.Param("subject_component_id")
	res, err := h.s.Subject().GetSubjectComponentDetailByID(c, subjectCompID)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) CreateOne(c *gin.Context) {
	var request rq.SubjectRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Subject().CreateOne(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateOneWithClassroomID(c *gin.Context) {
	classroomID := c.Param("classroom_id")

	var request rq.SubjectRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Subject().CreateOneWithClassroomID(c, classroomID, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateOneSubjectComponent(c *gin.Context) {
	var request rq.SubjectComponentRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Subject().CreateOneSubjectComponent(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateOneSubjectComponentWithValidation(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	var request rq.SubjectComponentRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Subject().CreateOneSubjectComponentWithValidation(c, gotAccount.Teacher.ID.String(), &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
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

	res, err := h.s.Subject().CreateMass(c, csvfile.CSVFile)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusCreated)
}

func (h *impHandler) EditOne(c *gin.Context) {
	id := c.Param("id")

	var request rq.SubjectRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	request.ID = &id

	err = h.s.Subject().EditOne(c, id, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditOneWithValidation(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	id := c.Param("id")

	var request rq.SubjectRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	request.ID = &id

	err = h.s.Subject().EditOneWithValidation(c, gotAccount.Teacher.ID.String(), id, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditOneSubjectComponent(c *gin.Context) {
	subjectCompID := c.Param("subject_component_id")

	var request rq.SubjectComponentRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Subject().EditOneSubjectComponent(c, subjectCompID, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditOneSubjectComponentWithValidation(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	subjectCompID := c.Param("subject_component_id")

	var request rq.SubjectComponentRequest
	err = c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Subject().EditOneSubjectComponentWithValidation(c, gotAccount.Teacher.ID.String(), subjectCompID, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) AssignClassroomsToSubject(c *gin.Context) {
	subjectId := c.Param("id")

	var request rq.IDsRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Subject().AssignClassroomsToSubject(c, subjectId, request.IDs)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) RemoveClassroomsFromSubject(c *gin.Context) {
	subjectId := c.Param("id")

	var request rq.IDsRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Subject().RemoveClassroomsFromSubject(c, subjectId, request.IDs)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusNoContent)
}
