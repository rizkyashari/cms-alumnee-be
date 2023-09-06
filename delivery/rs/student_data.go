package rs

import (
	"time"

	"github.com/google/uuid"
)

type StudentDataResponse struct {
	ID          uuid.UUID                 `json:"id"`
	StudentID   uuid.UUID                 `json:"student_id"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	Gender      int                       `json:"gender"`
	BirthPlace  string                    `json:"birth_place"`
	BirthDate   time.Time                 `json:"birth_date"`
	PhoneNumber string                    `json:"phone_number"`
	Religion    string                    `json:"religion"`
	Nationality int                       `json:"nationality"`
	EthnicGroup string                    `json:"ethnic_group"`
	FamilyData  StudentFamilyDataResponse `json:"family_data"`
	AddressData AddressDataResponse       `json:"address_data"`
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
	Village     int       `json:"village"`
	SubDistrict string    `json:"sub_district"`
	District    string    `json:"district"`
	Province    string    `json:"province"`
	PostalCode  int       `json:"postal_code"`
}
