package handler

import (
	academicyear "github.com/fadhln/lms-be/handler/academic_year"
	account "github.com/fadhln/lms-be/handler/account"
	attendance "github.com/fadhln/lms-be/handler/attendance"
	auth "github.com/fadhln/lms-be/handler/auth"
	classroom "github.com/fadhln/lms-be/handler/classroom"
	event "github.com/fadhln/lms-be/handler/event"
	eventdraft "github.com/fadhln/lms-be/handler/event_draft"
	feedback "github.com/fadhln/lms-be/handler/feedback"
	log "github.com/fadhln/lms-be/handler/log"
	masscreate "github.com/fadhln/lms-be/handler/mass_create"
	midtrans "github.com/fadhln/lms-be/handler/midtrans"
	paymentlink "github.com/fadhln/lms-be/handler/payment_link"
	rewardpunishment "github.com/fadhln/lms-be/handler/reward_punishment"
	school "github.com/fadhln/lms-be/handler/school"
	score "github.com/fadhln/lms-be/handler/score"
	student "github.com/fadhln/lms-be/handler/student"
	subject "github.com/fadhln/lms-be/handler/subject"
	teacher "github.com/fadhln/lms-be/handler/teacher"
	"github.com/fadhln/lms-be/service"
)

type Handler struct {
	s service.Service

	Account          account.AccountHandler
	Attendance       attendance.AttendanceHandler
	Auth             auth.AuthHandler
	AcademicYear     academicyear.AcademicYearHandler
	Classroom        classroom.ClassroomHandler
	Event            event.EventHandler
	EventDraft       eventdraft.EventDraftHandler
	MassCreate       masscreate.MassCreateHandler
	School           school.SchoolHandler
	Score            score.ScoreHandler
	Student          student.StudentHandler
	Subject          subject.SubjectHandler
	Teacher          teacher.TeacherHandler
	RewardPunishment rewardpunishment.RewardPunishmentHandler
	Feedback         feedback.FeedbackHandler
	PaymentLink      paymentlink.PaymentLinkHandler
	Midtrans         midtrans.SnapHandler
	Log              log.LogHandler
}

func SetupHandler(s service.Service) *Handler {
	return &Handler{
		s:                s,
		Account:          account.Init(s),
		Attendance:       attendance.Init(s),
		Auth:             auth.Init(s),
		AcademicYear:     academicyear.Init(s),
		Classroom:        classroom.Init(s),
		Event:            event.Init(s),
		EventDraft:       eventdraft.Init(s),
		MassCreate:       masscreate.Init(s),
		School:           school.Init(s),
		Score:            score.Init(s),
		Student:          student.Init(s),
		Subject:          subject.Init(s),
		Teacher:          teacher.Init(s),
		RewardPunishment: rewardpunishment.Init(s),
		Feedback:         feedback.Init(s),
		PaymentLink:      paymentlink.Init(s),
		Midtrans:         midtrans.Init(s),
		Log:              log.Init(s),
	}
}
