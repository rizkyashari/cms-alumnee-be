package repo_studentdata

import (
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentDataRepo interface {
	GetStudentDataByStudentID(id uuid.UUID) (*model.StudentData, error)
	UpdateOne(tx *gorm.DB, studentID uuid.UUID, newStudentData *model.StudentData) error
	UpdateOneStudentFamilyData(tx *gorm.DB, studentDataID uuid.UUID, newStudentFamilyData *model.StudentFamilyData) error
	UpdateOneAddressData(tx *gorm.DB, studentDataID uuid.UUID, newAddressData *model.AddressData) error
}

type impRepo struct {
	db *gorm.DB
}

func Init(db *gorm.DB) StudentDataRepo {
	return &impRepo{
		db: db,
	}
}

func (r *impRepo) GetStudentDataByStudentID(id uuid.UUID) (*model.StudentData, error) {
	var studentData model.StudentData
	if err := r.db.
		Preload("StudentFamilyData").
		Preload("AddressData").
		Where("student_id = ?", id).First(&studentData).Error; err != nil {
		return nil, err
	}

	return &studentData, nil
}

func (r *impRepo) UpdateOne(tx *gorm.DB, studentID uuid.UUID, newStudentData *model.StudentData) error {
	if newStudentData == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Student Data"}
	}

	result := tx.Model(newStudentData).Where("student_id = ?", studentID).Updates(newStudentData)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) UpdateOneStudentFamilyData(tx *gorm.DB, studentDataID uuid.UUID, newStudentFamilyData *model.StudentFamilyData) error {
	if newStudentFamilyData == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Family Data"}
	}

	result := tx.Model(newStudentFamilyData).Where("student_data_id = ?", studentDataID).Updates(newStudentFamilyData)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *impRepo) UpdateOneAddressData(tx *gorm.DB, studentDataID uuid.UUID, newAddressData *model.AddressData) error {
	if newAddressData == nil {
		return &errmsg.ErrIsEmpty{FieldName: "Address Data"}
	}

	result := tx.Model(newAddressData).Where("student_data_id = ?", studentDataID).Updates(newAddressData)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
