package repo_classroom

import (
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClassroomRepo interface {
	GetDetailByID(id uuid.UUID) (*model.Classroom, error)
	GetDetailByCode(code string) (*model.Classroom, error)
	GetAll(params *rq.PaginationParams[model.Classroom]) (*[]model.Classroom, int, error)
	CreateOne(tx *gorm.DB, newClassroom *model.Classroom) error
	CreateMass(tx *gorm.DB, newClassrooms *[]model.Classroom) error
	UpdateOne(tx *gorm.DB, id uuid.UUID, newClassroom *model.Classroom) error
	DeleteOne(tx *gorm.DB, id uuid.UUID) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) ClassroomRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetDetailByID(id uuid.UUID) (*model.Classroom, error) {
	var classroom model.Classroom

	chain := r.db.Preload("AcademicYear").Preload("Teacher").Preload("Teacher.TeacherData")
	if err := chain.Where("id = ?", id).Find(&classroom).Error; err != nil {
		return nil, err
	}

	return &classroom, nil
}

func (r *impRepo) GetDetailByCode(code string) (*model.Classroom, error) {
	var classroom model.Classroom

	chain := r.db.Preload("AcademicYear").Preload("Teacher").Preload("Teacher.TeacherData")
	if err := chain.Where("code = ?", code).First(&classroom).Error; err != nil {
		return nil, err
	}

	return &classroom, nil
}

func (r *impRepo) GetAll(params *rq.PaginationParams[model.Classroom]) (*[]model.Classroom, int, error) {
	var classrooms []model.Classroom

	chain := r.db.Preload("AcademicYear").Preload("Teacher").Preload("Teacher.TeacherData")

	if params.Data.TeacherID != uuid.Nil {
		chain = chain.Where(r.db.Where("teacher_id = ?", params.Data.TeacherID.String()))
	}

	if params.Data.SchoolID != uuid.Nil {
		chain = chain.Where(r.db.Where(`school_id = ?`, params.Data.SchoolID.String()))
	}

	if params.Data.AcademicYearID != uuid.Nil {
		chain = chain.Where(r.db.Where("academic_year_id = ?", params.Data.AcademicYearID.String()))
	}

	if len(params.Data.Name) >= 2 {
		chain = chain.Where(r.db.Where("name ILIKE " + `'%` + params.Data.Name + `%'`))
	}

	maxPage := util.GetMaxPage(chain.Find(&classrooms), params.Limit)

	validColumnName := []string{
		"created_at",
		"updated_at",
		"name",
	}

	result := chain.Scopes(util.Pagination(params.Limit, params.Page, params.SortBy, params.SortOrder, validColumnName)).Find(&classrooms)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return &classrooms, maxPage, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newClassroom *model.Classroom) error {
	if err := tx.Create(newClassroom).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) CreateMass(tx *gorm.DB, newClassrooms *[]model.Classroom) error {
	if err := tx.Create(newClassrooms).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, id uuid.UUID, newClassroom *model.Classroom) error {
	if newClassroom == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Classroom"}
	}

	result := tx.Model(newClassroom).Where("id = ?", id).Updates(newClassroom)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) DeleteOne(tx *gorm.DB, id uuid.UUID) error {
	if id == uuid.Nil {
		return &errmsg.ErrIsEmpty{FieldName: "id"}
	}

	result := tx.Delete(&model.Classroom{}, id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
