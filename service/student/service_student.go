package service_student

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
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type StudentService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.Account]) (*rs.PaginationResponse[any, rs.AccountResponse], error)
	GetDetailByAccountID(c context.Context, id string) (*rs.StudentResponse, error)
	GetStudentDataByStudentID(c context.Context, id string) (*rs.StudentDataResponse, error)

	CreateOne(c context.Context, newStudent *rq.StudentRegisterRequest) error
	CreateOneWithDetail(c context.Context, newStudent *rq.StudentRegisterWithDetailRequest) error
	CreateMass(c *gin.Context, requestFile *multipart.FileHeader) (*rs.MassCreateResponse, error)

	EditOne(c context.Context, studentID string, newStudent *rq.StudentUpdateRequest) error
	EditFamilyData(c context.Context, studentID string, newStudentFamilyData *rq.StudentFamilyDataUpdateRequest) error
	EditAddressData(c context.Context, studentID string, newAddressData *rq.AddressDataUpdateRequest) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) StudentService {
	return &impService{
		repo: r,
	}
}

func getStudentCSV(requestFile *multipart.FileHeader) (*[]rq.StudentRegisterWithDetailRequest, error) {
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

	var students []rq.StudentRegisterWithDetailRequest
	for idx, rec := range records {
		if idx == 0 {
			continue
		}

		student := rq.StudentRegisterWithDetailRequest{
			Account: rq.BaseRegisterRequest{
				Email:    rec[0],
				Password: rec[1],
				Name:     &rec[2],
			},
		}

		if len(rec[3]) > 0 {
			if rec[3] == "Male" {
				gender := constants.GENDER_MALE
				student.Data.Gender = &gender
			} else if rec[3] == "Female" {
				gender := constants.GENDER_FEMALE
				student.Data.Gender = &gender
			}
		}

		if len(rec[4]) > 0 {
			student.Data.ClassroomCode = &rec[4]
		}

		if len(rec[5]) > 0 {
			student.Data.BirthPlace = &rec[5]
		}

		if len(rec[6]) > 0 {
			student.Data.BirthDate = &rec[6]
		}

		if len(rec[7]) > 0 {
			student.Data.PhoneNumber = &rec[7]
		}

		if len(rec[8]) > 0 {
			student.Data.Religion = &rec[8]
		}

		if len(rec[9]) > 0 {
			if rec[9] == "WNI" {
				status := constants.NATIONALITY_INDONESIAN
				student.Data.Nationality = &status
			} else if rec[9] == "WNA" {
				status := constants.NATIONALITY_FOREIGN
				student.Data.Nationality = &status
			}
		}

		if len(rec[10]) > 0 {
			student.Data.EthnicGroup = &rec[9]
		}

		students = append(students, student)
	}

	return &students, nil
}

func getStudentDataFromRequest(body *rq.StudentUpdateRequest) (*model.StudentData, bool, error) {
	isUpdate := false
	var newStudentData model.StudentData

	if body.Gender != nil {
		if !(util.IsValidConstant(*body.Gender, constants.GenderMap)) {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Gender"}
		}
		newStudentData.Gender = body.Gender
		isUpdate = true
	}

	if body.BirthPlace != nil {
		if len(*body.BirthPlace) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Birth Place"}
		}
		newStudentData.BirthPlace = body.BirthPlace
		isUpdate = true
	}

	if body.BirthDate != nil {
		if len(*body.BirthDate) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Birth Date"}
		}

		parsedBirthDate, err := time.Parse(constants.DateLayout, *body.BirthDate)
		if err != nil {
			return nil, false, &errmsg.ErrInternal{Err: err}
		}
		newStudentData.BirthDate = &parsedBirthDate
		isUpdate = true
	}

	if body.PhoneNumber != nil {
		if len(*body.PhoneNumber) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Phone Number"}
		}
		newStudentData.PhoneNumber = body.PhoneNumber
		isUpdate = true
	}

	if body.Religion != nil {
		if len(*body.Religion) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Religion"}
		}
		newStudentData.Religion = body.Religion
		isUpdate = true
	}

	if body.Nationality != nil {
		if !(util.IsValidConstant(*body.Nationality, constants.NationalityMap)) {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Nationality"}
		}
		newStudentData.Nationality = body.Nationality
		isUpdate = true
	}

	if body.EthnicGroup != nil {
		if len(*body.EthnicGroup) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Ethnic Group"}
		}
		newStudentData.EthnicGroup = body.EthnicGroup
		isUpdate = true
	}

	return &newStudentData, isUpdate, nil
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

	gotAccounts, maxPage, err := s.repo.Account().GetAllStudent(params)
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

		var tempStudent rs.StudentResponse
		err = copier.Copy(&tempStudent, account.Student)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		var tempStudentData rs.StudentDataResponse
		err = copier.Copy(&tempStudentData, account.Student.StudentData)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		tempAccount.Student = &tempStudent

		if tempAccount.Student != nil {
			var tempStudentData rs.StudentDataResponse
			err = copier.Copy(&tempStudentData, account.Student.StudentData)
			if err != nil {
				return nil, &errmsg.ErrInternal{Err: err}
			}
			tempAccount.Student.StudentData = &tempStudentData
		}

		response = append(response, tempAccount)
	}

	res := rs.PaginationResponse[any, rs.AccountResponse]{
		MaxPage:         maxPage,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) GetDetailByAccountID(c context.Context, id string) (*rs.StudentResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	gotStudent, err := s.repo.Student().GetDetailByAccountID(parsedID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.StudentResponse
	err = copier.Copy(&res, gotStudent)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}

func (s *impService) GetStudentDataByStudentID(c context.Context, id string) (*rs.StudentDataResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	gotStudentData, err := s.repo.StudentData().GetStudentDataByStudentID(parsedID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.StudentDataResponse
	err = copier.Copy(&res, gotStudentData)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var studentFamilyData rs.StudentFamilyDataResponse
	err = copier.Copy(&studentFamilyData, gotStudentData.StudentFamilyData)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}
	res.FamilyData = studentFamilyData

	var addressData rs.AddressDataResponse
	err = copier.Copy(&addressData, gotStudentData.AddressData)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}
	res.AddressData = addressData

	return &res, nil
}

func (s *impService) CreateOne(c context.Context, body *rq.StudentRegisterRequest) error {
	if len(body.Account.Email) <= 0 {
		return &errmsg.ErrIsEmpty{FieldName: "Email"}
	}

	if body.Data.ClassroomCode != nil {
		gotClassroom, err := s.repo.Classroom().GetDetailByCode(*body.Data.ClassroomCode)
		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}
		classroomID := gotClassroom.ID.String()
		body.Data.ClassroomID = &classroomID
	}

	parsedClassroomID, err := uuid.Parse(*body.Data.ClassroomID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Classroom ID"}
	}

	if parsedClassroomID == uuid.Nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Classroom ID"}
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
	newStudentID := uuid.New()

	newAccount := model.Account{
		Base:        model.Base{ID: newID},
		Name:        body.Account.Name,
		Email:       email,
		Password:    hashedPassword,
		AccountType: constants.ACCOUNT_STUDENT,
	}

	newStudent := model.Student{
		Base:        model.Base{ID: newStudentID},
		AccountID:   newID,
		ClassroomID: &parsedClassroomID,
		StudentData: model.StudentData{StudentID: newStudentID},
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Account().CreateOne(tx, &newAccount); err != nil {
			return err
		}

		if err := s.repo.Student().CreateOne(tx, &newStudent); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) CreateOneWithDetail(c context.Context, body *rq.StudentRegisterWithDetailRequest) error {
	if len(body.Account.Email) <= 0 {
		return &errmsg.ErrIsEmpty{FieldName: "Email"}
	}

	if body.Data.ClassroomCode != nil {
		gotClassroom, err := s.repo.Classroom().GetDetailByCode(*body.Data.ClassroomCode)
		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}
		classroomID := gotClassroom.ID.String()
		body.Data.ClassroomID = &classroomID
	}

	if body.Data.ClassroomID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Classroom ID"}
	}

	parsedClassroomID, err := uuid.Parse(*body.Data.ClassroomID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Classroom ID"}
	}

	if parsedClassroomID == uuid.Nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Classroom ID"}
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
	newStudentID := uuid.New()

	newAccount := model.Account{
		Base:        model.Base{ID: newAccountID},
		Name:        body.Account.Name,
		Email:       email,
		Password:    hashedPassword,
		AccountType: constants.ACCOUNT_STUDENT,
	}

	newStudent := model.Student{
		Base:        model.Base{ID: newStudentID},
		AccountID:   newAccountID,
		ClassroomID: &parsedClassroomID,
		StudentData: model.StudentData{StudentID: newStudentID},
	}

	newStudentData, _, err := getStudentDataFromRequest(&body.Data)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	newStudentData.StudentID = newStudentID

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Account().CreateOne(tx, &newAccount); err != nil {
			return err
		}

		if err := s.repo.Student().CreateOne(tx, &newStudent); err != nil {
			return err
		}

		gotStudentData, err := s.repo.StudentData().GetStudentDataByStudentID(newStudentID)
		if err != nil {
			return err
		}

		newStudentData.ID = gotStudentData.ID

		if err := s.repo.StudentData().UpdateOne(tx, newStudentID, newStudentData); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) processMassCreate(c context.Context, requests *[]rq.StudentRegisterWithDetailRequest, reportFileName string, newMassCreate *model.MassCreate) {
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
	for idx, newStudent := range *requests {
		err := s.CreateOneWithDetail(c, &newStudent)
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

func (s *impService) CreateMass(c *gin.Context, requestFile *multipart.FileHeader) (*rs.MassCreateResponse, error) {

	newStudents, err := getStudentCSV(requestFile)
	if err != nil {
		return nil, errmsg.ErrRequestFileInvalid
	}

	nowStr := time.Now().Format(constants.FilenameTimeFormat)
	requestFileName := fmt.Sprintf("student-request-%s-%s.csv", nowStr, uuid.New().String())
	reportFileName := fmt.Sprintf("student-report-%s-%s.csv", nowStr, uuid.New().String())

	err = c.SaveUploadedFile(requestFile, "./file/input/"+requestFileName)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	newMassCreateID := uuid.New()
	var newMassCreate = model.MassCreate{
		Base:           model.Base{ID: newMassCreateID},
		Status:         constants.MASS_CREATE_STATUS_PROCESSING,
		Destination:    constants.DEST_STUDENT,
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

	go s.processMassCreate(c, newStudents, reportFileName, &newMassCreate)

	var res rs.MassCreateResponse
	err = copier.Copy(&res, newMassCreate)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil

}

func (s *impService) EditOne(c context.Context, studentID string, body *rq.StudentUpdateRequest) error {
	parsedStudentID, err := uuid.Parse(studentID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Teacher ID"}
	}

	if parsedStudentID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	if body.ClassroomID != nil || body.ClassroomCode != nil {
		if body.ClassroomCode != nil {
			gotClassroom, err := s.repo.Classroom().GetDetailByCode(*body.ClassroomCode)
			if err != nil {
				return &errmsg.ErrInternal{Err: err}
			}
			classroomID := gotClassroom.ID.String()
			body.ClassroomID = &classroomID
		}

		parsedClassroomID, err := uuid.Parse(*body.ClassroomID)
		if err != nil {
			return &errmsg.ErrFieldIsWrong{FieldName: "Classroom ID"}
		}

		if parsedClassroomID == uuid.Nil {
			return &errmsg.ErrFieldIsWrong{FieldName: "Classroom ID"}
		}

		newStudent := model.Student{
			Base: model.Base{
				ID: parsedStudentID,
			},
			ClassroomID: &parsedClassroomID,
		}

		err = s.repo.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.Student().UpdateOne(tx, parsedStudentID, &newStudent); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}
	}

	newStudentData, isUpdateData, err := getStudentDataFromRequest(body)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	if isUpdateData {
		err = s.repo.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.StudentData().UpdateOne(tx, parsedStudentID, newStudentData); err != nil {
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

func (s *impService) EditFamilyData(c context.Context, studentID string, body *rq.StudentFamilyDataUpdateRequest) error {
	parsedStudentID, err := uuid.Parse(studentID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Teacher ID"}
	}

	if parsedStudentID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	var newFamilyData model.StudentFamilyData
	err = copier.Copy(&newFamilyData, body)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.StudentData().UpdateOneStudentFamilyData(tx, parsedStudentID, &newFamilyData); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) EditAddressData(c context.Context, studentID string, body *rq.AddressDataUpdateRequest) error {
	parsedStudentID, err := uuid.Parse(studentID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Teacher ID"}
	}

	if parsedStudentID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	var newAddressData model.AddressData
	err = copier.Copy(&newAddressData, body)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.StudentData().UpdateOneAddressData(tx, parsedStudentID, &newAddressData); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
