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
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountRepo interface {
	SetOneRedis(ctx context.Context, key string, acc *model.Account) error
	ReadOneRedis(ctx context.Context, key string) (*model.Account, error)
	GetAllTeacher(params *rq.PaginationParams[model.Account]) (*[]model.Account, int, error)
	GetAllStudent(params *rq.PaginationParams[model.Account]) (*[]model.Account, int, error)
	CreateOne(tx *gorm.DB, newAccount *model.Account) error
	ReadOneByEmail(email string) (*model.Account, error)
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

func (r *impRepo) GetAllTeacher(params *rq.PaginationParams[model.Account]) (*[]model.Account, int, error) {
	var accounts []model.Account

	var preloadTeacherArgs []any
	if params.Data.Teacher != nil && params.Data.Teacher.SchoolID != uuid.Nil {
		schoolIdArgs := []any{"school_id = ?", params.Data.Teacher.SchoolID.String()}
		preloadTeacherArgs = append(preloadTeacherArgs, schoolIdArgs...)
	}

	chain := r.db.Preload("Teacher", preloadTeacherArgs...)

	var preloadTeacherDataArgs []any
	if params.Data.Teacher != nil && params.Data.Teacher.TeacherData.Gender != nil {
		genderArgs := []any{"gender = (?)", params.Data.Teacher.TeacherData.Gender}
		preloadTeacherDataArgs = append(preloadTeacherDataArgs, genderArgs...)
	}

	if params.Data.Teacher != nil && params.Data.Teacher.TeacherData.EmploymentStatus != nil {
		employmentStatusArgs := []any{"employment_status = (?)", params.Data.Teacher.TeacherData.EmploymentStatus}
		preloadTeacherDataArgs = append(preloadTeacherDataArgs, employmentStatusArgs...)
	}

	chain = chain.Preload("Teacher.TeacherData", preloadTeacherDataArgs...)

	chain = chain.Where(r.db.Where("account_type = ?", constants.ACCOUNT_TEACHER))

	if len(*params.Data.Name) >= 2 {
		chain = chain.Where(r.db.Where("name ILIKE " + `'%` + *params.Data.Name + `%'`))
	}

	maxPage := util.GetMaxPage(chain.Find(&accounts), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"name",
	}

	result := chain.Scopes(
		util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&accounts)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return &accounts, maxPage, nil
}

func (r *impRepo) GetAllStudent(params *rq.PaginationParams[model.Account]) (*[]model.Account, int, error) {
	var accounts []model.Account

	var preloadStudentArgs []any
	if params.Data.Student != nil && params.Data.Student.ClassroomID != nil && *params.Data.Student.ClassroomID != uuid.Nil {
		classromIdArgs := []any{"classroom_id = ?", params.Data.Student.ClassroomID.String()}
		preloadStudentArgs = append(preloadStudentArgs, classromIdArgs...)
	}

	chain := r.db.Preload("Student", preloadStudentArgs...)

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

	maxPage := util.GetMaxPage(chain.Find(&accounts), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"name",
	}

	result := chain.Scopes(
		util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).
		Find(&accounts)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return &accounts, maxPage, nil
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
