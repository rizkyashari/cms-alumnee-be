package repo

import (
	"database/sql"

	account "github.com/fadhln/lms-be/repo/account"
	admin "github.com/fadhln/lms-be/repo/admin"
	student "github.com/fadhln/lms-be/repo/student"
	teacher "github.com/fadhln/lms-be/repo/teacher"
	"gorm.io/gorm"
)

type Repository interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error

	Account() account.AccountRepo
	Admin() admin.AdminRepo
	Student() student.StudentRepo
	Teacher() teacher.TeacherRepo
}

type impRepo struct {
	DB *gorm.DB
}

func SetupRepo(db *gorm.DB) Repository {
	return &impRepo{
		DB: db,
	}
}

func (r *impRepo) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return r.DB.Transaction(fc, opts...)
}

func (r *impRepo) Account() account.AccountRepo {
	return account.Init(r.DB)
}

func (r *impRepo) Admin() admin.AdminRepo {
	return admin.Init(r.DB)
}

func (r *impRepo) Student() student.StudentRepo {
	return student.Init(r.DB)
}

func (r *impRepo) Teacher() teacher.TeacherRepo {
	return teacher.Init(r.DB)
}
