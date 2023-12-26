package handler_classroom

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

type ClassroomHandler interface {
	GetAll(c *gin.Context)
	GetAllBySubjectID(c *gin.Context)
	GetAllByNotInSubjectID(c *gin.Context)
	GetAllByOwnTeacherID(c *gin.Context)
	GetDetailByID(c *gin.Context)

	CreateOne(c *gin.Context)
	CreateMass(c *gin.Context)

	EditOne(c *gin.Context)

	AssignSubjectsToClassroom(c *gin.Context)
	RemoveSubjectsFromClassroom(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) ClassroomHandler {
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
	searchTeacherID := c.DefaultQuery("teacher_id", "")
	searchSchoolID := c.DefaultQuery("school_id", "")
	academicYearID := c.DefaultQuery("academic_year_id", "")

	params := rq.PaginationParams[model.Classroom]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Classroom{
			Name: searchName,
		},
	}

	if len(searchTeacherID) >= 2 {
		parsedSearchTeacherID, _ := uuid.Parse(searchTeacherID)
		params.Data.TeacherID = parsedSearchTeacherID
	}

	if len(searchSchoolID) >= 2 {
		parsedSearchSchoolID, _ := uuid.Parse(searchSchoolID)
		params.Data.SchoolID = parsedSearchSchoolID
	}

	if len(academicYearID) >= 2 {
		parsedAcademicYearID, _ := uuid.Parse(academicYearID)
		params.Data.AcademicYearID = parsedAcademicYearID
	}

	res, err := h.s.Classroom().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllBySubjectID(c *gin.Context) {
	subjectID := c.Param("subject_id")

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchName := c.DefaultQuery("name", "")
	searchTeacherID := c.DefaultQuery("teacher_id", "")
	searchSchoolID := c.DefaultQuery("school_id", "")
	academicYearID := c.DefaultQuery("academic_year_id", "")

	params := rq.PaginationParams[model.Classroom]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Classroom{
			Name: searchName,
		},
	}

	if len(searchTeacherID) >= 2 {
		parsedSearchTeacherID, _ := uuid.Parse(searchTeacherID)
		params.Data.TeacherID = parsedSearchTeacherID
	}

	if len(searchSchoolID) >= 2 {
		parsedSearchSchoolID, _ := uuid.Parse(searchSchoolID)
		params.Data.SchoolID = parsedSearchSchoolID
	}

	if len(academicYearID) >= 2 {
		parsedAcademicYearID, _ := uuid.Parse(academicYearID)
		params.Data.AcademicYearID = parsedAcademicYearID
	}

	res, err := h.s.Classroom().GetAllClassroomBySubjectID(c, subjectID, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetAllByNotInSubjectID(c *gin.Context) {
	subjectID := c.Param("subject_id")

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchName := c.DefaultQuery("name", "")
	searchTeacherID := c.DefaultQuery("teacher_id", "")
	searchSchoolID := c.DefaultQuery("school_id", "")
	academicYearID := c.DefaultQuery("academic_year_id", "")

	params := rq.PaginationParams[model.Classroom]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Classroom{
			Name: searchName,
		},
	}

	if len(searchTeacherID) >= 2 {
		parsedSearchTeacherID, _ := uuid.Parse(searchTeacherID)
		params.Data.TeacherID = parsedSearchTeacherID
	}

	if len(searchSchoolID) >= 2 {
		parsedSearchSchoolID, _ := uuid.Parse(searchSchoolID)
		params.Data.SchoolID = parsedSearchSchoolID
	}

	if len(academicYearID) >= 2 {
		parsedAcademicYearID, _ := uuid.Parse(academicYearID)
		params.Data.AcademicYearID = parsedAcademicYearID
	}

	res, err := h.s.Classroom().GetAllClassroomNotInSubject(c, subjectID, &params)
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

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchName := c.DefaultQuery("name", "")
	searchTeacherID := c.DefaultQuery("teacher_id", "")
	searchSchoolID := c.DefaultQuery("school_id", "")
	academicYearID := c.DefaultQuery("academic_year_id", "")

	params := rq.PaginationParams[model.Classroom]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Classroom{
			Name: searchName,
		},
	}

	if len(searchTeacherID) >= 2 {
		parsedSearchTeacherID, _ := uuid.Parse(searchTeacherID)
		params.Data.TeacherID = parsedSearchTeacherID
	}

	if len(searchSchoolID) >= 2 {
		parsedSearchSchoolID, _ := uuid.Parse(searchSchoolID)
		params.Data.SchoolID = parsedSearchSchoolID
	}

	if len(academicYearID) >= 2 {
		parsedAcademicYearID, _ := uuid.Parse(academicYearID)
		params.Data.AcademicYearID = parsedAcademicYearID
	}

	res, err := h.s.Classroom().GetAllClassroomByTeacherID(c, gotAccount.Teacher.ID.String(), &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetDetailByID(c *gin.Context) {
	id := c.Param("id")
	res, err := h.s.Classroom().GetDetailByID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) CreateOne(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	var request rq.ClassroomRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Classroom().CreateOne(c, &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateMass(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	schoolID := c.DefaultQuery("school_id", "")
	academicYearID := c.DefaultQuery("academic_year_id", "")

	var csvfile rq.CSVFileUploadRequest
	if err := c.ShouldBind(&csvfile); err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, &errmsg.ErrRequestFileInvalid{Info: err.Error()})
		return
	}

	if csvfile.CSVFile == nil {
		util.SaveResponseBody(c, &errmsg.ErrRequestFileInvalid{Info: "csvFile is empty"}, nil)
		rs.ErrorResponse(c, &errmsg.ErrRequestFileInvalid{Info: "csvFile is empty"})
		return
	}

	res, err := h.s.Classroom().CreateMass(c, academicYearID, schoolID, csvfile.CSVFile)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, res, http.StatusCreated)
}

func (h *impHandler) EditOne(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	id := c.Param("id")

	var request rq.ClassroomRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	request.ID = &id

	err = h.s.Classroom().EditOne(c, &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) AssignSubjectsToClassroom(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	classroomID := c.Param("id")

	var request rq.IDsRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Classroom().AssignSubjectsToClassroom(c, classroomID, request.IDs)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) RemoveSubjectsFromClassroom(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	classroomID := c.Param("id")

	var request rq.IDsRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Classroom().RemoveSubjectsFromClassroom(c, classroomID, request.IDs)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}
