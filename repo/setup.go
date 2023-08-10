package repo

import "gorm.io/gorm"

type Repository interface{}

type impRepo struct {
	db *gorm.DB
}

func SetupRepo(db *gorm.DB) Repository {
	return &impRepo{
		db: db,
	}
}
