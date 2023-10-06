package service_rewardpunishment

import (
	"context"
	"errors"

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

type RewardPunishmentService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.RewardPunishment]) (*rs.PaginationResponse[any, rs.RewardPunishmentResponse], error)

	GetAllByStudentID(c context.Context, studentID string, params *rq.PaginationParams[model.RewardPunishment]) (*rs.PaginationResponse[any, rs.RewardPunishmentResponse], error)

	CreateOne(c context.Context, newRewardPunishment *rq.RewardPunishmentRequest) error

	EditOne(c context.Context, newRewardPunishment *rq.RewardPunishmentRequest) error

	GetAllRewardPunishmentForTeacher(c context.Context, teacherID string, params *rq.PaginationParams[model.RewardPunishment]) (*rs.PaginationResponse[any, rs.RewardPunishmentResponse], error)
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) RewardPunishmentService {
	return &impService{
		repo: r,
	}
}

func (s *impService) CreateOne(c context.Context, newRewardPunishment *rq.RewardPunishmentRequest) error {

	if newRewardPunishment.Point == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Point"}
	}

	if newRewardPunishment.Type == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Type"}
	}

	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", newRewardPunishment.StudentID)
	if err != nil {
		return err
	}

	rewardPunishment := model.RewardPunishment{
		StudentID:   *parsedStudentID,
		Point:       *newRewardPunishment.Point,
		Type:        *newRewardPunishment.Type,
		Description: newRewardPunishment.Description,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.RewardPunishment().CreateOne(tx, &rewardPunishment); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) EditOne(c context.Context, newRewardPunishment *rq.RewardPunishmentRequest) error {

	parsedRewardPunishmentID, err := serviceutil.GetUUIDFromStringWithValidation("Reward Punishment ID", newRewardPunishment.ID)
	if err != nil {
		return err
	}

	rewardPunishment := model.RewardPunishment{
		Base:        model.Base{ID: *parsedRewardPunishmentID},
		Point:       *newRewardPunishment.Point,
		Type:        *newRewardPunishment.Type,
		Description: newRewardPunishment.Description,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.RewardPunishment().UpdateOne(tx, &rewardPunishment); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[model.RewardPunishment]) (*rs.PaginationResponse[any, rs.RewardPunishmentResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotRewardPunishments, maxPage, rowCount, err := s.repo.RewardPunishment().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.RewardPunishmentResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotRewardPunishments == nil {
		return &rs.PaginationResponse[any, rs.RewardPunishmentResponse]{
			Data: []rs.RewardPunishmentResponse{},
		}, nil
	}

	if len(*gotRewardPunishments) < 1 {
		return &rs.PaginationResponse[any, rs.RewardPunishmentResponse]{
			Data: []rs.RewardPunishmentResponse{},
		}, nil
	}

	var datares []rs.RewardPunishmentResponse
	err = copier.Copy(&datares, gotRewardPunishments)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res := rs.PaginationResponse[any, rs.RewardPunishmentResponse]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            datares,
	}

	return &res, nil
}

func (s *impService) GetAllByStudentID(c context.Context, studentID string, params *rq.PaginationParams[model.RewardPunishment]) (*rs.PaginationResponse[any, rs.RewardPunishmentResponse], error) {
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &studentID)
	if err != nil {
		return nil, err
	}

	newParams := *params
	newParams.Data.StudentID = *parsedStudentID

	return s.GetAll(c, &newParams)
}

func (s *impService) GetAllRewardPunishmentForTeacher(c context.Context, teacherID string, params *rq.PaginationParams[model.RewardPunishment]) (*rs.PaginationResponse[any, rs.RewardPunishmentResponse], error) {
	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", &teacherID)
	if err != nil {
		return nil, err
	}

	classrooms, err := s.repo.Classroom().GetClassroomsByTeacherID(*parsedTeacherID)
	if err != nil {
		return nil, err
	}

	var rewardPunishments []rs.RewardPunishmentResponse
	var maxPage int
	var maxRow int

	for _, classroom := range classrooms {
		classroomStudents, err := s.repo.Student().GetStudentsByClassroomID(classroom.ID)
		if err != nil {
			return nil, err
		}

		for _, student := range classroomStudents {
			studentRewardPunishments, pageCount, rowCount, err := s.repo.RewardPunishment().GetRewardPunishmentsByStudentID(student.ID, params)
			if err != nil {
				return nil, err
			}

			maxRow += rowCount
			maxPage += pageCount

			for _, rp := range studentRewardPunishments {
				rewardPunishmentResponse := rs.RewardPunishmentResponse{
					ID:          rp.ID,
					CreatedAt:   rp.CreatedAt,
					UpdatedAt:   rp.UpdatedAt,
					StudentID:   rp.StudentID.String(),
					Type:        rp.Type,
					Point:       rp.Point,
					Description: *rp.Description,
				}
				rewardPunishments = append(rewardPunishments, rewardPunishmentResponse)
			}
		}
	}

	res := rs.PaginationResponse[any, rs.RewardPunishmentResponse]{
		MaxPage:     maxPage,
		RowCount:    maxRow,
		CurrentPage: params.Page,
		Data:        rewardPunishments,
	}

	return &res, nil
}
