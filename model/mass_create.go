package model

import "strconv"

const (
	DEST_CLASSROOM int = iota
	DEST_STUDENT
	DEST_TEACHER
	DEST_SUBJECT
	DEST_SUBJECTCOMP
)

type MassCreate struct {
	Base
	Destination    int
	SuccessCount   int
	ErrorCount     int
	RequestFileUrl string
	ReportFileUrl  string
	ErrorMessages  string
}

type ErrorMsg struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

func GetErrorMsgHeader() []string {
	return []string{"Row", "Message"}
}

func GetErrorMsgRow(msg *ErrorMsg) []string {
	return []string{strconv.Itoa(msg.Row), msg.Message}
}
