package rq

type ClassroomRequest struct {
	ID             *string `json:"id,omitempty"`
	Name           *string `json:"name,omitempty"`
	AcademicYearID *string `json:"academic_year_id,omitempty"`
	TeacherID      *string `json:"teacher_id,omitempty"`
}
