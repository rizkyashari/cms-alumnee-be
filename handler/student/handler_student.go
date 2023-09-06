package handler_student

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

type StudentHandler interface {
	GetAll(c *gin.Context)
	GetDetailByAccountID(c *gin.Context)
	GetStudentDataByStudentID(c *gin.Context)
	GetOwnDetail(c *gin.Context)
	GetOwnStudentData(c *gin.Context)

	CreateOne(c *gin.Context)
	CreateMass(c *gin.Context)

	EditOne(c *gin.Context)
	EditFamilyData(c *gin.Context)
	EditAddressData(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) StudentHandler {
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
	classroom_id := c.DefaultQuery("classroom_id", "")
	gender := c.DefaultQuery("gender", "")

	params := rq.PaginationParams[model.Account]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Account{
			Name: &searchName,
		},
	}

	if len(classroom_id) >= 2 {
		parsedClassroomID, _ := uuid.Parse(classroom_id)
		params.Data.Student.ClassroomID = &parsedClassroomID
	}

	if len(gender) >= 1 {
		genderInt, err := strconv.Atoi(gender)
		if err == nil && util.IsValidConstant(genderInt, constants.GenderMap) {
			params.Data.Student.StudentData.Gender = &genderInt
		}
	}

	res, err := h.s.Student().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetDetailByAccountID(c *gin.Context) {
	id := c.Param("account_id")
	res, err := h.s.Student().GetDetailByAccountID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetStudentDataByStudentID(c *gin.Context) {
	id := c.Param("student_id")
	res, err := h.s.Student().GetStudentDataByStudentID(c, id)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetOwnDetail(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	res, err := h.s.Student().GetDetailByAccountID(c, gotAccount.ID.String())
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) GetOwnStudentData(c *gin.Context) {
	gotAccount, err := util.GetAccountContext(c)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	res, err := h.s.Student().GetStudentDataByStudentID(c, gotAccount.Student.ID.String())
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusOK)
}

func (h *impHandler) CreateOne(c *gin.Context) {
	var request rq.StudentRegisterRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().CreateOne(c, &request)
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

	res, err := h.s.Student().CreateMass(c, csvfile.CSVFile)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, res, http.StatusCreated)
}

func (h *impHandler) EditOne(c *gin.Context) {
	studentID := c.Param("student_id")

	var request rq.StudentUpdateRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().EditOne(c, studentID, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditFamilyData(c *gin.Context) {
	studentID := c.Param("student_id")

	var request rq.StudentFamilyDataUpdateRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().EditFamilyData(c, studentID, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditAddressData(c *gin.Context) {
	studentID := c.Param("student_id")

	var request rq.AddressDataUpdateRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().EditAddressData(c, studentID, &request)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, nil, http.StatusAccepted)
}
