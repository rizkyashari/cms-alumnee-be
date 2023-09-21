package service_account

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/util"
	avatarutil "github.com/fadhln/lms-be/util/avatar_util"
	"github.com/fadhln/lms-be/util/errmsg"
	serviceutil "github.com/fadhln/lms-be/util/service_util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type AccountService interface {
	UploadAvatar(c *gin.Context, accountID string, requestFile *multipart.FileHeader) error
	GetDetail(context.Context, *rq.EmailOnlyRequest) (*rs.AccountResponse, error)
	GetDetailWithAccType(context.Context, *rq.EmailAndAccTypeRequest) (*rs.AccountResponse, error)
}

type impService struct {
	repo repo.Repository
}

func Init(r repo.Repository) AccountService {
	return &impService{
		repo: r,
	}
}

func (s *impService) UploadAvatar(c *gin.Context, accountID string, requestFile *multipart.FileHeader) error {
	parsedAccountID, err := serviceutil.GetUUIDFromStringWithValidation("Account ID", &accountID)
	if err != nil {
		return nil
	}

	gotAccount, err := s.repo.Account().ReadOneByID(*parsedAccountID)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	if gotAccount == nil {
		return gorm.ErrRecordNotFound
	}

	newID := uuid.New()

	requestFileName := "temp-" + newID.String() + requestFile.Filename
	err = c.SaveUploadedFile(requestFile, "./file/image/"+requestFileName)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}
	defer os.Remove("./file/image/" + requestFileName)

	gotProcessedAvatar, err := avatarutil.ProcessAvatar(requestFileName)
	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	err = s.repo.Transaction(func(tx *gorm.DB) error {
		newAccount := model.Account{
			Base:   model.Base{ID: gotAccount.ID},
			Avatar: gotProcessedAvatar,
		}

		if err := s.repo.Account().UpdateOne(tx, *parsedAccountID, &newAccount); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return &errmsg.ErrInternal{Err: err}
	}

	return nil
}

func (s *impService) GetDetail(c context.Context, req *rq.EmailOnlyRequest) (*rs.AccountResponse, error) {
	return s.GetDetailWithAccType(c, &rq.EmailAndAccTypeRequest{Email: req.Email, AccountType: -1})
}

func (s *impService) GetDetailWithAccType(c context.Context, body *rq.EmailAndAccTypeRequest) (*rs.AccountResponse, error) {
	if len(body.Email) <= 0 {
		return nil, &errmsg.ErrIsEmpty{FieldName: "Email"}
	}

	email := strings.ToLower(body.Email)

	if !(util.IsEmailValid(email)) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Email"}
	}

	var res rs.AccountResponse
	key := fmt.Sprintf("account:%s", email)

	gotAccount, _ := s.repo.Account().ReadOneRedis(c, key)
	if gotAccount != nil {
		err := copier.Copy(&res, gotAccount)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		return &res, nil
	}

	gotAccount, err := s.repo.Account().ReadOneByEmail(email)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	err = copier.Copy(&res, gotAccount)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if (body.AccountType != -1) && (body.AccountType != gotAccount.AccountType) {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Account Type"}
	}

	switch gotAccount.AccountType {
	case constants.ACCOUNT_ADMIN:
		break

	case constants.ACCOUNT_STUDENT:
		gotStudent, err := s.repo.Student().GetDetailByAccountID(gotAccount.ID)
		if err != nil {
			return nil, err
		}

		var studentRes rs.StudentResponse
		err = copier.Copy(&studentRes, gotStudent)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		gotAccount.Student = gotStudent
		res.Student = &studentRes

	case constants.ACCOUNT_TEACHER:
		gotTeacher, err := s.repo.Teacher().GetDetailByAccountID(gotAccount.ID)
		if err != nil {
			return nil, err
		}

		var teachRes rs.TeacherResponse
		err = copier.Copy(&teachRes, gotTeacher)
		if err != nil {
			return nil, &errmsg.ErrInternal{Err: err}
		}

		gotAccount.Teacher = gotTeacher
		res.Teacher = &teachRes

	default:
		return nil, &errmsg.ErrFieldIsWrong{FieldName: "Account Type"}
	}

	s.repo.Account().SetOneRedis(c, key, gotAccount)

	return &res, nil
}
