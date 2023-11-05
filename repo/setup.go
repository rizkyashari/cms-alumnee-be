package repo

import (
	"database/sql"

	"github.com/fadhln/lms-be/database"
	acadmicyear "github.com/fadhln/lms-be/repo/academic_year"
	account "github.com/fadhln/lms-be/repo/account"
	admin "github.com/fadhln/lms-be/repo/admin"
	attendance "github.com/fadhln/lms-be/repo/attendance"
	classroom "github.com/fadhln/lms-be/repo/classroom"
	event "github.com/fadhln/lms-be/repo/event"
	eventdraft "github.com/fadhln/lms-be/repo/event_draft"
	"github.com/fadhln/lms-be/repo/feedback"
	log "github.com/fadhln/lms-be/repo/log"
	masscreate "github.com/fadhln/lms-be/repo/mass_create"
	midtrans "github.com/fadhln/lms-be/repo/midtrans"
	relationclassroomsubject "github.com/fadhln/lms-be/repo/relation_classroom_subject"
	rewardpunishment "github.com/fadhln/lms-be/repo/reward_punishment"
	school "github.com/fadhln/lms-be/repo/school"
	score "github.com/fadhln/lms-be/repo/score"
	student "github.com/fadhln/lms-be/repo/student"
	studentdata "github.com/fadhln/lms-be/repo/student_data"
	subject "github.com/fadhln/lms-be/repo/subject"
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
	Attendance() attendance.AttendanceRepo
	Classroom() classroom.ClassroomRepo
	Event() event.EventRepo
	EventDraft() eventdraft.EventDraftRepo
	MassCreate() masscreate.MassCreateRepo
	RelationClassroomSubject() relationclassroomsubject.RelationClassroomSubjectRepo
	School() school.SchoolRepo
	Score() score.ScoreRepo
	Student() student.StudentRepo
	StudentData() studentdata.StudentDataRepo
	Subject() subject.SubjectRepo
	Teacher() teacher.TeacherRepo
	TeacherData() teacherdata.TeacherDataRepo
	RewardPunishment() rewardpunishment.RewardPunishmentRepo
	Feedback() feedback.FeedbackRepo
	Midtrans() midtrans.MidtransRepo
	Log() log.LogRepo
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

func (r *impRepo) Attendance() attendance.AttendanceRepo {
	return attendance.Init(r.DB)
}

func (r *impRepo) Classroom() classroom.ClassroomRepo {
	return classroom.Init(r.DB)
}

func (r *impRepo) Event() event.EventRepo {
	return event.Init(r.DB)
}

func (r *impRepo) EventDraft() eventdraft.EventDraftRepo {
	return eventdraft.Init(r.DB)
}

func (r *impRepo) MassCreate() masscreate.MassCreateRepo {
	return masscreate.Init(r.DB)
}

func (r *impRepo) RelationClassroomSubject() relationclassroomsubject.RelationClassroomSubjectRepo {
	return relationclassroomsubject.Init(r.DB)
}

func (r *impRepo) School() school.SchoolRepo {
	return school.Init(r.DB)
}

func (r *impRepo) Score() score.ScoreRepo {
	return score.Init(r.DB)
}

func (r *impRepo) Student() student.StudentRepo {
	return student.Init(r.DB)
}

func (r *impRepo) StudentData() studentdata.StudentDataRepo {
	return studentdata.Init(r.DB)
}

func (r *impRepo) Subject() subject.SubjectRepo {
	return subject.Init(r.DB)
}

func (r *impRepo) Teacher() teacher.TeacherRepo {
	return teacher.Init(r.DB)
}

func (r *impRepo) TeacherData() teacherdata.TeacherDataRepo {
	return teacherdata.Init(r.DB)
}

func (r *impRepo) RewardPunishment() rewardpunishment.RewardPunishmentRepo {
	return rewardpunishment.Init(r.DB)
}

func (r *impRepo) Feedback() feedback.FeedbackRepo {
	return feedback.Init(r.DB)
}

func (r *impRepo) Midtrans() midtrans.MidtransRepo {
	return midtrans.Init(r.DB)
}

func (r *impRepo) Log() log.LogRepo {
	return log.Init(r.DB)
}
