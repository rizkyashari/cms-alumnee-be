package service

import (
	"github.com/fadhln/lms-be/repo"
	academicyear "github.com/fadhln/lms-be/service/academic_year"
	account "github.com/fadhln/lms-be/service/account"
	auth "github.com/fadhln/lms-be/service/auth"
	classroom "github.com/fadhln/lms-be/service/classroom"
	event "github.com/fadhln/lms-be/service/event"
	feedback "github.com/fadhln/lms-be/service/feedback"
	masscreate "github.com/fadhln/lms-be/service/mass_create"
	midtrans "github.com/fadhln/lms-be/service/midtrans"
	paymentlink "github.com/fadhln/lms-be/service/payment_link"
	rewardpunishment "github.com/fadhln/lms-be/service/reward_punishment"
	school "github.com/fadhln/lms-be/service/school"
	student "github.com/fadhln/lms-be/service/student"
	subject "github.com/fadhln/lms-be/service/subject"
	teacher "github.com/fadhln/lms-be/service/teacher"
)

type Service interface {
	Auth() auth.AuthService
	Account() account.AccountService
	AcademicYear() academicyear.AcademicYearService
	Classroom() classroom.ClassroomService
	Event() event.EventService
	MassCreate() masscreate.MassCreateService
	School() school.SchoolService
	Student() student.StudentService
	Subject() subject.SubjectService
	Teacher() teacher.TeacherService
	RewardPunishment() rewardpunishment.RewardPunishmentService
	Feedback() feedback.FeedbackService
	PaymentLink() paymentlink.PaymentLinkService
	Midtrans() midtrans.SnapService
}

type impService struct {
	repo repo.Repository
}

func SetupService(r repo.Repository) Service {
	return &impService{
		repo: r,
	}
}

func (s *impService) Auth() auth.AuthService {
	return auth.Init(s.repo)
}

func (s *impService) Account() account.AccountService {
	return account.Init(s.repo)
}

func (s *impService) AcademicYear() academicyear.AcademicYearService {
	return academicyear.Init(s.repo)
}

func (s *impService) Classroom() classroom.ClassroomService {
	return classroom.Init(s.repo)
}

func (s *impService) Event() event.EventService {
	return event.Init(s.repo)
}

func (s *impService) MassCreate() masscreate.MassCreateService {
	return masscreate.Init(s.repo)
}

func (s *impService) School() school.SchoolService {
	return school.Init(s.repo)
}

func (s *impService) Student() student.StudentService {
	return student.Init(s.repo)
}

func (s *impService) Subject() subject.SubjectService {
	return subject.Init(s.repo)
}

func (s *impService) Teacher() teacher.TeacherService {
	return teacher.Init(s.repo)
}

func (s *impService) RewardPunishment() rewardpunishment.RewardPunishmentService {
	return rewardpunishment.Init(s.repo)
}

func (s *impService) Feedback() feedback.FeedbackService {
	return feedback.Init(s.repo)
}

func (s *impService) PaymentLink() paymentlink.PaymentLinkService {
	return paymentlink.NewPaymentLinkService()
}

func (s *impService) Midtrans() midtrans.SnapService {
	return midtrans.Init(s.repo)
}
