package rq

type AcademicYearRequest struct {
	ID       *string `json:"id,omitempty"`
	Year     *string `json:"year,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}
