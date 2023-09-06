package rq

type ClassroomRequest struct {
	ID             *string `json:"id,omitempty"`
	Name           *string `json:"name,omitempty"`
	Code           *string `json:"code,omitempty"`
	AcademicYearID *string `json:"academic_year_id,omitempty"`
	SchoolID       *string `json:"school_id,omitempty"`
	TeacherID      *string `json:"teacher_id,omitempty"`
	TeacherEmail   *string `json:"teacher_email,omitempty"`
}
