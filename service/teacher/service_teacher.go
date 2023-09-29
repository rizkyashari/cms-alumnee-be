package service_teacher

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/auth"
	"github.com/fadhln/lms-be/util/errmsg"
	serviceutil "github.com/fadhln/lms-be/util/service_util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type TeacherService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.Account]) (*rs.PaginationResponse[any, rs.AccountResponse], error)
	GetAllClassroomSubject(c context.Context, teacherID string, academicYearID string, params *rq.PaginationParams[model.RelationClassroomSubject]) (*rs.PaginationResponse[any, rs.ClassroomSubject], error)
	GetDetailByAccountID(c context.Context, id string) (*rs.TeacherResponse, error)
	GetTeacherDataByTeacherID(c context.Context, id string) (*rs.TeacherDataResponse, error)

	CreateOne(c context.Context, newTeacher *rq.TeacherRegisterRequest) error
	CreateOneWithDetail(c context.Context, newTeacher *rq.TeacherRegisterWithDetailRequest) error
	CreateMass(c *gin.Context, school_id string, requestFile *multipart.FileHeader) (*rs.MassCreateResponse, error)

	EditOne(c context.Context, teacherID string, newTeacher *rq.TeacherUpdateRequest) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) TeacherService {
	return &impService{
		repo: r,
	}
}

func getTeacherCSV(school_id uuid.UUID, requestFile *multipart.FileHeader) (*[]rq.TeacherRegisterWithDetailRequest, error) {
	if requestFile == nil {
		return nil, errmsg.ErrRequestFileInvalid
	}

	file, err := requestFile.Open()
	if err != nil {
		return nil, errmsg.ErrRequestFileInvalid
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, errmsg.ErrRequestFileInvalid
	}

	var teachers []rq.TeacherRegisterWithDetailRequest
	for idx, rec := range records {
		if idx == 0 {
			continue
		}

		schoolIDString := school_id.String()

		teacher := rq.TeacherRegisterWithDetailRequest{
			Account: rq.BaseRegisterRequest{
				Email:    rec[0],
				Password: rec[1],
				Name:     &rec[2],
			},
			Data: rq.TeacherUpdateRequest{
				SchoolID: &schoolIDString,
			},
		}

		if len(rec[3]) > 0 {
			if rec[3] == "Male" {
				gender := constants.GENDER_MALE
				teacher.Data.Gender = &gender
			} else if rec[3] == "Female" {
				gender := constants.GENDER_FEMALE
				teacher.Data.Gender = &gender
			}
		}

		if len(rec[4]) > 0 {
			teacher.Data.NIK = &rec[4]
		}

		if len(rec[5]) > 0 {
			teacher.Data.NUPTK = &rec[5]
		}

		if len(rec[6]) > 0 {
			teacher.Data.NIP = &rec[6]
		}

		if len(rec[7]) > 0 {
			if rec[7] == "PNS" {
				status := constants.TEACHER_STATUS_PNS
				teacher.Data.EmploymentStatus = &status
			} else if rec[7] == "NON" {
				status := constants.TEACHER_STATUS_NON_PNS
				teacher.Data.EmploymentStatus = &status
			}
		}

		if len(rec[8]) > 0 {
			teacher.Data.BirthPlace = &rec[8]
		}

		if len(rec[9]) > 0 {
			teacher.Data.BirthDate = &rec[9]
		}

		if len(rec[10]) > 0 {
			teacher.Data.PhoneNumber = &rec[10]
		}

		teachers = append(teachers, teacher)
	}

	return &teachers, nil
}

func getTeacherDataFromRequest(body *rq.TeacherUpdateRequest) (*model.TeacherData, bool, error) {
	isUpdateTeacherData := false
	var newTeacherData model.TeacherData

	if body.Gender != nil {
		if !(util.IsValidConstant(*body.Gender, constants.GenderMap)) {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Gender"}
		}
		newTeacherData.Gender = body.Gender
		isUpdateTeacherData = true
	}

	if body.NIK != nil {
		if len(*body.NIK) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "NIK"}
		}
		newTeacherData.NIK = body.NIK
		isUpdateTeacherData = true
	}

	if body.NUPTK != nil {
		if len(*body.NUPTK) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "NUPTK"}
		}
		newTeacherData.NUPTK = body.NUPTK
		isUpdateTeacherData = true
	}

	if body.NIP != nil {
		if len(*body.NIP) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "NIP"}
		}
		newTeacherData.NIP = body.NIP
		isUpdateTeacherData = true
	}

	if body.EmploymentStatus != nil {
		if !(util.IsValidConstant(*body.EmploymentStatus, constants.TeacherStatusMap)) {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Employment Status"}
		}
		newTeacherData.EmploymentStatus = body.EmploymentStatus
		isUpdateTeacherData = true
	}

	if body.BirthPlace != nil {
		if len(*body.BirthPlace) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Birth Place"}
		}
		newTeacherData.BirthPlace = body.BirthPlace
		isUpdateTeacherData = true
	}

	if body.BirthDate != nil {
		if len(*body.BirthDate) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Birth Date"}
		}

		parsedBirthDate, err := time.Parse(constants.DateLayout, *body.BirthDate)
		if err != nil {
			return nil, false, &errmsg.ErrInternal{Err: err}
		}
		newTeacherData.BirthDate = &parsedBirthDate
		isUpdateTeacherData = true
	}

	if body.PhoneNumber != nil {
		if len(*body.PhoneNumber) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Phone Number"}
		}
		newTeacherData.PhoneNumber = body.PhoneNumber
		isUpdateTeacherData = true
	}

	return &newTeacherData, isUpdateTeacherData, nil
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[model.Account]) (*rs.PaginationResponse[any, rs.AccountResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotAccounts, maxPage, rowCount, err := s.repo.Account().GetAllTeacher(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.AccountResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotAccounts == nil {
		return &rs.PaginationResponse[any, rs.AccountResponse]{
			Data: []rs.AccountResponse{},
		}, nil
	}

	if len(*gotAccounts) < 1 {
		return &rs.PaginationResponse[any, rs.AccountResponse]{
			Data: []rs.AccountResponse{},
		}, nil
	}

	var response []rs.AccountResponse

	for _, account := range *gotAccounts {
		var tempAccount rs.AccountResponse
		err = copier.Copy(&tempAccount, account)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		var tempTeacher rs.TeacherResponse
		err = copier.Copy(&tempTeacher, account.Teacher)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}
		tempAccount.Teacher = &tempTeacher

		if tempAccount.Teacher != nil {
			var tempTeacherData rs.TeacherDataResponse
			err = copier.Copy(&tempTeacherData, account.Teacher.TeacherData)
			if err != nil {
				return nil, &errmsg.ErrInternal{Err: err}
			}
			if tempTeacherData.ID != uuid.Nil {
				tempAccount.Teacher.TeacherData = &tempTeacherData
			}
		}

		response = append(response, tempAccount)
	}

	res := rs.PaginationResponse[any, rs.AccountResponse]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) GetAllClassroomSubject(c context.Context, teacherID string, academicYearID string, params *rq.PaginationParams[model.RelationClassroomSubject]) (*rs.PaginationResponse[any, rs.ClassroomSubject], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", &teacherID)
	if err != nil {
		return nil, err
	}

	newParam := *params
	newParam.Data.Subject.TeacherID = *parsedTeacherID

	parsedAcademicYearID, err := serviceutil.GetUUIDFromStringWithValidation("Academic Year ID", &academicYearID)
	if err == nil {
		newParam.Data.Classroom.AcademicYearID = *parsedAcademicYearID
	}

	gotRelation, maxPage, rowCount, err := s.repo.RelationClassroomSubject().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.ClassroomSubject]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotRelation == nil {
		return &rs.PaginationResponse[any, rs.ClassroomSubject]{
			Data: []rs.ClassroomSubject{},
		}, nil
	}

	if len(*gotRelation) < 1 {
		return &rs.PaginationResponse[any, rs.ClassroomSubject]{
			Data: []rs.ClassroomSubject{},
		}, nil
	}

	var response []rs.ClassroomSubject

	for _, relation := range *gotRelation {
		var tempRelation rs.ClassroomSubject
		err = copier.Copy(&tempRelation, relation)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		var tempClassroom rs.ClassroomResponse
		err = copier.Copy(&tempClassroom, relation.Classroom)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}
		tempRelation.Classroom = tempClassroom

		var tempSubject rs.SubjectResponse
		err = copier.Copy(&tempSubject, relation.Subject)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}
		tempRelation.Subject = tempSubject

		response = append(response, tempRelation)
	}

	res := rs.PaginationResponse[any, rs.ClassroomSubject]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil

}

func (s *impService) GetDetailByAccountID(c context.Context, id string) (*rs.TeacherResponse, error) {
	parsedAccountID, err := serviceutil.GetUUIDFromStringWithValidation("Account ID", &id)
	if err != nil {
		return nil, err
	}

	gotTeacher, err := s.repo.Teacher().GetDetailByAccountID(*parsedAccountID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.TeacherResponse
	err = copier.Copy(&res, gotTeacher)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}

func (s *impService) GetTeacherDataByTeacherID(c context.Context, id string) (*rs.TeacherDataResponse, error) {
	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", &id)
	if err != nil {
		return nil, err
	}

	gotTeacherData, err := s.repo.TeacherData().GetTeacherDataByTeacherID(*parsedTeacherID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.TeacherDataResponse
	err = copier.Copy(&res, gotTeacherData)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}

func (s *impService) CreateOne(c context.Context, body *rq.TeacherRegisterRequest) error {
	if len(body.Account.Email) <= 0 {
		return &errmsg.ErrIsEmpty{FieldName: "Email"}
	}

	parsedSchoolID, err := serviceutil.GetUUIDFromStringWithValidation("School ID", &body.Data.SchoolID)
	if err != nil {
		return err
	}

	email := strings.ToLower(body.Account.Email)
	if !(util.IsEmailValid(email)) {
		return &errmsg.ErrFieldIsWrong{FieldName: "Email"}
	}

	var password string
	if len(body.Account.Password) <= 0 {
		password = os.Getenv("DEFAULT_PASSWORD")
	} else {
		password = body.Account.Password
	}

	hashedPassword, err := auth.HashAndSalt(password)

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	newID := uuid.New()
	newTeacherID := uuid.New()

	newAccount := model.Account{
		Base:        model.Base{ID: newID},
		Name:        body.Account.Name,
		Email:       email,
		Password:    hashedPassword,
		AccountType: constants.ACCOUNT_TEACHER,
	}

	newTeacher := model.Teacher{
		Base:        model.Base{ID: newTeacherID},
		AccountID:   newID,
		SchoolID:    *parsedSchoolID,
		TeacherData: model.TeacherData{TeacherID: newTeacherID},
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Account().CreateOne(tx, &newAccount); err != nil {
			return err
		}

		if err := s.repo.Teacher().CreateOne(tx, &newTeacher); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) CreateOneWithDetail(c context.Context, body *rq.TeacherRegisterWithDetailRequest) error {
	if len(body.Account.Email) <= 0 {
		return &errmsg.ErrIsEmpty{FieldName: "Email"}
	}

	if body.Data.SchoolID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "School ID"}
	}

	parsedSchoolID, err := uuid.Parse(*body.Data.SchoolID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "School ID"}
	}

	email := strings.ToLower(body.Account.Email)
	if !(util.IsEmailValid(email)) {
		return &errmsg.ErrFieldIsWrong{FieldName: "Email"}
	}

	var password string
	if len(body.Account.Password) <= 0 {
		password = os.Getenv("DEFAULT_PASSWORD")
	} else {
		password = body.Account.Password
	}

	hashedPassword, err := auth.HashAndSalt(password)

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	newAccountID := uuid.New()
	newTeacherID := uuid.New()

	newAccount := model.Account{
		Base:        model.Base{ID: newAccountID},
		Name:        body.Account.Name,
		Email:       email,
		Password:    hashedPassword,
		AccountType: constants.ACCOUNT_TEACHER,
	}

	newTeacher := model.Teacher{
		Base:      model.Base{ID: newTeacherID},
		AccountID: newAccountID,
		SchoolID:  parsedSchoolID,
	}

	newTeacherData, _, err := getTeacherDataFromRequest(&body.Data)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	newTeacherData.TeacherID = newTeacherID

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Account().CreateOne(tx, &newAccount); err != nil {
			return err
		}

		if err := s.repo.Teacher().CreateOne(tx, &newTeacher); err != nil {
			return err
		}

		if err := s.repo.TeacherData().CreateOne(tx, newTeacherData); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) processMassCreate(c context.Context, requests *[]rq.TeacherRegisterWithDetailRequest, reportFileName string, newMassCreate *model.MassCreate) {
	if requests == nil {
		newMassCreate.Status = constants.MASS_CREATE_STATUS_FAILED

		s.repo.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.MassCreate().UpdateOne(tx, newMassCreate.ID, newMassCreate); err != nil {
				return err
			}

			return nil
		})

		return
	}

	if len(*requests) >= 350 {
		newMassCreate.Status = constants.MASS_CREATE_STATUS_FAILED

		s.repo.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.MassCreate().UpdateOne(tx, newMassCreate.ID, newMassCreate); err != nil {
				return err
			}

			return nil
		})

		return
	}

	successCount := 0
	var errMsgs []model.ErrorMsg
	for idx, newTeacher := range *requests {
		err := s.CreateOneWithDetail(c, &newTeacher)
		if err != nil {
			errMsgs = append(errMsgs, model.ErrorMsg{
				Row:     idx + 1,
				Message: err.Error()})
			continue
		}
		successCount++
	}

	errMsgsStr, _ := json.Marshal(errMsgs)
	file, _ := os.Create("./file/output/" + reportFileName)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write(model.GetErrorMsgHeader())
	for _, errMsg := range errMsgs {
		writer.Write(model.GetErrorMsgRow(&errMsg))
	}

	newMassCreate.Status = constants.MASS_CREATE_STATUS_DONE
	newMassCreate.ReportFileUrl = reportFileName
	newMassCreate.ErrorCount = len(errMsgs)
	newMassCreate.SuccessCount = successCount
	newMassCreate.ErrorMessages = string(errMsgsStr)

	s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.MassCreate().UpdateOne(tx, newMassCreate.ID, newMassCreate); err != nil {
			return err
		}

		return nil
	})
}

func (s *impService) CreateMass(c *gin.Context, school_id string, requestFile *multipart.FileHeader) (*rs.MassCreateResponse, error) {
	parsedSchoolID, err := uuid.Parse(school_id)
	if err != nil {
		return nil, errmsg.ErrRequestParamsInvalid
	}

	if parsedSchoolID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "School ID"}
	}

	newTeachers, err := getTeacherCSV(parsedSchoolID, requestFile)
	if err != nil {
		return nil, errmsg.ErrRequestFileInvalid
	}

	nowStr := time.Now().Format(constants.FilenameTimeFormat)
	requestFileName := fmt.Sprintf("teacher-request-%s-%s.csv", nowStr, uuid.New().String())
	reportFileName := fmt.Sprintf("teacher-report-%s-%s.csv", nowStr, uuid.New().String())

	err = c.SaveUploadedFile(requestFile, "./file/input/"+requestFileName)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	newMassCreateID := uuid.New()
	var newMassCreate = model.MassCreate{
		Base:           model.Base{ID: newMassCreateID},
		Status:         constants.MASS_CREATE_STATUS_PROCESSING,
		Destination:    constants.DEST_TEACHER,
		RequestFileUrl: requestFileName,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.MassCreate().CreateOne(tx, &newMassCreate); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	go s.processMassCreate(c, newTeachers, reportFileName, &newMassCreate)

	var res rs.MassCreateResponse
	err = copier.Copy(&res, newMassCreate)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil

}

func (s *impService) EditOne(c context.Context, teacherID string, body *rq.TeacherUpdateRequest) error {
	parsedTeacherID, err := uuid.Parse(teacherID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Teacher ID"}
	}

	if parsedTeacherID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	if body.SchoolID != nil {
		parsedSchoolID, err := uuid.Parse(*body.SchoolID)
		if err != nil {
			return &errmsg.ErrFieldIsWrong{FieldName: "School ID"}
		}

		if parsedSchoolID == uuid.Nil {
			return &errmsg.ErrIsEmpty{FieldName: "School ID"}
		}

		newTeacher := model.Teacher{
			Base: model.Base{
				ID: parsedTeacherID,
			},
			SchoolID: parsedSchoolID,
		}

		err = s.repo.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.Teacher().UpdateOne(tx, parsedTeacherID, &newTeacher); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}
	}

	newTeacherData, isUpdateTeacherData, err := getTeacherDataFromRequest(body)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	if isUpdateTeacherData {
		err = s.repo.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.TeacherData().UpdateOne(tx, parsedTeacherID, newTeacherData); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}
	}

	return nil
}
