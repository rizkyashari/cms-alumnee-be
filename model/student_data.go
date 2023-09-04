package model

import (
	"time"

	"github.com/google/uuid"
)

type StudentData struct {
	Base
	StudentID         uuid.UUID
	Gender            *int
	BirthPlace        *string
	BirthDate         *time.Time
	PhoneNumber       *string
	Religion          *string
	Nationality       *int
	EthnicGroup       *string
	StudentFamilyData StudentFamilyData
	AddressData       AddressData
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
