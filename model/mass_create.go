package model

import "strconv"

type MassCreate struct {
	Base
	Status         int
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
