package service

import (
	"github.com/fadhln/lms-be/repo"
	academicyear "github.com/fadhln/lms-be/service/academic_year"
	account "github.com/fadhln/lms-be/service/account"
	auth "github.com/fadhln/lms-be/service/auth"
	classroom "github.com/fadhln/lms-be/service/classroom"
	masscreate "github.com/fadhln/lms-be/service/mass_create"
	school "github.com/fadhln/lms-be/service/school"
	student "github.com/fadhln/lms-be/service/student"
	teacher "github.com/fadhln/lms-be/service/teacher"
)

type Service interface {
	Auth() auth.AuthService
	Account() account.AccountService
	AcademicYear() academicyear.AcademicYearService
	Classroom() classroom.ClassroomService
	MassCreate() masscreate.MassCreateService
	School() school.SchoolService
	Student() student.StudentService
	Teacher() teacher.TeacherService
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

func (s *impService) MassCreate() masscreate.MassCreateService {
	return masscreate.Init(s.repo)
}

func (s *impService) School() school.SchoolService {
	return school.Init(s.repo)
}

func (s *impService) Student() student.StudentService {
	return student.Init(s.repo)
}

func (s *impService) Teacher() teacher.TeacherService {
	return teacher.Init(s.repo)
}
