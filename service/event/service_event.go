package service_event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	serviceutil "github.com/fadhln/lms-be/util/service_util"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type EventService interface {
	GetDetailByID(c context.Context, eventID string) (*rs.EventResponse, error)
	GetAll(c context.Context, params rq.EventParams) (*rs.ManyEventReponse, error)
	GetAllByClassroomID(c context.Context, classroomID string, params rq.EventParams) (*rs.ManyEventReponse, error)
	GetAllByTeacherID(c context.Context, teacherID string, params rq.EventParams) (*rs.ManyEventReponse, error)
	GetAllByStudentID(c context.Context, studentID string, params rq.EventParams) (*rs.ManyEventReponse, error)
	CreateOne(c context.Context, newEvent rq.CreateSingleEventRequest) error
	CreateRepeated(c context.Context, newEvents rq.CreateRepeatWeeklyEventRequest) error
	CreateManyRepeated(c context.Context, manyNewEvents []rq.CreateRepeatWeeklyEventRequest) error
	EditOne(c context.Context, newEvent rq.EditEventRequest) error
	DeleteOne(c context.Context, eventID string) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) EventService {
	return &impService{
		repo: r,
	}
}

func (s *impService) GetDetailByID(c context.Context, eventID string) (*rs.EventResponse, error) {
	parsedEventID, err := serviceutil.GetUUIDFromStringWithValidation("Event ID", &eventID)
	if err != nil {
		return nil, err
	}

	event, err := s.repo.Event().GetDetailByID(*parsedEventID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var eventResponse rs.EventResponse
	err = copier.Copy(&eventResponse, event)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var classroom rs.ClassroomResponse
	err = copier.Copy(&classroom, event.Classroom)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}
	eventResponse.Classroom = classroom

	if event.Type == constants.EVENT_TYPE_SUBJECT {
		var subject rs.SubjectResponse
		err = copier.Copy(&subject, event.RelationClassroomSubject.Subject)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		var teacher rs.AccountResponse
		gotTeacher, _ := s.repo.Teacher().GetDetailByTeacherID(event.RelationClassroomSubject.Subject.TeacherID)

		gotAccount, err := s.repo.Account().ReadOneByID(gotTeacher.AccountID)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		err = copier.Copy(&teacher, gotAccount)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		subject.Teacher = teacher
		eventResponse.Subject = &subject
	}

	return &eventResponse, nil
}

func (s *impService) GetAll(c context.Context, params rq.EventParams) (*rs.ManyEventReponse, error) {
	res := rs.ManyEventReponse{
		Begin: params.Begin,
		End:   params.End,
	}

	gotEvents, err := s.repo.Event().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Events = []rs.EventResponse{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if len(gotEvents) < 1 {
		return &res, nil
	}

	var eventResponses []rs.EventResponse
	for _, event := range gotEvents {
		var eventResponse rs.EventResponse
		err = copier.Copy(&eventResponse, event)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		var classroom rs.ClassroomResponse
		err = copier.Copy(&classroom, event.Classroom)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}
		eventResponse.Classroom = classroom

		if event.Type == constants.EVENT_TYPE_SUBJECT {
			var subject rs.SubjectResponse
			err = copier.Copy(&subject, event.RelationClassroomSubject.Subject)
			if err != nil {
				return nil, &errmsg.ErrInternal{Err: err}
			}

			var teacher rs.AccountResponse
			gotTeacher, _ := s.repo.Teacher().GetDetailByTeacherID(event.RelationClassroomSubject.Subject.TeacherID)

			gotAccount, err := s.repo.Account().ReadOneByID(gotTeacher.AccountID)
			if err != nil {
				return nil, &errmsg.ErrInternal{Err: err}
			}

			err = copier.Copy(&teacher, gotAccount)
			if err != nil {
				return nil, &errmsg.ErrInternal{Err: err}
			}

			subject.Teacher = teacher
			eventResponse.Subject = &subject
		}

		eventResponses = append(eventResponses, eventResponse)
	}

	res.Events = eventResponses
	return &res, nil
}

func (s *impService) GetAllByClassroomID(c context.Context, classroomID string, params rq.EventParams) (*rs.ManyEventReponse, error) {
	parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", &classroomID)
	if err != nil {
		return nil, err
	}

	newParams := params
	newParams.Data = &rq.EventParamsData{
		ClassroomID: parsedClassroomID,
	}

	return s.GetAll(c, newParams)
}

func (s *impService) GetAllByTeacherID(c context.Context, teacherID string, params rq.EventParams) (*rs.ManyEventReponse, error) {
	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", &teacherID)
	if err != nil {
		return nil, err
	}

	newParams := params
	newParams.Data = &rq.EventParamsData{
		TeacherID: parsedTeacherID,
	}

	return s.GetAll(c, newParams)
}

func (s *impService) GetAllByStudentID(c context.Context, studentID string, params rq.EventParams) (*rs.ManyEventReponse, error) {
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &studentID)
	if err != nil {
		return nil, err
	}

	newParams := params
	newParams.Data = &rq.EventParamsData{
		StudentID: parsedStudentID,
	}

	return s.GetAll(c, newParams)
}

func (s *impService) CreateOne(c context.Context, newEvent rq.CreateSingleEventRequest) error {
	newEventReq := model.Event{}
	parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", &newEvent.ClassroomID)
	if err != nil {
		return err
	}
	newEventReq.ClassroomID = *parsedClassroomID

	if !util.IsValidConstant(newEvent.Type, constants.EventTypeMap) {
		return &errmsg.ErrFieldIsWrong{FieldName: "Event Type"}
	}

	timeBegin, err := util.EventDateParse(newEvent.BeginDate)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Begin Date"}
	}

	timeEnd, err := util.EventDateParse(newEvent.EndDate)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "End Date"}
	}

	if !timeEnd.After(*timeBegin) {
		return &errmsg.ErrFieldIsWrong{FieldName: "End Date"}
	}

	newEventReq.BeginDate = *timeBegin
	newEventReq.EndDate = *timeEnd
	newEventReq.Description = newEvent.Description

	newEventReq.Type = newEvent.Type
	if newEventReq.Type == constants.EVENT_TYPE_NORMAL {
		newEventReq.Title = newEvent.Title
	} else if newEvent.Type == constants.EVENT_TYPE_SUBJECT {
		parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", newEvent.SubjectID)
		if err != nil {
			return err
		}

		gotRelation, err := s.repo.RelationClassroomSubject().GetByClassroomIDAndSubjectID(parsedClassroomID, parsedSubjectID)
		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}

		newEventReq.Description = newEvent.Description
		newEventReq.RelationClassroomSubjectID = &gotRelation.ID
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Event().CreateOne(tx, newEventReq); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) CreateRepeated(c context.Context, newEvents rq.CreateRepeatWeeklyEventRequest) error {
	// Validate props
	beginTime, err := util.TimeParse(newEvents.BeginTime)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Begin Time"}
	}

	endTime, err := util.TimeParse(newEvents.EndTime)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "End Time"}
	}

	if !endTime.After(*beginTime) {
		return &errmsg.ErrFieldIsWrong{FieldName: "End Time"}
	}

	beginHour := beginTime.Hour()
	beginMinute := beginTime.Minute()

	endHour := endTime.Hour()
	endMinute := endTime.Minute()

	if newEvents.Day < 0 || newEvents.Day > 6 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Day"}
	}

	if !util.IsValidConstant(newEvents.Type, constants.EventTypeMap) {
		return &errmsg.ErrFieldIsWrong{FieldName: "Event Type"}
	}

	parsedClassroomID, err := serviceutil.GetUUIDFromStringWithValidation("Classroom ID", &newEvents.ClassroomID)
	if err != nil {
		return err
	}

	// Process Repeat Begin and End Time
	repeatBeginTime, err := util.EventDateParse(newEvents.RepeatBeginDate)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Repeat Begin Date"}
	}

	repeatEndTime, err := util.EventDateParse(newEvents.RepeatEndDate)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Repeat End Date"}
	}

	if !repeatEndTime.After(*repeatBeginTime) {
		fmt.Println("repeatBeginTime", repeatBeginTime)
		fmt.Println("repeatEndTime", repeatEndTime)

		return &errmsg.ErrFieldIsWrong{FieldName: "End Date"}
	}

	weekCount := util.CountWeeksBetween(*repeatBeginTime, *repeatEndTime) + 1
	if weekCount < 1 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Week"}
	}

	var newEventsReq []model.Event
	startOfWeek := util.StartOfWeek(*repeatBeginTime)
	for weekIndex := 0; weekIndex < weekCount; weekIndex++ {
		correctDay := startOfWeek.AddDate(0, 0, (weekIndex*7)+newEvents.Day)
		correctBeginDate := correctDay.Add(time.Hour*time.Duration(beginHour) + time.Minute*time.Duration(beginMinute))
		correctEndDate := correctDay.Add(time.Hour*time.Duration(endHour) + time.Minute*time.Duration(endMinute))

		newEventReq := model.Event{
			BeginDate:   correctBeginDate,
			EndDate:     correctEndDate,
			Type:        newEvents.Type,
			ClassroomID: *parsedClassroomID,
			Description: newEvents.Description,
		}

		if newEvents.Type == constants.EVENT_TYPE_NORMAL {
			newEventReq.Title = newEvents.Title
		} else if newEvents.Type == constants.EVENT_TYPE_SUBJECT {
			parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", newEvents.SubjectID)
			if err != nil {
				return err
			}

			gotRelation, err := s.repo.RelationClassroomSubject().GetByClassroomIDAndSubjectID(parsedClassroomID, parsedSubjectID)
			if err != nil {
				return &errmsg.ErrInternal{Err: err}
			}

			newEventReq.Description = newEvents.Description
			newEventReq.RelationClassroomSubjectID = &gotRelation.ID
		}

		newEventsReq = append(newEventsReq, newEventReq)
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		for _, req := range newEventsReq {
			if err := s.repo.Event().CreateOne(tx, req); err != nil {
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

func (s *impService) CreateManyRepeated(c context.Context, manyNewEvents []rq.CreateRepeatWeeklyEventRequest) error {
	if len(manyNewEvents) < 1 {
		return &errmsg.ErrIsEmpty{FieldName: "Request"}
	}

	for _, newEvents := range manyNewEvents {
		if err := s.CreateRepeated(c, newEvents); err != nil {
			fmt.Println(newEvents)
			return err
		}
	}

	return nil
}

func (s *impService) EditOne(c context.Context, newEvent rq.EditEventRequest) error {
	timeBegin, err := util.EventDateParse(newEvent.BeginDate)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "Begin Date"}
	}

	timeEnd, err := util.EventDateParse(newEvent.EndDate)
	if err != nil {
		return &errmsg.ErrFieldIsWrong{FieldName: "End Date"}
	}

	if !timeEnd.After(*timeBegin) {
		return &errmsg.ErrFieldIsWrong{FieldName: "End Date"}
	}

	parsedEventID, err := serviceutil.GetUUIDFromStringWithValidation("Event ID", &newEvent.ID)
	if err != nil {
		return err
	}

	newEventReq := model.Event{
		Base:        model.Base{ID: *parsedEventID},
		BeginDate:   *timeBegin,
		EndDate:     *timeEnd,
		Title:       newEvent.Title,
		Description: newEvent.Description,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Event().UpdateOne(tx, newEventReq); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) DeleteOne(c context.Context, eventID string) error {
	parsedEventID, err := serviceutil.GetUUIDFromStringWithValidation("Event ID", &eventID)
	if err != nil {
		return err
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Event().DeleteOne(tx, *parsedEventID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
