package repo_account

import (
	"github.com/fadhln/lms-be/model"
	"gorm.io/gorm"
)

type AccountRepo interface {
	CreateOne(newAccount *model.Account) (*model.Account, error)

	ReadOneByEmail(email string) (*model.Account, error)
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) AccountRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) ReadOneByEmail(email string) (*model.Account, error) {
	var account model.Account
	if err := r.db.Where("email = ?", email).Find(&account).Error; err != nil {
		return nil, err
	}
	if *account.Email == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &account, nil
}
