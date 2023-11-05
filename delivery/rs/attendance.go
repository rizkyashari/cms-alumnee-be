package rs

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
