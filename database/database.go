package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	repo_academicyear "github.com/fadhln/lms-be/repo/academic_year"
	repo_account "github.com/fadhln/lms-be/repo/account"
	repo_admin "github.com/fadhln/lms-be/repo/admin"
	repo_classroom "github.com/fadhln/lms-be/repo/classroom"
	"github.com/fadhln/lms-be/repo/feedback"
	repo_masscreate "github.com/fadhln/lms-be/repo/mass_create"
	repo_midtrans "github.com/fadhln/lms-be/repo/midtrans"
	repo_relationclassroomsubject "github.com/fadhln/lms-be/repo/relation_classroom_subject"
	rewardpunishment "github.com/fadhln/lms-be/repo/reward_punishment"
	repo_school "github.com/fadhln/lms-be/repo/school"
	repo_student "github.com/fadhln/lms-be/repo/student"
	repo_studentdata "github.com/fadhln/lms-be/repo/student_data"
	repo_subject "github.com/fadhln/lms-be/repo/subject"
	repo_teacher "github.com/fadhln/lms-be/repo/teacher"
	repo_teacherdata "github.com/fadhln/lms-be/repo/teacher_data"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type RepoServer struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// AcademicYear implements repo.Repository.
func (*RepoServer) AcademicYear() repo_academicyear.AcademicYearRepo {
	panic("unimplemented")
}

// Account implements repo.Repository.
func (*RepoServer) Account() repo_account.AccountRepo {
	panic("unimplemented")
}

// Admin implements repo.Repository.
func (*RepoServer) Admin() repo_admin.AdminRepo {
	panic("unimplemented")
}

// Classroom implements repo.Repository.
func (*RepoServer) Classroom() repo_classroom.ClassroomRepo {
	panic("unimplemented")
}

// Feedback implements repo.Repository.
func (*RepoServer) Feedback() feedback.FeedbackRepo {
	panic("unimplemented")
}

// MassCreate implements repo.Repository.
func (*RepoServer) MassCreate() repo_masscreate.MassCreateRepo {
	panic("unimplemented")
}

// Midtrans implements repo.Repository.
func (*RepoServer) Midtrans() repo_midtrans.MidtransRepo {
	panic("unimplemented")
}

// RelationClassroomSubject implements repo.Repository.
func (*RepoServer) RelationClassroomSubject() repo_relationclassroomsubject.RelationClassroomSubjectRepo {
	panic("unimplemented")
}

// RewardPunishment implements repo.Repository.
func (*RepoServer) RewardPunishment() rewardpunishment.RewardPunishmentRepo {
	panic("unimplemented")
}

// School implements repo.Repository.
func (*RepoServer) School() repo_school.SchoolRepo {
	panic("unimplemented")
}

// Student implements repo.Repository.
func (*RepoServer) Student() repo_student.StudentRepo {
	panic("unimplemented")
}

// StudentData implements repo.Repository.
func (*RepoServer) StudentData() repo_studentdata.StudentDataRepo {
	panic("unimplemented")
}

// Subject implements repo.Repository.
func (*RepoServer) Subject() repo_subject.SubjectRepo {
	panic("unimplemented")
}

// Teacher implements repo.Repository.
func (*RepoServer) Teacher() repo_teacher.TeacherRepo {
	panic("unimplemented")
}

// TeacherData implements repo.Repository.
func (*RepoServer) TeacherData() repo_teacherdata.TeacherDataRepo {
	panic("unimplemented")
}

// Transaction implements repo.Repository.
func (*RepoServer) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	panic("unimplemented")
}

func ConnectDb() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Info),
		SkipDefaultTransaction: true,
	})

	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(2)
	}

	log.Println("DB is connected")
	db.Logger = logger.Default.LogMode(logger.Info)

	var strapiDBExists bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'strapi')").Scan(&strapiDBExists).Error
	if err != nil {
		log.Fatal("Failed to check if database exists. \n", err)
		os.Exit(2)
	}

	if !strapiDBExists {
		err := db.Exec("CREATE DATABASE strapi").Error
		if err != nil {
			log.Fatal("Failed to create 'strapi' database. \n", err)
			os.Exit(2)
		}
		log.Println("Database 'strapi' created successfully.")
	}

	return db
}

func ConnectRedis() *redis.Client {
	config := redis.Options{
		Addr:     fmt.Sprintf("%s:6379", os.Getenv("REDIS_HOST")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}

	rdb := redis.NewClient(&config)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Failed to connect to redis. \n", err)
		os.Exit(2)
	}

	log.Println("Redis is connected")

	return rdb
}
