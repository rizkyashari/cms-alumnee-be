package model

import (
	"time"

	"github.com/google/uuid"
)

type StudentData struct {
	Base
	StudentID              uuid.UUID
	Fullname               *string
	Nickname               *string
	Gender                 *int
	BirthPlace             *string
	BirthDate              *time.Time
	PhoneNumber            *string
	Religion               *string
	Nationality            *int
	EthnicGroup            *string
	OriginSchoolNumber     *string
	DistanceToSchool       *string
	TransportationToSchool *string
	Hobby                  *string
	Ideal                  *string
	StudentFamilyData      *StudentFamilyData
	AddressData            *AddressData
	MedicalHistoryData     *MedicalHistoryData
	SelfDevelopmentData    *SelfDevelopmentData
	AcademicData           *AcademicData
	SchoolTransferData     *SchoolTransferData
	StudentFatherData      *StudentFatherData
	StudentMotherData      *StudentMotherData
	StudentGuardianData    *StudentGuardianData
}

type StudentFamilyData struct {
	Base
	StudentDataID  uuid.UUID
	ChildNumber    *int // Anak Ke
	SiblingCount   *int // Jumlah Saudara
	ChildStatus    *int // Status Anak
	SpokenLanguage *string
}

type AddressData struct {
	Base
	StudentDataID uuid.UUID
	FullAddress   *string
	HouseNumber   *string
	RT            *int
	RW            *int
	Village       *int
	SubDistrict   *string
	District      *string
	Province      *string
	PostalCode    *int
}

type MedicalHistoryData struct {
	Base
	StudentDataID uuid.UUID
	BloodGroup    *string
	BodyWeight    *int
	Disease       *string
	SpecialNeeds  *string
}

type SelfDevelopmentData struct {
	Base
	StudentDataID            uuid.UUID
	MandatoryExtracurricular *string
	OptionalExtracurricular  *string
	QuranReadingLevel        *string
	NonAcademicAchievement   *string
}

type AcademicData struct {
	Base
	StudentDataID            uuid.UUID
	NationalExamNumber       *string
	OriginSchool             *string
	OriginSchoolType         *string
	SchoolStatus             *string
	SchoolAccreditation      *string
	SchoolAddress            *string
	NationalExamScore        *int
	NationalIslamicExamScore *int
	BankAccountNumber        *string
	BankDKIAccountNumber     *string
}

type SchoolTransferData struct {
	Base
	StudentDataID   uuid.UUID
	Origin          *string
	Reason          *string
	AcceptedInClass *string
}

type StudentFatherData struct {
	Base
	StudentDataID uuid.UUID
	Fullname      *string
	Existence     *string
	BirthPlace    *string
	BirthDate     *time.Time
	Religion      *string
	Education     *string
	Job           *string
	Income        *uint
	FullAddress   *string
	PhoneNumber   *string
	Email         *string
}

type StudentMotherData struct {
	Base
	StudentDataID uuid.UUID
	Fullname      *string
	Existence     *string
	BirthPlace    *string
	BirthDate     *time.Time
	Religion      *string
	Education     *string
	Job           *string
	Income        *uint
	FullAddress   *string
	PhoneNumber   *string
	Email         *string
}

type StudentGuardianData struct {
	Base
	StudentDataID uuid.UUID
	Fullname      *string
	Existence     *string
	BirthPlace    *string
	BirthDate     *time.Time
	Religion      *string
	Education     *string
	Job           *string
	Income        *uint
	FullAddress   *string
	PhoneNumber   *string
	Email         *string
}
