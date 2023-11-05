package model

import "github.com/google/uuid"

type Student struct {
	Base
	AccountID         uuid.UUID `gorm:"uniqueIndex"`
	StudentData       StudentData
	ClassroomID       *uuid.UUID
	Classroom         Classroom `gorm:"foreignKey:ClassroomID;references:ID"`
	Scores            []Score
	RewardPunishments []RewardPunishment
	Attendances       []Attendance
}

type CountStudent struct {
	SchoolID      uuid.UUID `json:"school_id"`
	SchoolName    string    `json:"school_name"`
	StudentAmount int       `json:"student_amount"`
}

type CountTotalStudent struct {
	TotalStudents int `json:"total_students"`
}

type CountPaidBill struct {
	SchoolID       uuid.UUID `json:"school_id"`
	SchoolName     string    `json:"school_name"`
	PaidBillAmount int       `json:"paid_bill_amount"`
}

type CountUnpaidBill struct {
	SchoolID         uuid.UUID `json:"school_id"`
	SchoolName       string    `json:"school_name"`
	UnpaidBillAmount int       `json:"unpaid_bill_amount"`
}
