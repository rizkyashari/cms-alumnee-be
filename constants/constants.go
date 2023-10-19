package constants

const (
	_ int = iota
	DEST_CLASSROOM
	DEST_STUDENT
	DEST_TEACHER
	DEST_SUBJECT
	DEST_SUBJECTCOMP
	DEST_BILL
)

var DestinationMap = map[int]bool{
	DEST_CLASSROOM:   true,
	DEST_STUDENT:     true,
	DEST_TEACHER:     true,
	DEST_SUBJECT:     true,
	DEST_SUBJECTCOMP: true,
	DEST_BILL:        true,
}

var DestinationMapString = map[int]string{
	DEST_CLASSROOM:   "Classroom",
	DEST_STUDENT:     "Student",
	DEST_TEACHER:     "Teacher",
	DEST_SUBJECT:     "Subject",
	DEST_SUBJECTCOMP: "Subject Component",
	DEST_BILL:        "Bill",
}

///

const (
	_ int = iota
	ACCOUNT_ADMIN
	ACCOUNT_STUDENT
	ACCOUNT_TEACHER
)

var AccountTypeMap = map[int]bool{
	ACCOUNT_ADMIN:   true,
	ACCOUNT_STUDENT: true,
	ACCOUNT_TEACHER: true,
}

///

const (
	_ int = iota
	GENDER_MALE
	GENDER_FEMALE
)

var GenderMap = map[int]bool{
	GENDER_MALE:   true,
	GENDER_FEMALE: true,
}

///

const (
	_ int = iota
	NATIONALITY_INDONESIAN
	NATIONALITY_FOREIGN
)

var NationalityMap = map[int]bool{
	NATIONALITY_INDONESIAN: true,
	NATIONALITY_FOREIGN:    true,
}

///

const (
	_ int = iota
	TEACHER_STATUS_PNS
	TEACHER_STATUS_NON_PNS
)

var TeacherStatusMap = map[int]bool{
	TEACHER_STATUS_PNS:     true,
	TEACHER_STATUS_NON_PNS: true,
}

///

const (
	_ int = iota
	STUDENT_CHILD_STATUS_BIOLOGICAL
	STUDENT_CHILD_STATUS_STEP
	STUDENT_CHILD_STATUS_ADOPTED
)

var StudentChildStatusMap = map[int]bool{
	STUDENT_CHILD_STATUS_BIOLOGICAL: true,
	STUDENT_CHILD_STATUS_STEP:       true,
	STUDENT_CHILD_STATUS_ADOPTED:    true,
}

///

const (
	_ int = iota
	MASS_CREATE_STATUS_PROCESSING
	MASS_CREATE_STATUS_DONE
	MASS_CREATE_STATUS_FAILED
)

var MassCreateStatusMap = map[int]bool{
	MASS_CREATE_STATUS_PROCESSING: true,
	MASS_CREATE_STATUS_DONE:       true,
	MASS_CREATE_STATUS_FAILED:     true,
}

///

const (
	_ int = iota
	TYPE_REWARD
	TYPE_PUNISHMENT
)

var RewardPunishmentTypeMap = map[int]bool{
	TYPE_REWARD:     true,
	TYPE_PUNISHMENT: true,
}

///

const (
	_ int = iota
	FEEDBACK_SCORE_STRONGLY_DISAGREE
	FEEDBACK_SCORE_DISAGREE
	FEEDBACK_SCORE_NEUTRAL
	FEEDBACK_SCORE_AGREE
	FEEDBACK_SCORE_STRONGLY_AGREE
)

var FeedbackScoreValueMap = map[int]bool{
	FEEDBACK_SCORE_STRONGLY_DISAGREE: true,
	FEEDBACK_SCORE_DISAGREE:          true,
	FEEDBACK_SCORE_NEUTRAL:           true,
	FEEDBACK_SCORE_AGREE:             true,
	FEEDBACK_SCORE_STRONGLY_AGREE:    true,
}

///

const (
	_ int = iota
	ATTENDANCE_STATUS_ATTEND
	ATTENDANCE_STATUS_ABSENT_PERMITTED
	ATTENDANCE_STATUS_ABSENT_SICK
	ATTENDANCE_STATUS_ABSENT_NO_REMARK
	ATTENDANCE_STATUS_OTHER
)

var AttendanceStatusMap = map[int]bool{
	ATTENDANCE_STATUS_ATTEND:           true,
	ATTENDANCE_STATUS_ABSENT_PERMITTED: true,
	ATTENDANCE_STATUS_ABSENT_SICK:      true,
	ATTENDANCE_STATUS_ABSENT_NO_REMARK: true,
	ATTENDANCE_STATUS_OTHER:            true,
}
