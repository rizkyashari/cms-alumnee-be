package repo

import (
	account "github.com/fadhln/lms-be/repo/account"
	"gorm.io/gorm"
)

type Repository interface {
	Account() account.AccountRepo
}

type impRepo struct {
	db *gorm.DB
}

func SetupRepo(db *gorm.DB) Repository {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) Account() account.AccountRepo {
	return account.Init(r.db)
}
