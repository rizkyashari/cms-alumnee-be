package repo_event

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepo interface {
	GetAll(params rq.EventParams) ([]model.Event, error)
	CreateOne(tx *gorm.DB, newEvent model.Event) error
	UpdateOne(tx *gorm.DB, newEvent model.Event) error
	DeleteOne(tx *gorm.DB, eventID uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) EventRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetAll(params rq.EventParams) ([]model.Event, error) {
	var events []model.Event

	chain := r.db.Preload("Classroom").Preload("RelationClassroomSubject").Preload("RelationClassroomSubject.Subject")

	chain = chain.Where(r.db.Where("begin_date >= ? AND end_date <= ?", params.Begin, params.End))

	if params.Data != nil {
		// Get By ClassroomID
		if params.Data.ClassroomID != nil && *params.Data.ClassroomID != uuid.Nil {
			chain = chain.Where(r.db.Where("classroom_id = ?", params.Data.ClassroomID))
		}

		// Get By TeacherID
		if params.Data.TeacherID != nil && *params.Data.TeacherID != uuid.Nil {
			chain = chain.Where(r.db.Where("relation_classroom_subject_id IN (?)",
				r.db.Debug().Model(&model.RelationClassroomSubject{}).Where("subject_id IN (?)",
					r.db.Debug().Model(&model.Subject{}).
						Where(r.db.Where("teacher_id = ?", params.Data.TeacherID)).Select("id")).
					Select("id"),
			))
		}

		// Get By StudentID
		if params.Data.StudentID != nil && *params.Data.StudentID != uuid.Nil {
			chain = chain.Where(r.db.Where("classroom_id = ?",
				r.db.Debug().Model(&model.Student{}).Where("id = ?", params.Data.StudentID).Select("classroom_id"),
			))
		}
	}

	validColumnName := []string{
		"begin_date",
	}

	result := chain.Scopes(util.Pagination(100, 1, "begin_date", "asc", validColumnName)).Find(&events)

	if result.Error != nil {
		return []model.Event{}, result.Error
	}

	return events, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newEvent model.Event) error {
	if err := tx.Create(&newEvent).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, newEvent model.Event) error {
	result := tx.Model(&newEvent).Where("id = ?", newEvent.ID).Updates(&newEvent)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, eventID uuid.UUID) error {
	if eventID == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.Event{}, eventID)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
