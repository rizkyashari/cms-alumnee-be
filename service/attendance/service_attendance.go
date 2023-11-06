package service_attendance

import (
	"context"
	"errors"

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

type AttendanceService interface {
	GetSummaryForStudent(c context.Context, gotStudentAccount *rs.AccountResponse) (*rs.AttendanceSummaryResponse, error)
	GetAll(c context.Context, params *rq.GetAllAttendanceParams) (*rs.ManyAttendanceResponse, error)
	CreateMany(c context.Context, newAttendances *rq.CreateAttendanceRequest) error
	EditOne(c context.Context, newAttendance *rq.EditAttendanceRequest) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) AttendanceService {
	return &impService{
		repo: r,
	}
}

func (s *impService) GetSummaryForStudent(c context.Context, gotStudentAccount *rs.AccountResponse) (*rs.AttendanceSummaryResponse, error) {
	var res rs.AttendanceSummaryResponse
	tempAttendances := []rs.SubjectAttendance{}

	if gotStudentAccount == nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Account"}
	}

	res.StudentID = gotStudentAccount.Student.ID
	gotStudentDetail, err := s.repo.Student().GetDetailByAccountID(gotStudentAccount.ID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	gotClassroom, err := s.repo.Classroom().GetDetailByID(*gotStudentDetail.ClassroomID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}
	res.ClassroomID = gotClassroom.ID

	subjectParams := rq.PaginationParams[model.Subject]{
		Limit: 1000,
		Page:  1,
		Data:  model.Subject{},
	}
	subjectParams.Data.RelationClassroomSubjects = []model.RelationClassroomSubject{
		{ClassroomID: gotClassroom.ID},
	}

	gotSubjects, _, _, err := s.repo.Subject().GetAll(&subjectParams)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	for _, gotsubject := range *gotSubjects {
		gotAttendances, err := s.GetAll(c, &rq.GetAllAttendanceParams{
			StudentID:   &gotStudentAccount.Student.ID,
			ClassroomID: &gotClassroom.ID,
			SubjectID:   &gotsubject.ID,
		})
		if err != nil {
			return nil, err
		}

		tempAttendanceCount := 0
		for _, stat := range gotAttendances.Statuses {
			if stat.Status == constants.ATTENDANCE_STATUS_ATTEND {
				tempAttendanceCount++
			}
		}

		tempAttendances = append(tempAttendances, rs.SubjectAttendance{
			SubjectID:             gotsubject.ID,
			TotalEventCountToDate: len(gotAttendances.Statuses),
			AttendCount:           tempAttendanceCount,
		})
	}

	res.Attendances = tempAttendances
	return &res, nil
}

func (s *impService) GetAll(c context.Context, params *rq.GetAllAttendanceParams) (*rs.ManyAttendanceResponse, error) {
	if params == nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Parameter"}
	}
	gotAttendances, err := s.repo.Attendance().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &rs.ManyAttendanceResponse{Statuses: []rs.AttendanceResponse{}}, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if len(gotAttendances) < 1 {
		return &rs.ManyAttendanceResponse{Statuses: []rs.AttendanceResponse{}}, nil
	}

	var attendanceResponses []rs.AttendanceResponse
	for _, attendance := range gotAttendances {
		var attendanceRes rs.AttendanceResponse
		err = copier.Copy(&attendanceRes, attendance)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		attendanceResponses = append(attendanceResponses, attendanceRes)
	}

	return &rs.ManyAttendanceResponse{Statuses: attendanceResponses}, nil
}

func (s *impService) CreateMany(c context.Context, reqs *rq.CreateAttendanceRequest) error {
	if reqs == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Attendance"}
	}

	if len(reqs.Statuses) < 1 {
		return &errmsg.ErrIsEmpty{FieldName: "Attendance"}
	}

	parsedEventID, err := serviceutil.GetUUIDFromStringWithValidation("Event ID", &reqs.EventID)
	if err != nil {
		return err
	}

	var newAttendances []model.Attendance
	for _, attendanceReq := range reqs.Statuses {
		if !util.IsValidConstant(attendanceReq.Status, constants.AttendanceStatusMap) {
			return &errmsg.ErrFieldIsWrong{FieldName: "Status"}
		}

		var newReq model.Attendance
		err = copier.Copy(&newReq, attendanceReq)
		if err != nil {
			return &errmsg.ErrInternal{Err: err}
		}

		if newReq.Status != constants.ATTENDANCE_STATUS_REMARK {
			newReq.Remarks = nil
		}

		newReq.EventID = *parsedEventID

		newAttendances = append(newAttendances, newReq)
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		for _, newAtt := range newAttendances {
			if err := s.repo.Attendance().CreateOne(tx, &newAtt); err != nil {
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

func (s *impService) EditOne(c context.Context, newAttendance *rq.EditAttendanceRequest) error {
	if newAttendance == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Attendance"}
	}

	parsedAttendanceID, err := serviceutil.GetUUIDFromStringWithValidation("Attendance ID", &newAttendance.ID)
	if err != nil {
		return err
	}

	var req model.Attendance
	err = copier.Copy(&req, newAttendance)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}
	req.Base.ID = *parsedAttendanceID

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Attendance().UpdateOne(tx, &req); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
