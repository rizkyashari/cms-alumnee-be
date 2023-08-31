package service_classroom

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
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
	CreateMass(c *gin.Context, requestFile *multipart.FileHeader) (*rs.MassCreateResponse, error)

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

func getClassroomCSV(requestFile *multipart.FileHeader) (*[]rq.ClassroomRequest, error) {
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
			AcademicYearID: &rec[1],
			TeacherID:      &rec[2],
		}

		fmt.Println(">>>>> classroom.AcademicYearID >>>>>", *classroom.AcademicYearID)

		classrooms = append(classrooms, classroom)
	}

	return &classrooms, nil
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[model.Classroom]) (*rs.PaginationResponse[any, rs.ClassroomResponse], error) {
	gotClassrooms, maxPage, err := s.repo.Classroom().GetAll(params)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var response []rs.ClassroomResponse
	err = copier.Copy(&response, gotClassrooms)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

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
	if len(*newClassroom.Name) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
	}

	parsedAcademicYearID, err := uuid.Parse(*newClassroom.AcademicYearID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Academic Year ID"}
	}

	if parsedAcademicYearID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "Academic Year ID"}
	}

	parsedTeacherID, err := uuid.Parse(*newClassroom.TeacherID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Teacher ID"}
	}

	if parsedTeacherID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	classroom := model.Classroom{
		Name:           *newClassroom.Name,
		AcademicYearID: parsedAcademicYearID,
		TeacherID:      parsedTeacherID,
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

func (s *impService) CreateMass(c *gin.Context, requestFile *multipart.FileHeader) (
	*rs.MassCreateResponse, error) {

	newClassrooms, err := getClassroomCSV(requestFile)
	if err != nil {
		return nil, errmsg.ErrRequestFileInvalid
	}

	nowStr := time.Now().Format("2006-Jan-02")
	requestFileName := fmt.Sprintf("classroom-request-%s-%s.csv", nowStr, uuid.New().String())
	reportFileName := fmt.Sprintf("classroom-report-%s-%s.csv", nowStr, uuid.New().String())

	err = c.SaveUploadedFile(requestFile, "./file/input/"+requestFileName)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if len(*newClassrooms) >= 350 {
		return nil, &errmsg.ErrMaxAmount{Amount: 350}
	}

	successCount := 0
	var errMsgs []model.ErrorMsg
	for idx, newClassroom := range *newClassrooms {
		err := s.CreateOne(c, &newClassroom)
		if err != nil {
			errMsgs = append(errMsgs, model.ErrorMsg{Row: idx + 1, Message: err.Error()})
			continue
		}
		successCount++
	}

	errMsgsStr, _ := json.Marshal(errMsgs)
	file, err := os.Create("./file/output/" + reportFileName)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write(model.GetErrorMsgHeader())
	for _, errMsg := range errMsgs {
		writer.Write(model.GetErrorMsgRow(&errMsg))
	}

	var newMassCreate = model.MassCreate{
		Destination:    model.DEST_CLASSROOM,
		RequestFileUrl: requestFileName,
		ReportFileUrl:  reportFileName,
		SuccessCount:   successCount,
		ErrorCount:     len(errMsgs),
		ErrorMessages:  string(errMsgsStr),
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

	var res rs.MassCreateResponse
	err = copier.Copy(&res, newMassCreate)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}
	res.ErrorMessages = errMsgs

	return &res, nil
}

func (s *impService) EditOne(c context.Context, newClassroom *rq.ClassroomRequest) error {
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
		if err := s.repo.Classroom().UpdateOne(tx, &classroom); err != nil {
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
