package service_classroom

import (
	"context"
	"encoding/csv"
	"encoding/json"
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
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type ClassroomService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.Classroom]) (*rs.PaginationResponse[any, rs.ClassroomResponse], error)
	GetDetailByID(c context.Context, id string) (*rs.ClassroomResponse, error)

	CreateOne(c context.Context, newClassroom *rq.ClassroomRequest) error
	CreateMass(c *gin.Context, academicYearID string, schoolID string, requestFile *multipart.FileHeader) (*rs.MassCreateResponse, error)

	EditOne(c context.Context, newClassroom *rq.ClassroomRequest) error

	DeleteOne(c context.Context, id string) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) ClassroomService {
	return &impService{
		repo: r,
	}
}

func getClassroomCSV(requestFile *multipart.FileHeader, academicYearID string, schoolID string) (*[]rq.ClassroomRequest, error) {
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

	var classrooms []rq.ClassroomRequest
	for idx, rec := range records {
		if idx == 0 {
			continue
		}

		classroom := rq.ClassroomRequest{
			Name:           &rec[0],
			TeacherEmail:   &rec[1],
			AcademicYearID: &academicYearID,
			SchoolID:       &schoolID,
		}

		classrooms = append(classrooms, classroom)
	}

	return &classrooms, nil
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[model.Classroom]) (*rs.PaginationResponse[any, rs.ClassroomResponse], error) {
	gotClassrooms, maxPage, err := s.repo.Classroom().GetAll(params)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotClassrooms == nil {
		return &rs.PaginationResponse[any, rs.ClassroomResponse]{
			Data: []rs.ClassroomResponse{},
		}, nil
	}

	if len(*gotClassrooms) < 1 {
		return &rs.PaginationResponse[any, rs.ClassroomResponse]{
			Data: []rs.ClassroomResponse{},
		}, nil
	}

	var response []rs.ClassroomResponse

	for _, classroom := range *gotClassrooms {
		var tempResponse rs.ClassroomResponse
		err = copier.Copy(&tempResponse, classroom)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		var tempAcademicYear rs.AcademicYearResponse
		err = copier.Copy(&tempAcademicYear, classroom.AcademicYear)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		var tempTeacher rs.TeacherResponse
		err = copier.Copy(&tempTeacher, classroom.Teacher)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		tempResponse.AcademicYear = tempAcademicYear
		tempResponse.Teacher = tempTeacher

		response = append(response, tempResponse)
	}

	res := rs.PaginationResponse[any, rs.ClassroomResponse]{
		MaxPage:         maxPage,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) GetDetailByID(c context.Context, id string) (*rs.ClassroomResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	gotClassroom, err := s.repo.Classroom().GetDetailByID(parsedID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.ClassroomResponse
	err = copier.Copy(&res, gotClassroom)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}

func (s *impService) CreateOne(c context.Context, newClassroom *rq.ClassroomRequest) error {
	if newClassroom.Name == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Name"}
	}

	if len(*newClassroom.Name) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
	}

	if newClassroom.AcademicYearID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Academic Year ID"}
	}

	parsedAcademicYearID, err := uuid.Parse(*newClassroom.AcademicYearID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Academic Year ID"}
	}

	if parsedAcademicYearID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "Academic Year ID"}
	}

	if newClassroom.SchoolID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "School ID"}
	}

	parsedSchoolID, err := uuid.Parse(*newClassroom.SchoolID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "School ID"}
	}

	if parsedSchoolID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "School ID"}
	}

	if newClassroom.TeacherEmail != nil {
		if len(*newClassroom.TeacherEmail) <= 0 {
			return &errmsg.ErrIsEmpty{FieldName: "Teacher Email"}
		}

		teacherEmail := strings.ToLower(*newClassroom.TeacherEmail)
		if !(util.IsEmailValid(teacherEmail)) {
			return &errmsg.ErrFieldIsWrong{FieldName: "Email"}
		}

		gotAccount, err := s.repo.Account().ReadOneByEmail(teacherEmail)
		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}

		gotTeacher, err := s.repo.Teacher().GetDetailByAccountID(gotAccount.ID)
		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}

		gotTeacherID := gotTeacher.ID.String()
		newClassroom.TeacherID = &gotTeacherID
	}

	if newClassroom.TeacherID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	parsedTeacherID, err := uuid.Parse(*newClassroom.TeacherID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Teacher ID"}
	}

	if parsedTeacherID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	if newClassroom.Code == nil {
		randomCode := util.RandString(6)
		newClassroom.Code = &randomCode
	}

	classroom := model.Classroom{
		Name:           *newClassroom.Name,
		AcademicYearID: parsedAcademicYearID,
		TeacherID:      parsedTeacherID,
		SchoolID:       parsedSchoolID,
		Code:           *newClassroom.Code,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Classroom().CreateOne(tx, &classroom); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) processMassCreate(c context.Context, requests *[]rq.ClassroomRequest, reportFileName string, newMassCreate *model.MassCreate) {
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
	for idx, newClassroom := range *requests {
		err := s.CreateOne(c, &newClassroom)
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

func (s *impService) CreateMass(c *gin.Context, academicYearID string, schoolID string, requestFile *multipart.FileHeader) (
	*rs.MassCreateResponse, error) {

	newClassrooms, err := getClassroomCSV(requestFile, academicYearID, schoolID)
	if err != nil {
		return nil, errmsg.ErrRequestFileInvalid
	}

	nowStr := time.Now().Format(constants.FilenameTimeFormat)
	requestFileName := fmt.Sprintf("classroom-request-%s-%s.csv", nowStr, uuid.New().String())
	reportFileName := fmt.Sprintf("classroom-report-%s-%s.csv", nowStr, uuid.New().String())

	err = c.SaveUploadedFile(requestFile, "./file/input/"+requestFileName)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	newMassCreateID := uuid.New()
	var newMassCreate = model.MassCreate{
		Base:           model.Base{ID: newMassCreateID},
		Status:         constants.MASS_CREATE_STATUS_PROCESSING,
		Destination:    constants.DEST_CLASSROOM,
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

	go s.processMassCreate(c, newClassrooms, reportFileName, &newMassCreate)

	var res rs.MassCreateResponse
	err = copier.Copy(&res, newMassCreate)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}

func (s *impService) EditOne(c context.Context, newClassroom *rq.ClassroomRequest) error {
	if newClassroom.ID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	parsedID, err := uuid.Parse(*newClassroom.ID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	classroom := model.Classroom{
		Base: model.Base{ID: parsedID},
	}

	if newClassroom.AcademicYearID != nil {
		parsedAcademicYearID, err := uuid.Parse(*newClassroom.AcademicYearID)
		if err != nil {
			return &errmsg.ErrFieldIsWrong{FieldName: "Academic Year ID"}
		}

		if parsedAcademicYearID == uuid.Nil {
			return &errmsg.ErrIsEmpty{FieldName: "Academic Year ID"}
		}

		classroom.AcademicYearID = parsedAcademicYearID
	}

	if newClassroom.TeacherID != nil {
		parsedTeacherID, err := uuid.Parse(*newClassroom.TeacherID)
		if err != nil {
			return &errmsg.ErrFieldIsWrong{FieldName: "Teacher ID"}
		}

		if parsedTeacherID == uuid.Nil {
			return &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
		}

		classroom.TeacherID = parsedTeacherID
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Classroom().UpdateOne(tx, parsedID, &classroom); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) DeleteOne(c context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Classroom().DeleteOne(tx, parsedID); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
