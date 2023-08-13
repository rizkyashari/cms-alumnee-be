package service_account

import (
	"strings"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/jinzhu/copier"
)

type AccountService interface {
	GetDetail(req *rq.EmailOnlyRequest) (*rs.AccountResponse, error)
	GetDetailWithAccType(req *rq.EmailAndAccTypeRequest) (*rs.AccountResponse, error)
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) AccountService {
	return &impService{
		repo: r,
	}
}

func (s *impService) GetDetail(req *rq.EmailOnlyRequest) (*rs.AccountResponse, error) {
	return s.GetDetailWithAccType(&rq.EmailAndAccTypeRequest{Email: req.Email, AccountType: -1})
}

func (s *impService) GetDetailWithAccType(body *rq.EmailAndAccTypeRequest) (*rs.AccountResponse, error) {
	if len(body.Email) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Email"}
	}

	email := strings.ToLower(body.Email)

	if !(util.IsEmailValid(email)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Email"}
	}

	gotAccount, err := s.repo.Account().ReadOneByEmail(email)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if (body.AccountType != -1) && (body.AccountType != gotAccount.AccountType) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Account Type"}
	}

	var res rs.AccountResponse

	switch gotAccount.AccountType {
	case model.ADMIN:
		break

	case model.STUDENT:
		gotStudent, err := s.repo.Student().GetDetailByAccountID(gotAccount.ID)
		if err != nil {
			return nil, err
		}

		var studentRes rs.StudentResponse
		err = copier.Copy(&studentRes, gotStudent)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		res = rs.AccountResponse{StudentData: &studentRes}

	case model.TEACHER:
		gotTeacher, err := s.repo.Teacher().GetDetailByAccountID(gotAccount.ID)
		if err != nil {
			return nil, err
		}

		var teachRes rs.TeacherResponse
		err = copier.Copy(&teachRes, gotTeacher)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		res = rs.AccountResponse{TeacherData: &teachRes}

	default:
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Account Type"}
	}

	err = copier.Copy(&res, gotAccount)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return nil, nil
}
