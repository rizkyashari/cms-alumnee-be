package rq

type FeedbackRequest struct {
	ID             *string                 `json:"id,omitempty"`
	Title          string                  `json:"title"`
	AcademicYearID *string                 `json:"academic_year_id"`
	StudentID      *string                 `json:"student_id"`
	TeacherID      *string                 `json:"teacher_id"`
	FeedbackScores *[]FeedbackScoreRequest `json:"feedback_score"`
}

type FeedbackQuestionRequest struct {
	ID       *string `json:"id,omitempty"`
	Question string  `json:"question,omitempty"`
}

type FeedbackScoreRequest struct {
	ID                 *string                  `json:"id,omitempty"`
	FeedbackID         *string                  `json:"feedback_id,omitempty"`
	FeedbackQuestionID *string                  `json:"feedback_question_id,omitempty"`
	FeedbackQuestion   *FeedbackQuestionRequest `json:"feedback_question"`
	Value              *int                     `json:"value,omitempty"`
}

type IsTaughtByTeacherRequest struct {
	ID                *string `json:"id,omitempty"`
	IsTaughtByTeacher *bool   `json:"is_taught_by_teacher,omitempty"`
}
