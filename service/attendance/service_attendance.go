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
