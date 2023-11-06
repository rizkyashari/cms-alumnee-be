package rs

import "github.com/google/uuid"

type AttendanceResponse struct {
	ID        string  `json:"id"`
	StudentID string  `json:"student_id"`
	EventID   string  `json:"event_id"`
	Status    int     `json:"status"`
	Remarks   *string `json:"remarks,omitempty"`
}

type ManyAttendanceResponse struct {
	Statuses []AttendanceResponse `json:"statuses"`
}

type SubjectAttendance struct {
	SubjectID             uuid.UUID `json:"subject_id"`
	TotalEventCountToDate int       `json:"total_event_count_to_date"`
	AttendCount           int       `json:"attend_count"`
}

type AttendanceSummaryResponse struct {
	StudentID   uuid.UUID           `json:"student_id"`
	ClassroomID uuid.UUID           `json:"classroom_id"`
	Attendances []SubjectAttendance `json:"attendances"`
}
