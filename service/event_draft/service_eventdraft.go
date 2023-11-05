package service_eventdraft

import (
	"context"
	"encoding/json"
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

type draftStruct struct {
	Event []rq.CreateSingleEventRequest
}

type EventDraftService interface {
	GetAll(c context.Context, params *rq.PaginationParams[model.EventDraft]) (*rs.PaginationResponse[any, rs.EventDraftResponse], error)
	CreateOrSaveOne(c context.Context, draftID *string, schoolID string, academicYearID string, studyHourUnit int, draft []rq.CreateSingleEventRequest) error
	DeleteOne(c context.Context, draftID string) error
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) EventDraftService {
	return &impService{
		repo: r,
	}
}

func (s *impService) GetAll(c context.Context, params *rq.PaginationParams[model.EventDraft]) (*rs.PaginationResponse[any, rs.EventDraftResponse], error) {
	checkParam := rq.PaginationParams[any]{
		Limit:     params.Limit,
		Page:      params.Page,
		SortBy:    params.SortBy,
		SortOrder: params.SortOrder,
	}
	if !(util.IsParamValid(&checkParam)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Parameter"}
	}

	gotDrafts, maxPage, rowCount, err := s.repo.EventDraft().GetAll(params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res := rs.PaginationResponse[any, rs.EventDraftResponse]{}
			return &res, nil
		}

		return nil, &errmsg.ErrInternal{Err: err}
	}

	if gotDrafts == nil {
		return &rs.PaginationResponse[any, rs.EventDraftResponse]{
			Data: []rs.EventDraftResponse{},
		}, nil
	}

	if len(*gotDrafts) < 1 {
		return &rs.PaginationResponse[any, rs.EventDraftResponse]{
			Data: []rs.EventDraftResponse{},
		}, nil
	}

	var response []rs.EventDraftResponse
	for _, draft := range *gotDrafts {
		var tempRes rs.EventDraftResponse
		err = copier.Copy(&tempRes, draft)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		var parsedDraft draftStruct

		err = json.Unmarshal([]byte(draft.Draft), &parsedDraft)
		if err != nil {
			return nil, err
		}

		tempRes.Draft = parsedDraft.Event
		response = append(response, tempRes)
	}

	res := rs.PaginationResponse[any, rs.EventDraftResponse]{
		MaxPage:         maxPage,
		RowCount:        rowCount,
		CurrentPage:     params.Page,
		AvailableFilter: nil,
		Data:            response,
	}

	return &res, nil
}

func (s *impService) CreateOrSaveOne(c context.Context, draftID *string, schoolID string, academicYearID string, studyHourUnit int, draft []rq.CreateSingleEventRequest) error {
	if studyHourUnit < 1 {
		return &errmsg.ErrFieldIsWrong{FieldName: "Study Hour Unit"}
	}

	parsedSchoolID, err := serviceutil.GetUUIDFromStringWithValidation("School ID", &schoolID)
	if err != nil {
		return err
	}

	parsedAcademicYearID, err := serviceutil.GetUUIDFromStringWithValidation("Academic Year ID", &academicYearID)
	if err != nil {
		return err
	}

	var parsedDraftID uuid.UUID
	if draftID != nil {
		tempParsedDraftID, err := serviceutil.GetUUIDFromStringWithValidation("Draft ID", draftID)
		if err != nil {
			return err
		}

		parsedDraftID = *tempParsedDraftID
	} else {
		parsedDraftID = uuid.New()
	}

	data, err := json.Marshal(&draftStruct{Event: draft})
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	newDraft := model.EventDraft{
		Base:           model.Base{ID: parsedDraftID},
		SchoolID:       *parsedSchoolID,
		AcademicYearID: *parsedAcademicYearID,
		StudyHourUnit:  studyHourUnit,
		Draft:          string(data),
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.EventDraft().CreateOrSaveOne(tx, newDraft); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) DeleteOne(c context.Context, draftID string) error {
	parsedDraftID, err := serviceutil.GetUUIDFromStringWithValidation("Draft ID", &draftID)
	if err != nil {
		return err
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.EventDraft().DeleteOne(tx, *parsedDraftID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}
