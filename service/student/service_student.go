package service_student

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strconv"
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
			student.Data.EthnicGroup = &rec[10]
		}

		if len(rec[11]) > 0 {
			student.Data.Nickname = &rec[11]
		}

		if len(rec[12]) > 0 {
			student.Data.OriginSchoolNumber = &rec[12]
		}

		if len(rec[13]) > 0 {
			student.Data.DistanceToSchool = &rec[13]
		}

		if len(rec[14]) > 0 {
			student.Data.TransportationToSchool = &rec[14]
		}

		if len(rec[15]) > 0 {
			student.Data.Hobby = &rec[15]
		}

		if len(rec[16]) > 0 {
			student.Data.Ideal = &rec[16]
		}

		if len(rec[17]) > 0 {
			strPtr := rec[17]
			intValue, _ := strconv.Atoi(strPtr)
			student.Data.FamilyData.ChildNumber = &intValue
		}

		if len(rec[18]) > 0 {
			if rec[18] == "Adopsi" {
				status := constants.STUDENT_CHILD_STATUS_ADOPTED
				student.Data.FamilyData.ChildStatus = &status
			} else if rec[18] == "Kandung" {
				status := constants.STUDENT_CHILD_STATUS_BIOLOGICAL
				student.Data.FamilyData.ChildStatus = &status
			} else if rec[18] == "Tiri" {
				status := constants.STUDENT_CHILD_STATUS_STEP
				student.Data.FamilyData.ChildStatus = &status
			}
		}

		if len(rec[19]) > 0 {
			strPtr := rec[19]
			intValue, _ := strconv.Atoi(strPtr)
			student.Data.FamilyData.SiblingCount = &intValue
		}

		if len(rec[20]) > 0 {
			student.Data.FamilyData.SpokenLanguage = &rec[20]
		}

		if len(rec[21]) > 0 {
			student.Data.AddressData.District = &rec[22]
		}
		if len(rec[22]) > 0 {
			student.Data.AddressData.FullAddress = &rec[23]
		}
		if len(rec[23]) > 0 {
			student.Data.AddressData.HouseNumber = &rec[24]
		}
		if len(rec[24]) > 0 {
			strPtr := rec[24]
			intValue, _ := strconv.Atoi(strPtr)
			student.Data.AddressData.PostalCode = &intValue
		}
		if len(rec[25]) > 0 {
			student.Data.AddressData.Province = &rec[25]
		}
		if len(rec[26]) > 0 {
			strPtr := rec[26]
			intValue, _ := strconv.Atoi(strPtr)
			student.Data.AddressData.RT = &intValue
		}
		if len(rec[27]) > 0 {
			strPtr := rec[27]
			intValue, _ := strconv.Atoi(strPtr)
			student.Data.AddressData.RW = &intValue
		}
		if len(rec[28]) > 0 {
			student.Data.AddressData.SubDistrict = &rec[28]
		}

		if len(rec[29]) > 0 {
			strPtr := rec[29]
			intValue, _ := strconv.Atoi(strPtr)
			student.Data.AddressData.Village = &intValue
		}

		if len(rec[30]) > 0 {
			student.Data.MedicalHistoryData.BloodGroup = &rec[30]
		}
		if len(rec[31]) > 0 {
			strPtr := rec[31]
			intValue, _ := strconv.Atoi(strPtr)
			student.Data.MedicalHistoryData.BodyWeight = &intValue
		}
		if len(rec[32]) > 0 {
			student.Data.MedicalHistoryData.Disease = &rec[32]
		}
		if len(rec[33]) > 0 {
			student.Data.MedicalHistoryData.SpecialNeeds = &rec[33]
		}

		if len(rec[34]) > 0 {
			student.Data.SelfDevelopmentData.MandatoryExtracurricular = &rec[34]
		}
		if len(rec[35]) > 0 {
			student.Data.SelfDevelopmentData.NonAcademicAchievement = &rec[35]
		}
		if len(rec[36]) > 0 {
			student.Data.SelfDevelopmentData.OptionalExtracurricular = &rec[36]
		}
		if len(rec[37]) > 0 {
			student.Data.SelfDevelopmentData.QuranReadingLevel = &rec[37]
		}

		if len(rec[38]) > 0 {
			student.Data.AcademicData.BankAccountNumber = &rec[38]
		}
		if len(rec[39]) > 0 {
			strPtr := rec[39]
			intValue, _ := strconv.Atoi(strPtr)
			student.Data.AcademicData.NationalExamScore = &intValue
		}

		if len(rec[40]) > 0 {
			strPtr := rec[40]
			intValue, _ := strconv.Atoi(strPtr)
			student.Data.AcademicData.NationalIslamicExamScore = &intValue
		}

		if len(rec[41]) > 0 {
			student.Data.AcademicData.BankDKIAccountNumber = &rec[41]
		}

		if len(rec[42]) > 0 {
			student.Data.AcademicData.NationalExamNumber = &rec[42]
		}
		if len(rec[43]) > 0 {
			student.Data.AcademicData.OriginSchool = &rec[43]
		}
		if len(rec[44]) > 0 {
			student.Data.AcademicData.OriginSchoolType = &rec[44]
		}
		if len(rec[45]) > 0 {
			student.Data.AcademicData.SchoolStatus = &rec[45]
		}

		if len(rec[46]) > 0 {
			student.Data.AcademicData.SchoolAddress = &rec[46]
		}
		if len(rec[47]) > 0 {
			student.Data.AcademicData.SchoolAccreditation = &rec[47]
		}

		if len(rec[48]) > 0 {
			student.Data.SchoolTransferData.Origin = &rec[48]
		}

		if len(rec[49]) > 0 {
			student.Data.SchoolTransferData.Reason = &rec[49]
		}
		if len(rec[50]) > 0 {
			student.Data.SchoolTransferData.AcceptedInClass = &rec[50]
		}

		if len(rec[51]) > 0 {
			student.Data.StudentFatherData.BirthDate = &rec[51]
		}
		if len(rec[52]) > 0 {
			student.Data.StudentFatherData.BirthPlace = &rec[52]
		}
		if len(rec[53]) > 0 {
			student.Data.StudentFatherData.Education = &rec[53]
		}
		if len(rec[54]) > 0 {
			student.Data.StudentFatherData.Email = &rec[54]
		}

		if len(rec[55]) > 0 {
			student.Data.StudentFatherData.Existence = &rec[55]
		}
		if len(rec[56]) > 0 {
			strPtr := rec[56]
			intValue, _ := strconv.Atoi(strPtr)
			uintValue := uint(intValue)
			student.Data.StudentFatherData.Income = &uintValue
		}

		if len(rec[57]) > 0 {
			student.Data.StudentFatherData.Job = &rec[57]
		}

		if len(rec[58]) > 0 {
			student.Data.StudentFatherData.PhoneNumber = &rec[58]
		}
		if len(rec[59]) > 0 {
			student.Data.StudentFatherData.Religion = &rec[59]
		}

		if len(rec[60]) > 0 {
			student.Data.StudentFatherData.FullAddress = &rec[60]
		}
		if len(rec[61]) > 0 {
			student.Data.StudentFatherData.Fullname = &rec[61]
		}

		if len(rec[62]) > 0 {
			student.Data.StudentMotherData.BirthPlace = &rec[62]
		}
		if len(rec[63]) > 0 {
			student.Data.StudentMotherData.Education = &rec[63]
		}
		if len(rec[64]) > 0 {
			student.Data.StudentMotherData.Email = &rec[64]
		}

		if len(rec[65]) > 0 {
			student.Data.StudentMotherData.Existence = &rec[65]
		}
		if len(rec[66]) > 0 {
			strPtr := rec[66]
			intValue, _ := strconv.Atoi(strPtr)
			uintValue := uint(intValue)
			student.Data.StudentMotherData.Income = &uintValue
		}

		if len(rec[67]) > 0 {
			student.Data.StudentMotherData.Job = &rec[67]
		}

		if len(rec[68]) > 0 {
			student.Data.StudentMotherData.PhoneNumber = &rec[68]
		}
		if len(rec[69]) > 0 {
			student.Data.StudentMotherData.Religion = &rec[69]
		}

		if len(rec[70]) > 0 {
			student.Data.StudentMotherData.FullAddress = &rec[70]
		}
		if len(rec[71]) > 0 {
			student.Data.StudentMotherData.Fullname = &rec[71]
		}

		if len(rec[72]) > 0 {
			student.Data.StudentGuardianData.BirthPlace = &rec[72]
		}
		if len(rec[73]) > 0 {
			student.Data.StudentGuardianData.Education = &rec[73]
		}
		if len(rec[74]) > 0 {
			student.Data.StudentGuardianData.Email = &rec[74]
		}

		if len(rec[75]) > 0 {
			student.Data.StudentGuardianData.Existence = &rec[75]
		}
		if len(rec[76]) > 0 {
			strPtr := rec[76]
			intValue, _ := strconv.Atoi(strPtr)
			uintValue := uint(intValue)
			student.Data.StudentGuardianData.Income = &uintValue
		}

		if len(rec[77]) > 0 {
			student.Data.StudentGuardianData.Job = &rec[77]
		}

		if len(rec[78]) > 0 {
			student.Data.StudentGuardianData.PhoneNumber = &rec[78]
		}
		if len(rec[79]) > 0 {
			student.Data.StudentGuardianData.Religion = &rec[79]
		}

		if len(rec[80]) > 0 {
			student.Data.StudentGuardianData.FullAddress = &rec[80]
		}
		if len(rec[81]) > 0 {
			student.Data.StudentGuardianData.Fullname = &rec[81]
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

	if body.Nickname != nil {
		if len(*body.Nickname) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Nickname"}
		}
		newStudentData.Nickname = body.EthnicGroup
		isUpdate = true
	}

	if body.OriginSchoolNumber != nil {
		if len(*body.OriginSchoolNumber) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Origin School Number"}
		}
		newStudentData.EthnicGroup = body.EthnicGroup
		isUpdate = true
	}

	if body.DistanceToSchool != nil {
		if len(*body.DistanceToSchool) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Distance To School"}
		}
		newStudentData.DistanceToSchool = body.EthnicGroup
		isUpdate = true
	}

	if body.TransportationToSchool != nil {
		if len(*body.TransportationToSchool) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Transportation To School"}
		}
		newStudentData.TransportationToSchool = body.EthnicGroup
		isUpdate = true
	}

	if body.Hobby != nil {
		if len(*body.Hobby) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Hobby"}
		}
		newStudentData.Hobby = body.EthnicGroup
		isUpdate = true
	}

	if body.Ideal != nil {
		if len(*body.Ideal) <= 0 {
			return nil, false, &errmsg.ErrFieldIsWrong{FieldName: "Ideal"}
		}
		newStudentData.Ideal = body.EthnicGroup
		isUpdate = true
	}

	return &newStudentData, isUpdate, nil
}

func (s *impService) convertToAccountResponse(studentAccount *model.Account) (*rs.AccountResponse, error) {
	var tempAccount rs.AccountResponse
	err := copier.Copy(&tempAccount, studentAccount)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var tempStudent rs.StudentResponse
	err = copier.Copy(&tempStudent, studentAccount.Student)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var tempStudentData rs.StudentDataResponse
	err = copier.Copy(&tempStudentData, studentAccount.Student.StudentData)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	tempAccount.Student = &tempStudent

	if tempAccount.Student != nil {
		var tempStudentData rs.StudentDataResponse
		err = copier.Copy(&tempStudentData, studentAccount.Student.StudentData)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}
		tempAccount.Student.StudentData = &tempStudentData
	}

	return &tempAccount, nil
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

	gotAccounts, maxPage, rowCount, err := s.repo.Account().GetAllStudent(params)
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
		tempAccount, err := s.convertToAccountResponse(&account)
		if err != nil {
			return nil, err
		}

		response = append(response, *tempAccount)
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

func (s *impService) GetDetailByAccountID(c context.Context, id string) (*rs.StudentResponse, error) {
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &id)
	if err != nil {
		return nil, err
	}

	gotStudent, err := s.repo.Student().GetDetailByAccountID(*parsedStudentID)
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
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &id)
	if err != nil {
		return nil, err
	}

	gotStudentData, err := s.repo.StudentData().GetStudentDataByStudentID(*parsedStudentID)
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

	parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", body.Data.ClassroomID)
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
		ClassroomID: parsedClassroomID,
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

	parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", body.Data.ClassroomID)
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
		ClassroomID: parsedClassroomID,
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

		if err := s.repo.StudentData().CreateOne(tx, newStudentData); err != nil {
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
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &studentID)
	if err != nil {
		return err
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
				ID: *parsedStudentID,
			},
			ClassroomID: &parsedClassroomID,
		}

		err = s.repo.Transaction(func(tx *gorm.DB) error {
			if err := s.repo.Student().UpdateOne(tx, *parsedStudentID, &newStudent); err != nil {
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
			if err := s.repo.StudentData().UpdateOne(tx, *parsedStudentID, newStudentData); err != nil {
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
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &studentID)
	if err != nil {
		return err
	}

	var newFamilyData model.StudentFamilyData
	err = copier.Copy(&newFamilyData, body)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.StudentData().UpdateOneStudentFamilyData(tx, *parsedStudentID, &newFamilyData); err != nil {
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
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &studentID)
	if err != nil {
		return err
	}

	var newAddressData model.AddressData
	err = copier.Copy(&newAddressData, body)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.StudentData().UpdateOneAddressData(tx, *parsedStudentID, &newAddressData); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
