package rq

type StudentRegisterRequest struct {
	Account BaseRegisterRequest `json:"account"`
	Data    StudentRequest      `json:"data"`
}

type StudentRegisterWithDetailRequest struct {
	Account BaseRegisterRequest  `json:"account"`
	Data    StudentUpdateRequest `json:"data"`
}

type StudentRequest struct {
	ClassroomID   *string `json:"classroom_id"`
	ClassroomCode *string `json:"classroom_code,omitempty"`
}

type StudentUpdateRequest struct {
	StudentID              *string                          `json:"student_id,omitempty"`
	ClassroomID            *string                          `json:"classroom_id,omitempty"`
	ClassroomCode          *string                          `json:"classroom_code,omitempty"`
	Gender                 *int                             `json:"gender,omitempty"`
	BirthPlace             *string                          `json:"birth_place,omitempty"`
	BirthDate              *string                          `json:"birth_date,omitempty"`
	PhoneNumber            *string                          `json:"phone_number,omitempty"`
	Religion               *string                          `json:"religion,omitempty"`
	Nationality            *int                             `json:"nationality,omitempty"`
	EthnicGroup            *string                          `json:"ethnic_group,omitempty"`
	Nickname               *string                          `json:"nickname,omitempty"`
	OriginSchoolNumber     *string                          `json:"origin_school_number,omitempty"`
	DistanceToSchool       *string                          `json:"distance_to_school,omitempty"`
	TransportationToSchool *string                          `json:"transportation_to_school,omitempty"`
	Hobby                  *string                          `json:"hobby,omitempty"`
	Ideal                  *string                          `json:"ideal,omitempty"`
	FamilyData             StudentFamilyDataUpdateRequest   `json:"student_family_data,omitempty"`
	AddressData            AddressDataUpdateRequest         `json:"address_data,omitempty"`
	MedicalHistoryData     MedicalHistoryDataUpdateRequest  `json:"medical_history_data,omitempty"`
	SelfDevelopmentData    SelfDevelopmentDataUpdateRequest `json:"self_development_data,omitempty"`
	AcademicData           AcademicDataUpdateRequest        `json:"academic_data,omitempty"`
	SchoolTransferData     SchoolTransferDataUpdateRequest  `json:"school_transfer_data,omitempty"`
	StudentFatherData      StudentFatherDataUpdateRequest   `json:"student_father_data,omitempty"`
	StudentMotherData      StudentMotherDataUpdateRequest   `json:"student_mother_data,omitempty"`
	StudentGuardianData    StudentGuardianDataUpdateRequest `json:"student_guardian_data,omitempty"`
}

type StudentFamilyDataUpdateRequest struct {
	ChildNumber    *int    `json:"child_number,omitempty"`
	SiblingCount   *int    `json:"sibling_count,omitempty"`
	ChildStatus    *int    `json:"child_status,omitempty"`
	SpokenLanguage *string `json:"spoken_language,omitempty"`
}

type AddressDataUpdateRequest struct {
	FullAddress *string `json:"full_address,omitempty"`
	HouseNumber *string `json:"house_number,omitempty"`
	RT          *int    `json:"rt,omitempty"`
	RW          *int    `json:"rw,omitempty"`
	Village     *string `json:"village,omitempty"`
	SubDistrict *string `json:"sub_district,omitempty"`
	District    *string `json:"district,omitempty"`
	Province    *string `json:"province,omitempty"`
	PostalCode  *int    `json:"postal_code,omitempty"`
}

type MedicalHistoryDataUpdateRequest struct {
	BloodGroup   *string `json:"blood_group,omitempty"`
	BodyWeight   *int    `json:"body_group,omitempty"`
	Disease      *string `json:"disease,omitempty"`
	SpecialNeeds *string `json:"special_needs,omitempty"`
}

type SelfDevelopmentDataUpdateRequest struct {
	MandatoryExtracurricular *string `json:"mandatory_extracurricular,omitempty"`
	OptionalExtracurricular  *string `json:"optional_extracurricular,omitempty"`
	QuranReadingLevel        *string `json:"quran_reading_level,omitempty"`
	NonAcademicAchievement   *string `json:"nonacademic_achievement,omitempty"`
}

type AcademicDataUpdateRequest struct {
	NationalExamNumber       *string `json:"national_exam_number,omitempty"`
	OriginSchool             *string `json:"origin_school,omitempty"`
	OriginSchoolType         *string `json:"origin_school_type,omitempty"`
	SchoolStatus             *string `json:"school_type,omitempty"`
	SchoolAccreditation      *string `json:"school_accreditation,omitempty"`
	SchoolAddress            *string `json:"school_address,omitempty"`
	NationalExamScore        *int    `json:"national_exam_score,omitempty"`
	NationalIslamicExamScore *int    `json:"national_islamic_exam_score,omitempty"`
	BankAccountNumber        *string `json:"bank_account_number,omitempty"`
	BankDKIAccountNumber     *string `json:"bank_dki_account_number,omitempty"`
}

type SchoolTransferDataUpdateRequest struct {
	Origin          *string `json:"origin,omitempty"`
	Reason          *string `json:"reason,omitempty"`
	AcceptedInClass *string `json:"accepted_in_class,omitempty"`
}

type StudentFatherDataUpdateRequest struct {
	Fullname    *string `json:"fullname,omitempty"`
	Existence   *string `json:"existence,omitempty"`
	BirthPlace  *string `json:"birth_place,omitempty"`
	BirthDate   *string `json:"birth_date,omitempty"`
	Religion    *string `json:"religion,omitempty"`
	Education   *string `json:"education,omitempty"`
	Job         *string `json:"job,omitempty"`
	Income      *uint   `json:"income,omitempty"`
	FullAddress *string `json:"full_address,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Email       *string `json:"email,omitempty"`
}

type StudentMotherDataUpdateRequest struct {
	Fullname    *string `json:"fullname,omitempty"`
	Existence   *string `json:"existence,omitempty"`
	BirthPlace  *string `json:"birth_place,omitempty"`
	BirthDate   *string `json:"birth_date,omitempty"`
	Religion    *string `json:"religion,omitempty"`
	Education   *string `json:"education,omitempty"`
	Job         *string `json:"job,omitempty"`
	Income      *uint   `json:"income,omitempty"`
	FullAddress *string `json:"full_address,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Email       *string `json:"email,omitempty"`
}

type StudentGuardianDataUpdateRequest struct {
	Fullname    *string `json:"fullname,omitempty"`
	Existence   *string `json:"existence,omitempty"`
	BirthPlace  *string `json:"birth_place,omitempty"`
	BirthDate   *string `json:"birth_date,omitempty"`
	Religion    *string `json:"religion,omitempty"`
	Education   *string `json:"education,omitempty"`
	Job         *string `json:"job,omitempty"`
	Income      *uint   `json:"income,omitempty"`
	FullAddress *string `json:"full_address,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Email       *string `json:"email,omitempty"`
}
