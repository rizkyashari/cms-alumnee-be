package repo_account

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountRepo interface {
	SetOneRedis(ctx context.Context, key string, acc *model.Account) error
	ReadOneRedis(ctx context.Context, key string) (*model.Account, error)
	GetAllTeacher(params *rq.PaginationParams[model.Account]) (*[]model.Account, int, int, error)
	GetAllStudent(params *rq.PaginationParams[model.Account]) (*[]model.Account, int, int, error)
	CreateOne(tx *gorm.DB, newAccount *model.Account) error
	ReadOneByEmail(email string) (*model.Account, error)
	ReadOneByID(id uuid.UUID) (*model.Account, error)
	UpdateOne(tx *gorm.DB, accountID uuid.UUID, newAccount *model.Account) error
}

type impRepo struct {
	db    *gorm.DB
	redis *redis.Client
}

func Init(db *gorm.DB, redis *redis.Client) AccountRepo {
	return &impRepo{
		db:    db,
		redis: redis,
	}
}

func (r *impRepo) SetOneRedis(ctx context.Context, key string, acc *model.Account) error {
	jwtDurationMin := os.Getenv("JWT_DURATION_MIN")
	data, err := json.Marshal(acc)
	if err != nil {
		return err
	}
	intDurationMin, err := strconv.Atoi(jwtDurationMin)
	if err != nil {
		return err
	}

	duration := time.Duration(intDurationMin) * time.Minute

	return r.redis.Set(ctx, key, string(data), duration).Err()
}

func (r *impRepo) ReadOneRedis(ctx context.Context, key string) (*model.Account, error) {
	res := r.redis.Get(ctx, key)
	if res.Err() != nil {
		return nil, res.Err()
	}

	value, err := res.Result()
	if err != nil {
		return nil, err
	}

	var gotAccount model.Account

	err = json.Unmarshal([]byte(value), &gotAccount)
	if err != nil {
		return nil, err
	}

	return &gotAccount, nil
}

func (r *impRepo) GetAllTeacher(params *rq.PaginationParams[model.Account]) (*[]model.Account, int, int, error) {
	var accounts []model.Account

	chain := r.db.Preload("Teacher.TeacherData")

	if params.Data.Teacher != nil && params.Data.Teacher.SchoolID != uuid.Nil {
		teacherChain := r.db.Debug().Model(&model.Teacher{}).Where("school_id IN (?)", params.Data.Teacher.SchoolID)

		var teacherDataArgs *model.TeacherData
		if params.Data.Teacher != nil && params.Data.Teacher.TeacherData.Gender != nil {
			teacherDataArgs.Gender = params.Data.Teacher.TeacherData.Gender
		}

		if params.Data.Teacher != nil && params.Data.Teacher.TeacherData.EmploymentStatus != nil {
			teacherDataArgs.EmploymentStatus = params.Data.Teacher.TeacherData.EmploymentStatus
		}

		if teacherDataArgs != nil {
			teacherChain = teacherChain.Where(r.db.Where("id IN (?)"),
				r.db.Debug().Model(&model.TeacherData{}).Where(teacherDataArgs).Select("teacher_id"))
		}

		chain = chain.Where(r.db.Where("id IN (?)", teacherChain.Select("account_id")))
	}

	chain = chain.Where(r.db.Where("account_type = ?", constants.ACCOUNT_TEACHER))

	if len(*params.Data.Name) >= 2 {
		chain = chain.Where(r.db.Where("name ILIKE " + `'%` + *params.Data.Name + `%'`))
	}

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&accounts), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"name",
	}

	result := chain.Scopes(
		util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&accounts)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &accounts, maxPage, rowCount, nil
}

func (r *impRepo) GetAllStudent(params *rq.PaginationParams[model.Account]) (*[]model.Account, int, int, error) {
	var accounts []model.Account

	var preloadStudentArgs []any
	chain := r.db.Model(&model.Account{}).Joins("LEFT JOIN students ON students.account_id = accounts.id")

	if params.Data.Student != nil && params.Data.Student.ClassroomID != nil && *params.Data.Student.ClassroomID != uuid.Nil {
		classroomID := params.Data.Student.ClassroomID.String()
		chain = chain.Preload("Student.StudentData", preloadStudentArgs...).
			Where("students.classroom_id = ?", classroomID).
			Where("account_type = ?", constants.ACCOUNT_STUDENT)
	} else {
		chain = chain.Preload("Student.StudentData", preloadStudentArgs...).
			Where("account_type = ?", constants.ACCOUNT_STUDENT)
	}

	var preloadStudentDataArgs []any
	if params.Data.Student != nil && params.Data.Student.StudentData.Gender != nil {
		genderArgs := []any{"gender = (?)", params.Data.Student.StudentData.Gender}
		preloadStudentDataArgs = append(preloadStudentDataArgs, genderArgs...)
	}

	chain = chain.Preload("Student.StudentData", preloadStudentDataArgs...)

	chain = chain.Where(r.db.Where("account_type = ?", constants.ACCOUNT_STUDENT))

	if params.Data.Name != nil && len(*params.Data.Name) >= 2 {
		chain = chain.Where(r.db.Where("name ILIKE " + `'%` + *params.Data.Name + `%'`))
	}

	maxPage, rowCount := util.GetMaxPageAndRowCount(chain.Find(&accounts), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"name",
	}

	result := chain.Scopes(
		util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&accounts)

	if result.Error != nil {
		return nil, 0, 0, result.Error
	}

	return &accounts, maxPage, rowCount, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newAccount *model.Account) error {
	if err := tx.Create(newAccount).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) ReadOneByEmail(email string) (*model.Account, error) {
	var account model.Account
	if err := r.db.Where("email = ?", email).First(&account).Error; err != nil {
		return nil, err
	}
	if account.Email == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &account, nil
}

func (r *impRepo) ReadOneByID(accountID uuid.UUID) (*model.Account, error) {
	var account model.Account
	if err := r.db.Where("id = ?", accountID).First(&account).Error; err != nil {
		return nil, err
	}
	if account.Email == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &account, nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, accountID uuid.UUID, newAccount *model.Account) error {
	if newAccount == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Account"}
	}

	result := tx.Model(newAccount).Where("id = ?", accountID).Updates(newAccount)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
