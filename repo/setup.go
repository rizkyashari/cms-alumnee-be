package repo

import (
	"database/sql"

	"github.com/fadhln/lms-be/database"
	acadmicyear "github.com/fadhln/lms-be/repo/academic_year"
	account "github.com/fadhln/lms-be/repo/account"
	admin "github.com/fadhln/lms-be/repo/admin"
	classroom "github.com/fadhln/lms-be/repo/classroom"
	masscreate "github.com/fadhln/lms-be/repo/mass_create"
	school "github.com/fadhln/lms-be/repo/school"
	student "github.com/fadhln/lms-be/repo/student"
	studentdata "github.com/fadhln/lms-be/repo/student_data"
	teacher "github.com/fadhln/lms-be/repo/teacher"
	teacherdata "github.com/fadhln/lms-be/repo/teacher_data"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Repository interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error

	Account() account.AccountRepo
	AcademicYear() acadmicyear.AcademicYearRepo
	Admin() admin.AdminRepo
	Classroom() classroom.ClassroomRepo
	MassCreate() masscreate.MassCreateRepo
	School() school.SchoolRepo
	Student() student.StudentRepo
	StudentData() studentdata.StudentDataRepo
	Teacher() teacher.TeacherRepo
	TeacherData() teacherdata.TeacherDataRepo
}

type impRepo struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func SetupRepo(server *database.RepoServer) Repository {
	return &impRepo{
		DB:    server.DB,
		Redis: server.Redis,
	}
}

func (r *impRepo) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return r.DB.Transaction(fc, opts...)
}

func (r *impRepo) Account() account.AccountRepo {
	return account.Init(r.DB, r.Redis)
}

func (r *impRepo) AcademicYear() acadmicyear.AcademicYearRepo {
	return acadmicyear.Init(r.DB)
}

func (r *impRepo) Admin() admin.AdminRepo {
	return admin.Init(r.DB)
}

func (r *impRepo) Classroom() classroom.ClassroomRepo {
	return classroom.Init(r.DB)
}

func (r *impRepo) MassCreate() masscreate.MassCreateRepo {
	return masscreate.Init(r.DB)
}

func (r *impRepo) School() school.SchoolRepo {
	return school.Init(r.DB)
}

func (r *impRepo) Student() student.StudentRepo {
	return student.Init(r.DB)
}

func (r *impRepo) StudentData() studentdata.StudentDataRepo {
	return studentdata.Init(r.DB)
}

func (r *impRepo) Teacher() teacher.TeacherRepo {
	return teacher.Init(r.DB)
}

func (r *impRepo) TeacherData() teacherdata.TeacherDataRepo {
	return teacherdata.Init(r.DB)
}
