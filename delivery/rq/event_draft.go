package rq

type EventDraftRequest struct {
	ID             *string                    `json:"id,omitempty"`
	AcademicYearID string                     `json:"academic_year_id"`
	SchoolID       string                     `json:"school_id"`
	StudyHourUnit  int                        `json:"study_hour_unit"`
	Events         []CreateSingleEventRequest `json:"events"`
}
