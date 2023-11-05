package service_subject

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	serviceutil "github.com/fadhln/lms-be/util/service_util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type SubjectService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error)
	GetAllSubjectByTeacherID(c context.Context, teacherID string, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error)
	GetAllSubjectByClassroomID(c context.Context, classroomID string, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error)

	GetDetailByID(c context.Context, id string) (*rs.SubjectResponse, error)

	GetAllSubjectComponent(c context.Context, params *rq.PaginationParams[model.SubjectComponent]) (*rs.PaginationResponse[any, rs.SubjectComponentResponse], error)
	GetAllSubjectComponentBySubjectID(c context.Context, subjectID string, params *rq.PaginationParams[model.SubjectComponent]) (*rs.PaginationResponse[any, rs.SubjectComponentResponse], error)

	GetSubjectComponentDetailByID(c context.Context, id string) (*rs.SubjectComponentResponse, error)

	CreateOne(c context.Context, newSubject *rq.SubjectRequest) error
	CreateOneWithClassroomID(c context.Context, classroomID string, newSubject *rq.SubjectRequest) error
	CreateOneSubjectComponent(c context.Context, newSubjectComp *rq.SubjectComponentRequest) error
	CreateOneSubjectComponentWithValidation(c context.Context, teacherId string, newSubjectComp *rq.SubjectComponentRequest) error

	CreateMass(c *gin.Context, requestFile *multipart.FileHeader) (*rs.MassCreateResponse, error)

	EditOne(c context.Context, subjectID string, newSubject *rq.SubjectRequest) error
	EditOneWithValidation(c context.Context, teacherId string, subjectID string, newSubject *rq.SubjectRequest) error
	EditOneSubjectComponent(c context.Context, subjectCompID string, newSubjectComp *rq.SubjectComponentRequest) error
	EditOneSubjectComponentWithValidation(c context.Context, teacherId string, subjectCompID string, newSubjectComp *rq.SubjectComponentRequest) error

	GetAllSubjectNotInClassroom(c context.Context, classroomID string, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error)
	AssignClassroomsToSubject(c context.Context, subjectID string, classroomIDs []string) error
	RemoveClassroomsFromSubject(c context.Context, subjectID string, classroomIDs []string) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) SubjectService {
	return &impService{
		repo: r,
	}
}

func (s *impService) convertToResponse(subject *model.Subject) (*rs.SubjectResponse, error) {
	var tempResponse rs.SubjectResponse
	err := copier.Copy(&tempResponse, subject)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.AccountResponse
	gotAccount, err := s.repo.Account().ReadOneByID(subject.Teacher.AccountID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	err = copier.Copy(&res, gotAccount)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var tempTeacher rs.TeacherResponse
	err = copier.Copy(&tempTeacher, subject.Teacher)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var tempSubjectComponent []rs.SubjectComponentResponse
	err = copier.Copy(&tempSubjectComponent, subject.SubjectComponents)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res.Teacher = &tempTeacher
	tempResponse.Teacher = res
	tempResponse.SubjectComponents = tempSubjectComponent

	return &tempResponse, nil
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotSubjects, maxPage, rowCount, err := s.repo.Subject().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.SubjectResponse]{
				Data: []rs.SubjectResponse{},
			}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotSubjects == nil {
		return &rs.PaginationResponse[any, rs.SubjectResponse]{
			Data: []rs.SubjectResponse{},
		}, nil
	}

	if len(*gotSubjects) < 1 {
		return &rs.PaginationResponse[any, rs.SubjectResponse]{
			Data: []rs.SubjectResponse{},
		}, nil
	}

	var response []rs.SubjectResponse

	for _, subject := range *gotSubjects {
		tempResponse, err := s.convertToResponse(&subject)
		if err != nil {
			return nil, err
		}

		response = append(response, *tempResponse)
	}

	res := rs.PaginationResponse[any, rs.SubjectResponse]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) GetAllSubjectByTeacherID(c context.Context, teacherID string, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error) {
	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", &teacherID)
	if err != nil {
		return nil, err
	}

	newParams := *params
	newParams.Data.TeacherID = *parsedTeacherID

	return s.GetAll(c, &newParams)
}

func (s *impService) GetAllSubjectByClassroomID(c context.Context, classroomID string, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error) {
	parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", &classroomID)
	if err != nil {
		return nil, err
	}

	newParams := *params
	newParams.Data.RelationClassroomSubjects = []model.RelationClassroomSubject{
		{ClassroomID: *parsedClassroomID},
	}

	return s.GetAll(c, &newParams)
}

func (s *impService) GetDetailByID(c context.Context, id string) (*rs.SubjectResponse, error) {
	parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", &id)
	if err != nil {
		return nil, err
	}

	gotSubject, err := s.repo.Subject().GetDetailByID(*parsedSubjectID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res, err := s.convertToResponse(gotSubject)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *impService) GetAllSubjectComponent(c context.Context, params *rq.PaginationParams[model.SubjectComponent]) (*rs.PaginationResponse[any, rs.SubjectComponentResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotSubjectCompoents, maxPage, rowCount, err := s.repo.Subject().GetAllComponent(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.SubjectComponentResponse]{
				Data: []rs.SubjectComponentResponse{},
			}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotSubjectCompoents == nil {
		return &rs.PaginationResponse[any, rs.SubjectComponentResponse]{
			Data: []rs.SubjectComponentResponse{},
		}, nil
	}

	if len(*gotSubjectCompoents) < 1 {
		return &rs.PaginationResponse[any, rs.SubjectComponentResponse]{
			Data: []rs.SubjectComponentResponse{},
		}, nil
	}

	var response []rs.SubjectComponentResponse
	err = copier.Copy(&response, gotSubjectCompoents)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res := rs.PaginationResponse[any, rs.SubjectComponentResponse]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) GetAllSubjectComponentBySubjectID(c context.Context, subjectID string, params *rq.PaginationParams[model.SubjectComponent]) (*rs.PaginationResponse[any, rs.SubjectComponentResponse], error) {
	parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", &subjectID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	newParams := *params
	newParams.Data.SubjectID = *parsedSubjectID

	return s.GetAllSubjectComponent(c, &newParams)
}

func (s *impService) GetSubjectComponentDetailByID(c context.Context, id string) (*rs.SubjectComponentResponse, error) {
	parsedSubjectComponentID, err := serviceutil.GetUUIDFromStringWithValidation("Subject Component ID", &id)
	if err != nil {
		return nil, err
	}

	gotSubjectComponent, err := s.repo.Subject().GetComponentDetailByComponentID(*parsedSubjectComponentID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.SubjectComponentResponse
	err = copier.Copy(&res, gotSubjectComponent)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}

func (s *impService) CreateOne(c context.Context, newSubject *rq.SubjectRequest) error {
	if newSubject.Name == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Name"}
	}

	if len(*newSubject.Name) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
	}

	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", newSubject.TeacherID)
	if err != nil {
		return err
	}

	subject := model.Subject{
		TeacherID: *parsedTeacherID,
		Name:      *newSubject.Name,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Subject().CreateOne(tx, &subject); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) CreateOneWithClassroomID(c context.Context, classroomID string, newSubject *rq.SubjectRequest) error {
	if newSubject.Name == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Name"}
	}

	if len(*newSubject.Name) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
	}

	parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", &classroomID)
	if err != nil {
		return err
	}

	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", newSubject.TeacherID)
	if err != nil {
		return err
	}

	gotClassroom, err := s.repo.Classroom().GetDetailByID(*parsedClassroomID)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	gotTeacher, err := s.repo.Teacher().GetTeacherByID(*parsedTeacherID)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	if gotTeacher.SchoolID.String() != gotClassroom.SchoolID.String() {
		return &errmsg.ErrANotSameB{A: "Teacher School", B: "Classroom School"}
	}

	newSubjectID := uuid.New()

	subject := model.Subject{
		Base:      model.Base{ID: newSubjectID},
		TeacherID: *parsedTeacherID,
		Name:      *newSubject.Name,
		RelationClassroomSubjects: []model.RelationClassroomSubject{
			{
				SubjectID:   newSubjectID,
				ClassroomID: gotClassroom.ID,
			},
		},
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Subject().CreateOne(tx, &subject); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) validateSubjectComponentPercentage(c context.Context, percentage *int, subjectID string) error {
	if percentage == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Percentage"}
	}

	if *percentage < 0 {
		return &errmsg.ErrMinAmount{Amount: 0}
	}

	if *percentage > 100 {
		return &errmsg.ErrMaxAmount{Amount: 100}
	}

	getAllComponentParams := rq.PaginationParams[model.SubjectComponent]{
		Page:      1,
		SortOrder: "ASC",
		Limit:     999,
	}

	gotAllComponentFromSubject, err := s.GetAllSubjectComponentBySubjectID(c, subjectID, &getAllComponentParams)
	if err != nil {
		if !(errors.Is(err, gorm.ErrEmptySlice)) {
			return err
		}
	}

	var totalComponentPercentage int
	for _, gotComponent := range gotAllComponentFromSubject.Data {
		totalComponentPercentage = totalComponentPercentage + gotComponent.Percentage
	}

	if (totalComponentPercentage + *percentage) > 100 {
		return &errmsg.ErrMaxAmount{Amount: 100}
	}

	return nil
}

func (s *impService) CreateOneSubjectComponent(c context.Context, newSubjectComp *rq.SubjectComponentRequest) error {
	if newSubjectComp.Name == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Name"}
	}

	if len(*newSubjectComp.Name) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
	}

	if newSubjectComp.Percentage == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Percentage"}
	}

	if newSubjectComp.SubjectID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Subject ID"}
	}

	parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", newSubjectComp.SubjectID)
	if err != nil {
		return err
	}

	err = s.validateSubjectComponentPercentage(c, newSubjectComp.Percentage, *newSubjectComp.SubjectID)
	if err != nil {
		return err
	}

	subjectComp := model.SubjectComponent{
		Name:       *newSubjectComp.Name,
		SubjectID:  *parsedSubjectID,
		Percentage: *newSubjectComp.Percentage,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Subject().CreateOneComponent(tx, &subjectComp); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) CreateOneSubjectComponentWithValidation(c context.Context, teacherId string, newSubjectComp *rq.SubjectComponentRequest) error {
	if newSubjectComp.SubjectID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Subject ID"}
	}

	gotSubject, err := s.GetDetailByID(c, *newSubjectComp.SubjectID)
	if err != nil {
		return err
	}

	if gotSubject.TeacherID.String() != teacherId {
		return &errmsg.ErrUserIsNot{FieldName: "Subject Teacher"}
	}

	return s.CreateOneSubjectComponent(c, newSubjectComp)
}

func (s *impService) EditOne(c context.Context, subjectID string, body *rq.SubjectRequest) error {
	parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", &subjectID)
	if err != nil {
		return err
	}

	gotSubject, err := s.repo.Subject().GetDetailByID(*parsedSubjectID)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	var newSubject model.Subject

	if body.Name != nil {
		if len(*body.Name) <= 3 {
			return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
		}

		newSubject.Name = *body.Name
	}

	if body.TeacherID != nil {
		parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", body.TeacherID)
		if err != nil {
			return err
		}

		newSubject.TeacherID = *parsedTeacherID

		gotRelation, _, _, _ := s.repo.RelationClassroomSubject().GetAll(
			&rq.PaginationParams[model.RelationClassroomSubject]{
				Limit: 99,
				Page:  1,
				Data: model.RelationClassroomSubject{
					SubjectID: gotSubject.ID,
				},
			})

		if gotRelation != nil {
			for _, relation := range *gotRelation {
				if gotSubject.Teacher.SchoolID.String() != relation.Classroom.ID.String() {
					return &errmsg.ErrANotSameB{A: "Teacher School", B: "Classroom School"}
				}
			}
		}
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Subject().UpdateOne(tx, *parsedSubjectID, &newSubject); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func getSubjectCSV(requestFile *multipart.FileHeader) (*[]rq.SubjectRequestWithTeacherEmail, error) {
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

	var subjects []rq.SubjectRequestWithTeacherEmail
	for idx, rec := range records {
		if idx == 0 {
			continue
		}

		subject := rq.SubjectRequestWithTeacherEmail{
			Name:         rec[0],
			TeacherEmail: rec[1],
		}

		subjects = append(subjects, subject)
	}

	return &subjects, nil
}

func (s *impService) processMassCreate(c context.Context, requests *[]rq.SubjectRequestWithTeacherEmail, reportFileName string, newMassCreate *model.MassCreate) {
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
	for idx, tempSubject := range *requests {
		gotTeacherAccount, err := s.repo.Account().ReadOneByEmail(tempSubject.TeacherEmail)
		if err != nil {
			errMsgs = append(errMsgs, model.ErrorMsg{
				Row:     idx + 1,
				Message: err.Error()})
			continue
		}

		gotTeacher, err := s.repo.Teacher().GetDetailByAccountID(gotTeacherAccount.ID)
		if err != nil {
			errMsgs = append(errMsgs, model.ErrorMsg{
				Row:     idx + 1,
				Message: err.Error()})
			continue
		}
		if gotTeacher == nil {
			errMsgs = append(errMsgs, model.ErrorMsg{
				Row:     idx + 1,
				Message: "Teacher is missing"})
			continue
		}

		teacherID := gotTeacher.ID.String()

		err = s.CreateOne(c, &rq.SubjectRequest{Name: &tempSubject.Name, TeacherID: &teacherID})
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

	newSubjects, err := getSubjectCSV(requestFile)
	if err != nil {
		return nil, errmsg.ErrRequestFileInvalid
	}

	nowStr := time.Now().Format(constants.FilenameTimeFormat)
	requestFileName := fmt.Sprintf("subject-request-%s-%s.csv", nowStr, uuid.New().String())
	reportFileName := fmt.Sprintf("subject-report-%s-%s.csv", nowStr, uuid.New().String())

	err = c.SaveUploadedFile(requestFile, "./file/input/"+requestFileName)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	newMassCreateID := uuid.New()
	var newMassCreate = model.MassCreate{
		Base:           model.Base{ID: newMassCreateID},
		Status:         constants.MASS_CREATE_STATUS_PROCESSING,
		Destination:    constants.DEST_SUBJECT,
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

	go s.processMassCreate(c, newSubjects, reportFileName, &newMassCreate)

	var res rs.MassCreateResponse
	err = copier.Copy(&res, newMassCreate)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}

func (s *impService) EditOneWithValidation(c context.Context, teacherId string, subjectID string, newSubject *rq.SubjectRequest) error {
	gotSubject, err := s.GetDetailByID(c, subjectID)
	if err != nil {
		return err
	}

	if gotSubject.TeacherID.String() != teacherId {
		return &errmsg.ErrUserIsNot{FieldName: "Subject Teacher"}
	}

	return s.EditOne(c, subjectID, newSubject)
}

func (s *impService) EditOneSubjectComponent(c context.Context, subjectCompID string, body *rq.SubjectComponentRequest) error {
	if len(subjectCompID) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Subject Component ID"}
	}

	parsedSubjectComponentID, err := serviceutil.GetUUIDFromStringWithValidation("Subject Component", &subjectCompID)
	if err != nil {
		return err
	}

	var newSubjectComponent model.SubjectComponent

	if body.Name != nil {
		if len(*body.Name) <= 3 {
			return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
		}

		newSubjectComponent.Name = *body.Name
	}

	if body.SubjectID != nil {
		parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", body.SubjectID)
		if err != nil {
			return err
		}

		newSubjectComponent.SubjectID = *parsedSubjectID
	}

	if body.Percentage != nil {
		if body.SubjectID != nil {
			err = s.validateSubjectComponentPercentage(c, body.Percentage, *body.SubjectID)
			if err != nil {
				return err
			}
		} else {
			existingSubjectComponent, err := s.GetSubjectComponentDetailByID(c, subjectCompID)
			if err != nil {
				return err
			}

			err = s.validateSubjectComponentPercentage(c, body.Percentage, existingSubjectComponent.SubjectID.String())
			if err != nil {
				return err
			}
		}

		newSubjectComponent.Percentage = *body.Percentage
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Subject().UpdateOneComponent(tx, *parsedSubjectComponentID, &newSubjectComponent); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) EditOneSubjectComponentWithValidation(c context.Context, teacherId string, subjectCompID string, newSubjectComp *rq.SubjectComponentRequest) error {
	gotSubjectComponent, err := s.GetSubjectComponentDetailByID(c, subjectCompID)
	if err != nil {
		return err
	}

	gotSubject, err := s.GetDetailByID(c, gotSubjectComponent.SubjectID.String())
	if err != nil {
		return err
	}

	if gotSubject.TeacherID.String() != teacherId {
		return &errmsg.ErrUserIsNot{FieldName: "Subject Teacher"}
	}

	return s.EditOneSubjectComponent(c, subjectCompID, newSubjectComp)

}

func (s *impService) GetAllSubjectNotInClassroom(c context.Context, classroomID string, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error) {
	if len(classroomID) < 3 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Classroom ID"}
	}

	parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", &classroomID)
	if err != nil {
		return nil, err
	}

	newParams := *params
	newParams.Data.RelationClassroomSubjects = []model.RelationClassroomSubject{
		{}, {ClassroomID: *parsedClassroomID},
	}

	return s.GetAll(c, &newParams)
}

func (s *impService) AssignClassroomsToSubject(c context.Context, subjectID string, classroomIDs []string) error {
	if len(subjectID) < 3 {
		return &errmsg.ErrIsEmpty{FieldName: "Subject ID"}
	}

	parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", &subjectID)
	if err != nil {
		return err
	}

	gotSubject, err := s.GetDetailByID(c, parsedSubjectID.String())
	if err != nil {
		return err
	}

	parsedClassroomIDs := []uuid.UUID{}
	for _, classroomID := range classroomIDs {
		parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", &classroomID)
		if err != nil {
			return err
		}

		gotClassroom, err := s.repo.Classroom().GetDetailByID(*parsedClassroomID)
		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}

		if gotSubject.Teacher.Teacher.SchoolID != gotClassroom.SchoolID {
			return &errmsg.ErrANotSameB{A: "Teacher School", B: "Classroom School"}
		}

		parsedClassroomIDs = append(parsedClassroomIDs, *parsedClassroomID)
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		for _, parsedClassroomID := range parsedClassroomIDs {
			newRelation := model.RelationClassroomSubject{
				ClassroomID: parsedClassroomID,
				SubjectID:   *parsedSubjectID,
			}

			if err := s.repo.RelationClassroomSubject().CreateOne(tx, &newRelation); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) RemoveClassroomsFromSubject(c context.Context, subjectID string, classroomIDs []string) error {
	if len(subjectID) < 3 {
		return &errmsg.ErrIsEmpty{FieldName: "Subject ID"}
	}

	parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", &subjectID)
	if err != nil {
		return err
	}

	parsedClassroomIDs := []uuid.UUID{}
	for _, classroomID := range classroomIDs {
		parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", &classroomID)
		if err != nil {
			return err
		}

		parsedClassroomIDs = append(parsedClassroomIDs, *parsedClassroomID)
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		for _, parsedClassroomID := range parsedClassroomIDs {
			if err := s.repo.RelationClassroomSubject().DeleteOne(tx, parsedClassroomID, *parsedSubjectID); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
