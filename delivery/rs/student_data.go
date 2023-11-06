package rs

import (
	"time"

	"github.com/google/uuid"
)

type StudentDataResponse struct {
	ID                  uuid.UUID                   `json:"id"`
	StudentID           uuid.UUID                   `json:"student_id"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	Gender              int                         `json:"gender"`
	BirthPlace          string                      `json:"birth_place"`
	BirthDate           time.Time                   `json:"birth_date"`
	PhoneNumber         string                      `json:"phone_number"`
	Religion            string                      `json:"religion"`
	Nationality         int                         `json:"nationality"`
	EthnicGroup         string                      `json:"ethnic_group"`
	FamilyData          StudentFamilyDataResponse   `json:"family_data"`
	AddressData         AddressDataResponse         `json:"address_data"`
	MedicalHistoryData  MedicalHistoryDataResponse  `json:"medical_history_data"`
	SelfDevelopmentData SelfDevelopmentDataResponse `json:"self_development_data"`
	AcademicData        AcademicDataResponse        `json:"academic_data"`
	SchoolTransferData  SchoolTransferDataResponse  `json:"school_transfer_data"`
	StudentFatherData   StudentFatherDataResponse   `json:"student_father_data"`
	StudentMotherData   StudentMotherDataResponse   `json:"student_mother_data"`
	StudentGuardianData StudentGuardianDataResponse `json:"student_guardian_data"`
}

type StudentFamilyDataResponse struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ChildNumber    int       `json:"child_number"`
	SiblingCount   int       `json:"sibling_coung"`
	ChildStatus    int       `json:"child_status"`
	SpokenLanguage string    `json:"spoken_language"`
}

type AddressDataResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FullAddress string    `json:"full_address"`
	HouseNumber string    `json:"house_number"`
	RT          int       `json:"rt"`
	RW          int       `json:"rw"`
	Village     string    `json:"village"`
	SubDistrict string    `json:"sub_district"`
	District    string    `json:"district"`
	Province    string    `json:"province"`
	PostalCode  int       `json:"postal_code"`
}

type MedicalHistoryDataResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	BloodGroup   string    `json:"blood_group"`
	BodyWeight   int       `json:"body_weight"`
	Disease      string    `json:"disease"`
	SpecialNeeds string    `json:"special_needs"`
}

type SelfDevelopmentDataResponse struct {
	ID                       uuid.UUID `json:"id"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
	MandatoryExtracurricular string    `json:"mandatory_extracurricular"`
	OptionalExtracurricular  string    `json:"optional_extracurricular"`
	QuranReadingLevel        string    `json:"quran_reading_level"`
	NonAcademicAchievement   string    `json:"nonacademic_achivement"`
}

type AcademicDataResponse struct {
	ID                       uuid.UUID `json:"id"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
	NationalExamNumber       string    `json:"national_exam_number"`
	OriginSchool             string    `json:"origin_school"`
	OriginSchoolType         string    `json:"origin_school_type"`
	SchoolStatus             string    `json:"school_status"`
	SchoolAccreditation      string    `json:"school_accreditation"`
	SchoolAddress            string    `json:"school_address"`
	NationalExamScore        int       `json:"national_exam_score"`
	NationalIslamicExamScore int       `json:"national_islamic_exam_score"`
	BankAccountNumber        string    `json:"bank_account_number"`
	BankDKIAccountNumber     string    `json:"bank_dki_account_number"`
}

type SchoolTransferDataResponse struct {
	ID              uuid.UUID `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Origin          string    `json:"origin"`
	Reason          string    `json:"reason"`
	AcceptedInClass string    `json:"accepted_in_class"`
}

type StudentFatherDataResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Fullname    string    `json:"fullname"`
	Existence   string    `json:"existence"`
	BirthPlace  string    `json:"birth_place"`
	BirthDate   time.Time `json:"birth_date"`
	Religion    string    `json:"religion"`
	Education   string    `json:"education"`
	Job         string    `json:"job"`
	Income      uint      `json:"income"`
	FullAddress string    `json:"full_address"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
}

type StudentMotherDataResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Fullname    string    `json:"fullname"`
	Existence   string    `json:"existence"`
	BirthPlace  string    `json:"birth_place"`
	BirthDate   time.Time `json:"birth_date"`
	Religion    string    `json:"religion"`
	Education   string    `json:"education"`
	Job         string    `json:"job"`
	Income      uint      `json:"income"`
	FullAddress string    `json:"full_address"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
}

type StudentGuardianDataResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Fullname    string    `json:"fullname"`
	Existence   string    `json:"existence"`
	BirthPlace  string    `json:"birth_place"`
	BirthDate   time.Time `json:"birth_date"`
	Religion    string    `json:"religion"`
	Education   string    `json:"education"`
	Job         string    `json:"job"`
	Income      uint      `json:"income"`
	FullAddress string    `json:"full_address"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
}
