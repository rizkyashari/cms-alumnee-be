package rq

type CreateSingleEventRequest struct {
	BeginDate   string  `json:"begin_date"`
	EndDate     string  `json:"end_date"`
	Type        string  `json:"type"`
	ClassroomID string  `json:"classroom_id"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	SubjectID   *string `json:"subject_id,omitempty"`
}

type CreateRepeatWeeklyEventRequest struct {
	RepeatBeginDate string  `json:"repeat_begin_date"`
	RepeatEndDate   string  `json:"repeat_end_date"`
	BeginTime       string  `json:"begin_time"`
	EndTime         string  `json:"end_time"`
	Day             int     `json:"day"`
	Type            string  `json:"type"`
	ClassroomID     string  `json:"classroom_id"`
	Title           *string `json:"title,omitempty"`
	Description     *string `json:"description,omitempty"`
	SubjectID       *string `json:"subject_id,omitempty"`
}

type CreateManyRepeatEventRequest struct {
	Events []CreateRepeatWeeklyEventRequest `json:"events"`
}

type EditEventRequest struct {
	ID          string  `json:"id"`
	BeginDate   string  `json:"begin_date"`
	EndDate     string  `json:"end_date"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}
