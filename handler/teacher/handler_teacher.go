package handler_teacher

import (
	"net/http"
	"strconv"

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

type TeacherHandler interface {
	GetAll(c *gin.Context)
	GetOwnAllClassroomSubject(c *gin.Context)
	GetAllClassroomStudents(c *gin.Context)
	GetDetailByAccountID(c *gin.Context)
	GetTeacherDataByTeacherID(c *gin.Context)
	GetOwnDetail(c *gin.Context)
	GetOwnTeacherData(c *gin.Context)

	CreateOne(c *gin.Context)
	CreateMass(c *gin.Context)

	EditOne(c *gin.Context)

	EditOwnData(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) TeacherHandler {
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
	schoolId := c.DefaultQuery("school_id", "")
	gender := c.DefaultQuery("gender", "")
	employmentStatus := c.DefaultQuery("employment_status", "")

	params := rq.PaginationParams[model.Account]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Account{
			Name: &searchName,
			Teacher: &model.Teacher{
				TeacherData: model.TeacherData{},
			},
		},
	}

	if len(schoolId) >= 2 {
		parsedSchoolID, _ := uuid.Parse(schoolId)
		params.Data.Teacher.SchoolID = parsedSchoolID
	}

	if len(gender) >= 1 {
		genderInt, err := strconv.Atoi(gender)
		if err == nil && util.IsValidConstant(genderInt, constants.GenderMap) {
			params.Data.Teacher.TeacherData.Gender = &genderInt
		}
	}

	if len(employmentStatus) >= 1 {
		statusInt, err := strconv.Atoi(employmentStatus)
		if err == nil && util.IsValidConstant(statusInt, constants.TeacherStatusMap) {
			params.Data.Teacher.TeacherData.EmploymentStatus = &statusInt
		}
	}

	res, err := h.s.Teacher().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetOwnAllClassroomSubject(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	academicYearID := c.DefaultQuery("academic_year_id", "")

	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	params := rq.PaginationParams[model.RelationClassroomSubject]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data:      model.RelationClassroomSubject{},
	}

	res, err := h.s.Teacher().GetAllClassroomSubject(c, gotAccount.ID.String(), academicYearID, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetDetailByAccountID(c *gin.Context) {
	id := c.Param("account_id")
	res, err := h.s.Teacher().GetDetailByAccountID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetTeacherDataByTeacherID(c *gin.Context) {
	id := c.Param("teacher_id")
	res, err := h.s.Teacher().GetTeacherDataByTeacherID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetOwnDetail(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	res, err := h.s.Teacher().GetDetailByAccountID(c, gotAccount.ID.String())
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetOwnTeacherData(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	res, err := h.s.Teacher().GetTeacherDataByTeacherID(c, gotAccount.Teacher.ID.String())
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) CreateOne(c *gin.Context) {
	var request rq.TeacherRegisterRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Teacher().CreateOne(c, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusCreated)
}

func (h *impHandler) CreateMass(c *gin.Context) {
	schoolID := c.DefaultQuery("school_id", "")

	var csvfile rq.CSVFileUploadRequest
	if err := c.ShouldBind(&csvfile); err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestFileInvalid)
		return
	}

	if csvfile.CSVFile == nil {
		rs.ErrorResponse(c, errmsg.ErrRequestFileInvalid)
		return
	}

	res, err := h.s.Teacher().CreateMass(c, schoolID, csvfile.CSVFile)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusCreated)
}

func (h *impHandler) EditOne(c *gin.Context) {
	teacherID := c.Param("teacher_id")

	var request rq.TeacherUpdateRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Teacher().EditOne(c, teacherID, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditOwnData(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	var request rq.TeacherUpdateRequest
	err = c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Teacher().EditOne(c, gotAccount.Teacher.ID.String(), &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) GetAllClassroomStudents(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_TEACHER)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	classrooms, students, err := h.s.Teacher().GetAllClassroomStudents(c, gotAccount.Teacher.ID.String())
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	response := rs.ClassroomStudentsResponse{
		Classrooms: classrooms,
		Students:   students,
	}

	rs.SuccessResponse(c, response, http.StatusOK)
}
