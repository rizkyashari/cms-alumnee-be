package service_score

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util/errmsg"
	serviceutil "github.com/fadhln/lms-be/util/service_util"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type ScoreService interface {
	GetTotalScoreForSubjectIDAndStudentID(subjectID string, studentID string) (*rs.TotalScoreResponse, error)
	GetByStudentIDAndComponentID(studentID string, subjectComponentID string) (*rs.StudentScoreResponse, error)
	SaveForStudentID(newScores *rq.StudentScoreRequest) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) ScoreService {
	return &impService{
		repo: r,
	}
}

func (s *impService) GetByStudentIDAndComponentID(studentID string, subjectComponentID string) (*rs.StudentScoreResponse, error) {
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &studentID)
	if err != nil {
		return nil, err
	}

	parsedSubjectComponentID, err := serviceutil.GetUUIDFromStringWithValidation("Subject Component ID", &subjectComponentID)
	if err != nil {
		return nil, err
	}

	gotScores, err := s.repo.Score().GetByStudentIDAndSubjectComponentID(*parsedStudentID, *parsedSubjectComponentID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	totalPercentage := 0
	totalScore := 0.0
	var scoresRes []rs.ScoreResponse
	for _, gotScore := range *gotScores {
		var res rs.ScoreResponse
		err = copier.Copy(&res, gotScore)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}
		totalPercentage = totalPercentage + gotScore.Percentage
		totalScore = totalScore + gotScore.Value
		scoresRes = append(scoresRes, res)
	}

	totalFinalScore := 0.0
	for _, gotScore := range *gotScores {
		actualValue := gotScore.Value * (float64(gotScore.Percentage) / float64(totalPercentage))
		totalFinalScore = totalFinalScore + actualValue
	}

	res := rs.StudentScoreResponse{
		StudentID:          *parsedStudentID,
		SubjectComponentID: *parsedSubjectComponentID,
		Total:              totalFinalScore,
		Scores:             scoresRes,
	}

	return &res, nil
}

func (s *impService) SaveForStudentID(newScores *rq.StudentScoreRequest) error {
	if newScores == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Score"}
	}

	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &newScores.StudentID)
	if err != nil {
		return err
	}

	parsedSubjectComponentID, err := serviceutil.GetUUIDFromStringWithValidation("Subject Component ID", &newScores.SubjectComponentID)
	if err != nil {
		return err
	}

	totalPercentage := 0
	var newScoreReq []model.Score
	for _, newScore := range newScores.Scores {
		var req model.Score
		req.StudentID = *parsedStudentID
		req.SubjectComponentID = *parsedSubjectComponentID
		req.Description = newScore.Description

		if newScore.Value < 0 {
			return &errmsg.ErrFieldIsWrong{FieldName: "Score Value"}
		}
		req.Value = newScore.Value

		if newScore.Percentage < 0 {
			return &errmsg.ErrFieldIsWrong{FieldName: "Percentage"}
		}
		totalPercentage = totalPercentage + newScore.Percentage
		if totalPercentage > 100 {
			return &errmsg.ErrMaxAmount{Amount: 100}
		}
		req.Percentage = newScore.Percentage

		newScoreReq = append(newScoreReq, req)
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.Score().DeleteAllForStudentAndComponentID(tx, *parsedStudentID, *parsedSubjectComponentID); err != nil {
			return err
		}

		if err := s.repo.Score().SaveAll(tx, &newScoreReq); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) GetTotalScoreForSubjectIDAndStudentID(subjectID string, studentID string) (*rs.TotalScoreResponse, error) {
	parsedStudentID, err := serviceutil.GetUUIDFromStringWithValidation("Student ID", &studentID)
	if err != nil {
		return nil, err
	}

	parsedSubjectID, err := serviceutil.GetUUIDFromStringWithValidation("Subject ID", &subjectID)
	if err != nil {
		return nil, err
	}

	res := rs.TotalScoreResponse{
		StudentID: *parsedStudentID,
		SubjectID: *parsedSubjectID,
	}

	gotSubject, err := s.repo.Subject().GetDetailByID(*parsedSubjectID)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	totalScore := 0.0
	for _, component := range gotSubject.SubjectComponents {
		gotScore, err := s.GetByStudentIDAndComponentID(studentID, component.ID.String())
		if err != nil {
			return nil, err
		}

		if len(gotScore.Scores) < 1 {
			return &res, nil
		}

		totalScore = totalScore + (gotScore.Total * (float64(component.Percentage) / 100))
	}

	res.Total = &totalScore
	return &res, nil
}
