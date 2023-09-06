package repo_schedule

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScheduleRepo interface {
	GetDetailByID(id uuid.UUID) (*model.Schedule, error)
	GetAll(params *rq.ScheduleParams) (*[]model.Schedule, error)
	CreateOne(tx *gorm.DB, newSchedule *model.Schedule) error
	CreateMass(tx *gorm.DB, newSchedule *[]model.Schedule) error
	UpdateOne(tx *gorm.DB, id uuid.UUID, newSchedule *model.Schedule) error
	DeleteOne(tx *gorm.DB, id uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) ScheduleRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetDetailByID(id uuid.UUID) (*model.Schedule, error) {
	var schedule model.Schedule
	if err := r.db.Where("id = ?", id).Find(&schedule).Error; err != nil {
		return nil, err
	}

	return &schedule, nil
}

func (r *impRepo) GetAll(params *rq.ScheduleParams) (*[]model.Schedule, error) {
	var schedules []model.Schedule

	chain := r.db

	if params.Data.ClassroomID != uuid.Nil {
		chain = chain.Where(r.db.Where("classroom_id = ?", params.Data.ClassroomID.String()))
	}

	if params.Data.SubjectID != uuid.Nil {
		chain = chain.Where(r.db.Where("subject_id = ?", params.Data.SubjectID.String()))
	}

	chain.Where(r.db.Where("start >= ?", params.Start))
	chain.Where(r.db.Where("end <= ?", params.End))

	validColumnName := []string{
		"start",
	}

	result := chain.Scopes(
		util.Pagination(99, 1, "start", "ASC", validColumnName)).
		Find(&schedules)

	if result.Error != nil {
		return nil, result.Error
	}

	return &schedules, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, schedule *model.Schedule) error {
	if err := tx.Create(schedule).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) CreateMass(tx *gorm.DB, newSchedule *[]model.Schedule) error {
	if err := tx.Create(newSchedule).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, id uuid.UUID, newSchedule *model.Schedule) error {
	if newSchedule == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Schedule"}
	}

	result := tx.Model(newSchedule).Where("id = ?", id).Updates(newSchedule)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, id uuid.UUID) error {
	if id == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.Schedule{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
