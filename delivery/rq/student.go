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
	StudentID     *string `json:"student_id,omitempty"`
	ClassroomID   *string `json:"classroom_id,omitempty"`
	ClassroomCode *string `json:"classroom_code,omitempty"`
	Gender        *int    `json:"gender,omitempty"`
	BirthPlace    *string `json:"birth_place,omitempty"`
	BirthDate     *string `json:"birth_date,omitempty"`
	PhoneNumber   *string `json:"phone_number,omitempty"`
	Religion      *string `json:"religion,omitempty"`
	Nationality   *int    `json:"nationality,omitempty"`
	EthnicGroup   *string `json:"ethnic_group,omitempty"`
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
	Village     *int    `json:"village,omitempty"`
	SubDistrict *string `json:"sub_district,omitempty"`
	District    *string `json:"district,omitempty"`
	Province    *string `json:"province,omitempty"`
	PostalCode  *int    `json:"postal_code,omitempty"`
}
