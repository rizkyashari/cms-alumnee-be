package repo_attendance

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttendanceRepo interface {
	GetAll(params *rq.GetAllAttendanceParams) ([]model.Attendance, error)
	CreateOne(tx *gorm.DB, newAttendance *model.Attendance) error
	UpdateOne(tx *gorm.DB, newAttendance *model.Attendance) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) AttendanceRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetAll(params *rq.GetAllAttendanceParams) ([]model.Attendance, error) {
	var attendances []model.Attendance

	chain := r.db

	if params.EventID != nil && *params.EventID != uuid.Nil {
		chain = chain.Where(r.db.Where("event_id = ?", params.EventID))
	}

	if params.StudentID != nil && *params.StudentID != uuid.Nil {
		chain = chain.Where(r.db.Where("student_id = ?`", params.StudentID))

		if (params.SubjectID != nil && *params.SubjectID != uuid.Nil) && (params.ClassroomID != nil && *params.ClassroomID != uuid.Nil) {
			chain = chain.Where(r.db.Where("event_id IN (?)",
				r.db.Debug().Model(&model.Event{}).Where("relation_classroom_subject_id IN (?)",
					r.db.Debug().Model(&model.RelationClassroomSubject{}).Where("subject_id = ? AND classroom_id = ?", params.SubjectID, params.ClassroomID).Select("id"),
				).Select("id"),
			))
		}
	}

	result := chain.Find(&attendances)

	if result.Error != nil {
		return nil, result.Error
	}

	return attendances, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newAttendance *model.Attendance) error {
	if err := tx.Create(newAttendance).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, newAttendance *model.Attendance) error {
	if newAttendance == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Attendance"}
	}

	result := tx.Model(newAttendance).Where("id = ?", newAttendance.ID).Updates(newAttendance)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
