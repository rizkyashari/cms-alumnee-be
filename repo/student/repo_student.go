package repo_student

import (
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepo interface {
	GetDetailByAccountID(id uuid.UUID) (*model.Student, error)

	CreateOne(tx *gorm.DB, newStudent *model.Student) error

	UpdateOne(tx *gorm.DB, id uuid.UUID, newStudent *model.Student) error

	GetStudentsByClassroomID(classroomID uuid.UUID) ([]model.Student, error)
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) StudentRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetDetailByAccountID(id uuid.UUID) (*model.Student, error) {
	student := model.Student{
		StudentData: model.StudentData{
			MedicalHistoryData: &model.MedicalHistoryData{},
		},
	}
	if err := r.db.Where("account_id = ?", id).Preload("StudentData.StudentFamilyData").Preload("StudentData.AddressData").Preload("StudentData.MedicalHistoryData").Preload("StudentData.SelfDevelopmentData").Preload("StudentData.AcademicData").Preload("StudentData.SchoolTransferData").Preload("StudentData.StudentFatherData").Preload("StudentData.StudentMotherData").Preload("StudentData.StudentGuardianData").Find(&student).Error; err != nil {
		return nil, err
	}

	return &student, nil
}

func (r *impRepo) CreateOne(tx *gorm.DB, newStudent *model.Student) error {
	if err := tx.Create(newStudent).Error; err != nil {
		return err
	}

	return nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, id uuid.UUID, newStudent *model.Student) error {
	if newStudent == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Student"}
	}

	result := tx.Model(newStudent).Where("id = ?", id).Updates(newStudent)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) GetStudentsByClassroomID(classroomID uuid.UUID) ([]model.Student, error) {
	var students []model.Student
	if err := r.db.Preload("StudentData").Where("classroom_id = ?", classroomID).Find(&students).Error; err != nil {
		return nil, err
	}

	return students, nil
}
