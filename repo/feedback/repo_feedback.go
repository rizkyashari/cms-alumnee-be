package feedback

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeedbackRepo interface {
	GetAll(params *rq.PaginationParams[model.Feedback]) (*[]model.Feedback, int, error)
	GetAllFeedbackQuestions(params *rq.PaginationParams[model.FeedbackQuestion]) (*[]model.FeedbackQuestion, int, error)
	GetByID(id uuid.UUID) (*model.Feedback, error)
	CreateOne(tx *gorm.DB, newFeedback *model.Feedback) error
	CreateOneFeedbackQuestion(tx *gorm.DB, newFeedbackQuestion *model.FeedbackQuestion) error
	UpdateOne(tx *gorm.DB, newFeedback *model.Feedback) error
	UpdateFeedbackQuestion(tx *gorm.DB, id uuid.UUID, feedbackQuestion *model.FeedbackQuestion) error
	UpdateFeedbackScoreValue(feedbackID, scoreID uuid.UUID, value int) error
	DeleteOne(tx *gorm.DB, id uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) FeedbackRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) CreateOne(tx *gorm.DB, newFeedback *model.Feedback) error {
	if err := tx.Create(newFeedback).Error; err != nil {
		return err
	}
	return nil
}

func (r *impRepo) CreateOneFeedbackQuestion(tx *gorm.DB, newFeedbackQuestion *model.FeedbackQuestion) error {
	if err := tx.Create(newFeedbackQuestion).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, id uuid.UUID) error {
	if id == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.Feedback{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) GetAll(params *rq.PaginationParams[model.Feedback]) (*[]model.Feedback, int, error) {
	var feedbacks []model.Feedback

	chain := r.db.Preload("FeedbackScores").Preload("FeedbackScores.FeedbackQuestion")

	if params.Data.AcademicYearID != uuid.Nil {
		chain = chain.Where(r.db.Where("academic_year_id = ?", params.Data.AcademicYearID.String()))
	}

	if params.Data.StudentID != uuid.Nil {
		chain = chain.Where(r.db.Where("student_id = ?", params.Data.StudentID.String()))
	}

	if params.Data.TeacherID != uuid.Nil {
		chain = chain.Where(r.db.Where("teacher_id = ?", params.Data.TeacherID.String()))
	}

	maxPage := util.GetMaxPage(chain.Find(&feedbacks), params.Limit)

	validColumnTeacherID := []string{
		"teacher_id",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnTeacherID)).Find(&feedbacks)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return &feedbacks, maxPage, nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, newFeedback *model.Feedback) error {

	if newFeedback == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Feedback"}
	}

	feedbackQuestionID := newFeedback.FeedbackScores[0].FeedbackQuestionID

	result := tx.Model(newFeedback).
		Where("id = ?", newFeedback.ID).
		Updates(newFeedback)

	if result.Error != nil {
		return result.Error
	}

	existingQuestion := model.FeedbackQuestion{}
	result = tx.Where("id = ?", feedbackQuestionID).First(&existingQuestion)

	if result.Error != nil {
		return result.Error
	}

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) UpdateFeedbackQuestion(tx *gorm.DB, id uuid.UUID, feedbackQuestion *model.FeedbackQuestion) error {
	if feedbackQuestion == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Question"}
	}

	result := tx.Model(feedbackQuestion).Where("id = ?", id).Updates(feedbackQuestion)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *impRepo) GetByID(id uuid.UUID) (*model.Feedback, error) {
	var feedback model.Feedback
	if err := r.db.Preload("FeedbackScores").Preload("FeedbackScores.FeedbackQuestion").First(&feedback, id).Error; err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *impRepo) UpdateFeedbackScoreValue(feedbackID, scoreID uuid.UUID, value int) error {
	var score model.FeedbackScore
	if err := r.db.First(&score, "feedback_id = ? AND id = ?", feedbackID, scoreID).Error; err != nil {
		return err
	}

	score.Value = value
	if err := r.db.Save(&score).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) GetAllFeedbackQuestions(params *rq.PaginationParams[model.FeedbackQuestion]) (*[]model.FeedbackQuestion, int, error) {
	var question []model.FeedbackQuestion

	chain := r.db

	if len(params.Data.Question) >= 2 {
		chain = chain.Where(r.db.Where("question ILIKE " + `'%` + params.Data.Question + `%'`))
	}

	maxPage := util.GetMaxPage(chain.Find(&question), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"question",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).Find(&question)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return &question, maxPage, nil
}
