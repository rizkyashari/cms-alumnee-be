package model

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
