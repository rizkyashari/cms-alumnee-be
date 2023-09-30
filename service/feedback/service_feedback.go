package service_feedback

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
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type FeedbackService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.Feedback]) (*rs.PaginationResponse[any, rs.FeedbackResponse], error)
	GetAllByStudentID(c context.Context, studentID string, params *rq.PaginationParams[model.Feedback]) (*rs.PaginationResponse[any, rs.FeedbackResponse], error)
	GetAllByTeacherID(c context.Context, teacherID string, params *rq.PaginationParams[model.Feedback]) (*rs.PaginationResponse[any, rs.FeedbackResponse], error)
	GetDetailByID(c context.Context, id string) (*rs.FeedbackResponse, error)
	CreateOne(c context.Context, newFeedback *rq.FeedbackRequest) error
	EditFeedbackForAdmin(c context.Context, newRewardPunishment *rq.FeedbackRequest) error
	EditOne(c context.Context, newRewardPunishment *rq.FeedbackRequest) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) FeedbackService {
	return &impService{
		repo: r,
	}
}

func (s *impService) CreateOne(c context.Context, newFeedback *rq.FeedbackRequest) error {

	if newFeedback.AcademicYearID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "AcademicYearID"}
	}

	if newFeedback.StudentID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "StudentID"}
	}

	if newFeedback.TeacherID == nil {
		return &errmsg.ErrIsEmpty{FieldName: "TeacherID"}
	}

	parsedAcademicYearID, err := serviceutil.GetUUIDFromStringWithValidation("Academic Year ID", newFeedback.AcademicYearID)
	if err != nil {
		return err
	}

	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", newFeedback.StudentID)
	if err != nil {
		return err
	}

	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", newFeedback.TeacherID)
	if err != nil {
		return err
	}

	newFeedbackQuestionID := uuid.New()
	newFeedbackID := uuid.New()

	var feedbackScores []model.FeedbackScore
	for _, scoreRequest := range *newFeedback.FeedbackScores {
		if scoreRequest.FeedbackQuestion.Question != "" {
			feedbackQuestion := model.FeedbackQuestion{
				Question: scoreRequest.FeedbackQuestion.Question,
			}
			feedbackScores = append(feedbackScores, model.FeedbackScore{
				FeedbackQuestionID: newFeedbackQuestionID,
				FeedbackID:         newFeedbackID,
				FeedbackQuestion:   feedbackQuestion,
			})
		}
	}

	feedback := model.Feedback{
		StudentID:      *parsedStudentID,
		AcademicYearID: *parsedAcademicYearID,
		TeacherID:      *parsedTeacherID,
		FeedbackScores: feedbackScores,
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Feedback().CreateOne(tx, &feedback); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) EditOne(c context.Context, newFeedback *rq.FeedbackRequest) error {

	parsedFeedbackID, err := serviceutil.GetUUIDFromStringWithValidation("Feedback ID", newFeedback.ID)
	if err != nil {
		return err
	}

	for _, scoreRequest := range *newFeedback.FeedbackScores {
		if newFeedback.ID != nil && scoreRequest.FeedbackQuestionID != nil {

			parsedScoreID, err := serviceutil.GetUUIDFromStringWithValidation("Feedback Score ID", scoreRequest.ID)
			if err != nil {
				return err
			}

			if err := s.repo.Feedback().UpdateFeedbackScoreValue(*parsedFeedbackID, *parsedScoreID, *scoreRequest.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *impService) EditFeedbackForAdmin(c context.Context, newFeedback *rq.FeedbackRequest) error {

	parsedFeedbackID, err := serviceutil.GetUUIDFromStringWithValidation("Feedback ID", newFeedback.ID)
	if err != nil {
		return err
	}

	parsedAcademicYearID, err := serviceutil.GetUUIDFromStringWithValidation("Academic Year ID", newFeedback.AcademicYearID)
	if err != nil {
		return err
	}

	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", newFeedback.StudentID)
	if err != nil {
		return err
	}

	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", newFeedback.TeacherID)
	if err != nil {
		return err
	}

	for _, scoreRequest := range *newFeedback.FeedbackScores {
		if scoreRequest.FeedbackQuestion != nil {

			parsedScoreID, err := serviceutil.GetUUIDFromStringWithValidation("Feedback Score ID", scoreRequest.ID)
			if err != nil {
				return err
			}

			question := scoreRequest.FeedbackQuestion.Question

			parsedQuestionID, err := serviceutil.GetUUIDFromStringWithValidation("Feedback Question ID", scoreRequest.FeedbackQuestionID)
			if err != nil {
				return err
			}

			feedback := model.Feedback{
				Base:           model.Base{ID: *parsedFeedbackID},
				StudentID:      *parsedStudentID,
				AcademicYearID: *parsedAcademicYearID,
				TeacherID:      *parsedTeacherID,
				FeedbackScores: []model.FeedbackScore{
					{
						Base: model.Base{ID: *parsedScoreID},
						FeedbackQuestion: model.FeedbackQuestion{
							Base:     model.Base{ID: *parsedQuestionID},
							Question: question,
						},
					},
				},
			}

			err = s.repo.Transaction(func(tx *gorm.DB) error {
				return s.repo.Feedback().UpdateOne(tx, &feedback)
			})

			if err != nil {
				return err
			}

		} else {
			return errors.New("feedback question is missing")
		}
	}

	return nil
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[model.Feedback]) (*rs.PaginationResponse[any, rs.FeedbackResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotFeedbacks, maxPage, err := s.repo.Feedback().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.FeedbackResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotFeedbacks == nil {
		return &rs.PaginationResponse[any, rs.FeedbackResponse]{
			Data: []rs.FeedbackResponse{},
		}, nil
	}

	if len(*gotFeedbacks) < 1 {
		return &rs.PaginationResponse[any, rs.FeedbackResponse]{
			Data: []rs.FeedbackResponse{},
		}, nil
	}

	var datares []rs.FeedbackResponse
	err = copier.Copy(&datares, gotFeedbacks)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	res := rs.PaginationResponse[any, rs.FeedbackResponse]{
		MaxPage:         maxPage,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            datares,
	}

	return &res, nil
}

func (s *impService) GetAllByStudentID(c context.Context, studentID string, params *rq.PaginationParams[model.Feedback]) (*rs.PaginationResponse[any, rs.FeedbackResponse], error) {
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &studentID)
	if err != nil {
		return nil, err
	}

	newParams := *params
	newParams.Data.StudentID = *parsedStudentID

	return s.GetAll(c, &newParams)
}

func (s *impService) GetAllByTeacherID(c context.Context, teacherID string, params *rq.PaginationParams[model.Feedback]) (*rs.PaginationResponse[any, rs.FeedbackResponse], error) {
	parsedTeacherID, err := serviceutil.GetUUIDFromStringWithValidation("Teacher ID", &teacherID)
	if err != nil {
		return nil, err
	}

	newParams := *params
	newParams.Data.TeacherID = *parsedTeacherID

	return s.GetAll(c, &newParams)
}

func (s *impService) GetDetailByID(c context.Context, id string) (*rs.FeedbackResponse, error) {
	parsedFeedbackID, err := serviceutil.GetUUIDFromStringWithValidation("Academic Year ID", &id)
	if err != nil {
		return nil, err
	}

	gotFeedback, err := s.repo.Feedback().GetByID(*parsedFeedbackID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	var res rs.FeedbackResponse
	err = copier.Copy(&res, gotFeedback)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &res, nil
}
