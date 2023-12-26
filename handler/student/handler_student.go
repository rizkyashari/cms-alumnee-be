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

	EditOwnData(c *gin.Context)
	EditOwnFamilyData(c *gin.Context)
	EditOwnAddressData(c *gin.Context)

	GetStudentCount(c *gin.Context)
	GetStudentPaidBillCount(c *gin.Context)
	GetStudentUnpaidBillCount(c *gin.Context)

	CountStudents(c *gin.Context)
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
	school_id := c.DefaultQuery("school_id", "")
	classroom_id := c.DefaultQuery("classroom_id", "")
	subject_id := c.DefaultQuery("subject_id", "")
	gender := c.DefaultQuery("gender", "")

	params := rq.PaginationParams[model.Account]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Account{
			Name: &searchName,
			Student: &model.Student{
				Classroom: model.Classroom{},
			},
		},
	}

	if len(classroom_id) >= 2 {
		parsedClassroomID, _ := uuid.Parse(classroom_id)
		params.Data.Student.ClassroomID = &parsedClassroomID
	}

	if len(subject_id) >= 2 {
		parsedSubjectID, _ := uuid.Parse(subject_id)
		params.Data.Student.Classroom.RelationClassroomSubjects = []model.RelationClassroomSubject{{
			SubjectID: parsedSubjectID,
		}}
	}

	if len(school_id) >= 2 {
		parsedSchoolID, _ := uuid.Parse(school_id)
		params.Data.Student.Classroom.SchoolID = parsedSchoolID
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
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
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
	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
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
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	var request rq.StudentRegisterRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().CreateOne(c, &request)
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

	var csvfile rq.CSVFileUploadRequest
	if err := c.ShouldBind(&csvfile); err != nil {
		util.SaveResponseBody(c, &errmsg.ErrRequestFileInvalid{Info: err.Error()}, nil)
		rs.ErrorResponse(c, &errmsg.ErrRequestFileInvalid{Info: err.Error()})
		return
	}

	if csvfile.CSVFile == nil {
		util.SaveResponseBody(c, &errmsg.ErrRequestFileInvalid{Info: "csvfile is empty"}, nil)
		rs.ErrorResponse(c, &errmsg.ErrRequestFileInvalid{Info: "csvfile is empty"})
		return
	}

	res, err := h.s.Student().CreateMass(c, csvfile.CSVFile)
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
	studentID := c.Param("student_id")

	var request rq.StudentUpdateRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().EditOne(c, studentID, &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditFamilyData(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	studentID := c.Param("student_id")

	var request rq.StudentFamilyDataUpdateRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().EditFamilyData(c, studentID, &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditAddressData(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	studentID := c.Param("student_id")

	var request rq.AddressDataUpdateRequest
	err := c.ShouldBindJSON(&request)

	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().EditAddressData(c, studentID, &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditOwnData(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	var request rq.StudentUpdateRequest
	err = c.ShouldBindJSON(&request)

	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().EditOne(c, gotAccount.Student.ID.String(), &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditOwnFamilyData(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	var request rq.StudentFamilyDataUpdateRequest
	err = c.ShouldBindJSON(&request)

	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().EditFamilyData(c, gotAccount.Student.ID.String(), &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) EditOwnAddressData(c *gin.Context) {
	if err := util.SaveRequestBody(c, c.Request.Body); err != nil {
		return
	}

	gotAccount, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	var request rq.AddressDataUpdateRequest
	err = c.ShouldBindJSON(&request)

	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	err = h.s.Student().EditAddressData(c, gotAccount.Student.ID.String(), &request)
	if err != nil {
		util.SaveResponseBody(c, err.Error(), nil)
		rs.ErrorResponse(c, err)
		return
	}

	util.SaveResponseBody(c, nil, "success")
	rs.SuccessResponse(c, nil, http.StatusAccepted)
}

func (h *impHandler) GetStudentCount(c *gin.Context) {

	schoolsResponse, err := h.s.School().GetAll(c, &rq.PaginationParams[model.School]{
		Limit: 10000000,
		Page:  1,
	})

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	// Check if any schools were returned
	if len(schoolsResponse.Data) == 0 {
		rs.ErrorResponse(c, &errmsg.ErrNotFound{})
		return
	}

	// Initialize a slice to store the school data with student counts
	schoolDataList := make([]model.CountStudent, 0)

	for _, school := range schoolsResponse.Data {

		schoolID := school.ID.String()

		students, err := h.getStudentsBySchoolID(c, schoolID)

		if err != nil {
			rs.ErrorResponse(c, err)
			return
		}

		studentCount := len(students.Data)

		schoolDataWithCount := model.CountStudent{
			SchoolID:      school.ID,
			SchoolName:    school.Name,
			StudentAmount: studentCount,
		}

		schoolDataList = append(schoolDataList, schoolDataWithCount)
	}

	rs.SuccessResponse(c, schoolDataList, http.StatusOK)
}

func (h *impHandler) getStudentsBySchoolID(c *gin.Context, schoolID string) (*rs.PaginationResponse[any, rs.AccountResponse], error) {
	parsedSchoolID, err := uuid.Parse(schoolID)
	if err != nil {
		return nil, err
	}

	_, _, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return nil, err
	}

	params := rq.PaginationParams[model.Account]{
		Limit:     10000000,
		Page:      1,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Account{
			Student: &model.Student{
				Classroom: model.Classroom{
					SchoolID: parsedSchoolID,
				},
			},
		},
	}

	response, err := h.s.Student().GetAll(c, &params)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (h *impHandler) GetStudentPaidBillCount(c *gin.Context) {
	schoolsResponse, err := h.s.School().GetAll(c, &rq.PaginationParams[model.School]{
		Limit: 10000000,
		Page:  1,
	})

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	if len(schoolsResponse.Data) == 0 {
		rs.ErrorResponse(c, &errmsg.ErrNotFound{})
		return
	}

	schoolDataList := make([]model.CountPaidBill, 0)

	for _, school := range schoolsResponse.Data {
		schoolID := school.ID.String()

		students, err := h.getStudentsBySchoolID(c, schoolID)
		if err != nil {
			rs.ErrorResponse(c, err)
			return
		}

		// Create a map to store the count of paid bills for each student
		paidBillCounts := make(map[uuid.UUID]int)

		for _, student := range students.Data {
			searchAccountID := student.ID
			_, _, sortBy, sortOrder, err := util.ParseQuery(c)
			if err != nil {
				rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
				return
			}

			searchDescription := c.DefaultQuery("description", "")

			billsResponse, err := h.s.Midtrans().GetAllBills(c, &rq.PaginationParams[model.Bill]{
				Limit:     10000000,
				Page:      1,
				SortBy:    sortBy,
				SortOrder: sortOrder,
				Data: model.Bill{
					Description: &searchDescription,
					AccountID:   searchAccountID,
				},
			})

			if err != nil {
				rs.ErrorResponse(c, err)
				return
			}

			// Count the paid bills for the current student
			paidBillCount := 0
			for _, billResponse := range billsResponse.Data {
				if billResponse.RemainingAmount == 0 {
					paidBillCount++
				}
			}

			paidBillCounts[student.ID] = paidBillCount
		}

		// Calculate the total paid bills for the school
		totalPaidBills := 0
		for _, count := range paidBillCounts {
			totalPaidBills += count
		}

		schoolDataWithCount := model.CountPaidBill{
			SchoolID:       school.ID,
			SchoolName:     school.Name,
			PaidBillAmount: totalPaidBills,
		}

		schoolDataList = append(schoolDataList, schoolDataWithCount)
	}

	rs.SuccessResponse(c, schoolDataList, http.StatusOK)
}

func (h *impHandler) GetStudentUnpaidBillCount(c *gin.Context) {
	schoolsResponse, err := h.s.School().GetAll(c, &rq.PaginationParams[model.School]{
		Limit: 10000000,
		Page:  1,
	})

	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	if len(schoolsResponse.Data) == 0 {
		rs.ErrorResponse(c, &errmsg.ErrNotFound{})
		return
	}

	schoolDataList := make([]model.CountUnpaidBill, 0)

	for _, school := range schoolsResponse.Data {
		schoolID := school.ID.String()

		students, err := h.getStudentsBySchoolID(c, schoolID)
		if err != nil {
			rs.ErrorResponse(c, err)
			return
		}

		// Create a map to store the count of paid bills for each student
		paidBillCounts := make(map[uuid.UUID]int)

		for _, student := range students.Data {
			searchAccountID := student.ID
			_, _, sortBy, sortOrder, err := util.ParseQuery(c)
			if err != nil {
				rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
				return
			}

			searchDescription := c.DefaultQuery("description", "")

			billsResponse, err := h.s.Midtrans().GetAllBills(c, &rq.PaginationParams[model.Bill]{
				Limit:     10000000,
				Page:      1,
				SortBy:    sortBy,
				SortOrder: sortOrder,
				Data: model.Bill{
					Description: &searchDescription,
					AccountID:   searchAccountID,
				},
			})

			if err != nil {
				rs.ErrorResponse(c, err)
				return
			}

			// Count the paid bills for the current student
			paidBillCount := 0
			for _, billResponse := range billsResponse.Data {
				if billResponse.RemainingAmount != 0 {
					paidBillCount++
				}
			}

			paidBillCounts[student.ID] = paidBillCount
		}

		// Calculate the total paid bills for the school
		totalPaidBills := 0
		for _, count := range paidBillCounts {
			totalPaidBills += count
		}

		schoolDataWithCount := model.CountUnpaidBill{
			SchoolID:         school.ID,
			SchoolName:       school.Name,
			UnpaidBillAmount: totalPaidBills,
		}

		schoolDataList = append(schoolDataList, schoolDataWithCount)
	}

	rs.SuccessResponse(c, schoolDataList, http.StatusOK)
}

func (h *impHandler) CountStudents(c *gin.Context) {
	searchName := c.DefaultQuery("name", "")
	gender := c.DefaultQuery("gender", "")

	_, _, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	params := rq.PaginationParams[model.Account]{
		Limit:     10000000,
		Page:      1,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data: model.Account{
			Name: &searchName,
			Teacher: &model.Teacher{
				TeacherData: model.TeacherData{},
			},
		},
	}

	if len(gender) >= 1 {
		genderInt, err := strconv.Atoi(gender)
		if err == nil && util.IsValidConstant(genderInt, constants.GenderMap) {
			params.Data.Teacher.TeacherData.Gender = &genderInt
		}
	}

	// Retrieve the list of teachers based on search criteria
	students, err := h.s.Student().GetAll(c, &params)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	// Count the number of teachers based on the result
	studentCount := len(students.Data)

	// Create a response structure to return the teacher count
	response := model.CountTotalStudent{
		TotalStudents: studentCount,
	}

	rs.SuccessResponse(c, response, http.StatusOK)
}
