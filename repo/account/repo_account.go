package repo_account

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/go-redis/redis/v8"
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

	chain := r.db.Preload("Teacher")

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

	chain := r.db.Preload("Student")

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

func (r *impRepo) CreateOne(tx *gorm.DB, newAccount *model.Account) error {
	if err := tx.Create(newAccount).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) ReadOneByEmail(email string) (*model.Account, error) {
	var account model.Account
	if err := r.db.Where("email = ?", email).Find(&account).Error; err != nil {
		return nil, err
	}
	if account.Email == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &account, nil
}
