package service_subject

import (
	"context"
	"errors"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
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
	CreateOneSubjectComponent(c context.Context, newSubjectComp *rq.SubjectComponentRequest) error
	CreateOneSubjectComponentWithValidation(c context.Context, teacherId string, newSubjectComp *rq.SubjectComponentRequest) error

	EditOne(c context.Context, subjectID string, newSubject *rq.SubjectRequest) error
	EditOneWithValidation(c context.Context, teacherId string, subjectID string, newSubject *rq.SubjectRequest) error
	EditOneSubjectComponent(c context.Context, subjectCompID string, newSubjectComp *rq.SubjectComponentRequest) error
	EditOneSubjectComponentWithValidation(c context.Context, teacherId string, subjectCompID string, newSubjectComp *rq.SubjectComponentRequest) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) SubjectService {
	return &impService{
		repo: r,
	}
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

	gotSubjects, maxPage, err := s.repo.Subject().GetAll(params)
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
		var tempResponse rs.SubjectResponse
		err = copier.Copy(&tempResponse, subject)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		var tempSubjectComponent rs.SubjectComponentResponse
		err = copier.Copy(&tempSubjectComponent, subject.SubjectComponents)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}
	}

	res := rs.PaginationResponse[any, rs.SubjectResponse]{
		MaxPage:         maxPage,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) GetAllSubjectByTeacherID(c context.Context, teacherID string, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error) {
	if len(teacherID) < 3 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	parsedTeacherID, err := uuid.Parse(teacherID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if parsedTeacherID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	newParams := *params
	newParams.Data.TeacherID = parsedTeacherID

	return s.GetAll(c, &newParams)
}

func (s *impService) GetAllSubjectByClassroomID(c context.Context, classroomID string, params *rq.PaginationParams[model.Subject]) (*rs.PaginationResponse[any, rs.SubjectResponse], error) {
	if len(classroomID) < 3 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Classroom ID"}
	}

	parsedClassroomID, err := uuid.Parse(classroomID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if parsedClassroomID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Classroom ID"}
	}

	newParams := *params
	newParams.Data.Schedules = []model.Schedule{
		{ClassroomID: parsedClassroomID},
	}

	return s.GetAll(c, &newParams)
}

func (s *impService) GetDetailByID(c context.Context, id string) (*rs.SubjectResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	gotSubject, err := s.repo.Subject().GetDetailByID(parsedID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.SubjectResponse
	err = copier.Copy(&res, gotSubject)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var tempComponents []rs.SubjectComponentResponse
	err = copier.Copy(&tempComponents, gotSubject.SubjectComponents)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res.SubjectComponents = tempComponents

	return &res, nil
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

	gotSubjectCompoents, maxPage, err := s.repo.Subject().GetAllComponent(params)
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
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) GetAllSubjectComponentBySubjectID(c context.Context, subjectID string, params *rq.PaginationParams[model.SubjectComponent]) (*rs.PaginationResponse[any, rs.SubjectComponentResponse], error) {
	if len(subjectID) < 3 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Subject ID"}
	}

	parsedSubjectID, err := uuid.Parse(subjectID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if parsedSubjectID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Classroom ID"}
	}

	newParams := *params
	newParams.Data.SubjectID = parsedSubjectID

	return s.GetAllSubjectComponent(c, &newParams)
}

func (s *impService) GetSubjectComponentDetailByID(c context.Context, id string) (*rs.SubjectComponentResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "id"}
	}

	if parsedID == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	gotSubjectComponent, err := s.repo.Subject().GetComponentDetailByComponentID(parsedID)
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

	if newSubject.TeacherID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Teacher ID"}
	}

	parsedTeacherID, err := uuid.Parse(*newSubject.TeacherID)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	if parsedTeacherID == uuid.Nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Teacher ID"}
	}

	subject := model.Subject{
		TeacherID: parsedTeacherID,
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
		Limit: 999,
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

	parsedSubjectID, err := uuid.Parse(*newSubjectComp.SubjectID)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	if parsedSubjectID == uuid.Nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Subject ID"}
	}

	err = s.validateSubjectComponentPercentage(c, newSubjectComp.Percentage, *newSubjectComp.SubjectID)
	if err != nil {
		return err
	}

	subjectComp := model.SubjectComponent{
		Name:       *newSubjectComp.Name,
		SubjectID:  parsedSubjectID,
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
	if len(subjectID) <= 3 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Subject ID"}
	}

	parsedSubjectID, err := uuid.Parse(subjectID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Subject ID"}
	}

	if parsedSubjectID == uuid.Nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Subject ID"}
	}

	var newSubject model.Subject

	if body.Name != nil {
		if len(*body.Name) <= 3 {
			return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
		}

		newSubject.Name = *body.Name
	}

	if body.TeacherID != nil {
		parsedTeacherID, err := uuid.Parse(*body.TeacherID)
		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}

		if parsedTeacherID == uuid.Nil {
			return &errmsg.ErrFieldIsWrong{FieldName: "Teacher ID"}
		}

		newSubject.TeacherID = parsedTeacherID
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Subject().UpdateOne(tx, parsedSubjectID, &newSubject); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
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

	parsedSubjectComponentID, err := uuid.Parse(subjectCompID)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Subject Component ID"}
	}

	if parsedSubjectComponentID == uuid.Nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Subject Component ID"}
	}

	var newSubjectComponent model.SubjectComponent

	if body.Name != nil {
		if len(*body.Name) <= 3 {
			return &errmsg.ErrFieldIsWrong{FieldName: "Name"}
		}

		newSubjectComponent.Name = *body.Name
	}

	if body.SubjectID != nil {
		parsedSubjectID, err := uuid.Parse(*body.SubjectID)
		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}

		if parsedSubjectID == uuid.Nil {
			return &errmsg.ErrFieldIsWrong{FieldName: "Subject ID"}
		}

		newSubjectComponent.SubjectID = parsedSubjectID
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
		if err := s.repo.Subject().UpdateOneComponent(tx, parsedSubjectComponentID, &newSubjectComponent); err != nil {
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
