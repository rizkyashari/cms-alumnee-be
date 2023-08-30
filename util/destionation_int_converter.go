package util

import "github.com/fadhln/lms-be/model"

func GetDestStr(destination int) string {
	destinationMap := map[int]string{
		model.DEST_CLASSROOM:   "Classroom",
		model.DEST_STUDENT:     "Student",
		model.DEST_TEACHER:     "Teacher",
		model.DEST_SUBJECT:     "Subject",
		model.DEST_SUBJECTCOMP: "Subject Component",
	}

	return destinationMap[destination]
}
