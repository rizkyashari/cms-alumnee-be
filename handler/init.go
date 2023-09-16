package handler

import (
	"github.com/fadhln/lms-be/database"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/service"
	"github.com/gin-gonic/gin"
)

func InitRouter(server *database.RepoServer) *gin.Engine {
	r := gin.Default()

	repository := repo.SetupRepo(server)
	service := service.SetupService(repository)
	handler := SetupHandler(service)

	h := handler

	r.Use(CORSMiddleware())
	r.POST("/login", h.Auth.Login)

	// Register for public is not exist
	// r.POST("/register", h.Auth.Register)

	r.GET("/academic_year", h.AcademicYear.GetAll)
	r.GET("/academic_year/:id", h.AcademicYear.GetDetailByID)

	r.GET("/classroom", h.Classroom.GetAll)
	r.GET("/classroom/:id", h.Classroom.GetDetailByID)

	r.GET("/school", h.School.GetAll)
	r.GET("/school/:id", h.School.GetDetailByID)

	r.GET("/student", h.Student.GetAll)
	r.GET("/student/:account_id", h.Student.GetDetailByAccountID)
	r.GET("/student/data/:student_id", h.Student.GetStudentDataByStudentID)

	r.GET("/teacher", h.Teacher.GetAll)
	r.GET("/teacher/:account_id", h.Teacher.GetDetailByAccountID)
	r.GET("/teacher/data/:teacher_id", h.Teacher.GetTeacherDataByTeacherID)

	r.GET("/subject", h.Subject.GetAll)
	r.GET("/subject/:id", h.Subject.GetDetailByID)
	r.GET("/subject/:id/component", h.Subject.GetAllSubjectComponentBySubjectID)
	r.GET("/subject-component/:subject_component_id", h.Subject.GetSubjectComponentDetailByID)

	r.Use(h.Auth.CheckAuth())
	r.GET("/account-detail", h.Auth.GetOwnAccountDetail)

	admin := r.Group("/4dm1n")
	{
		admin.Use(h.Auth.CheckAdmin())

		mass_create := admin.Group("/mass_create")
		{
			mass_create.GET("/", h.MassCreate.GetAll)
			mass_create.GET("/:id", h.MassCreate.GetDetailByID)
		}

		academic_year := admin.Group("/academic_year")
		{
			academic_year.POST("/", h.AcademicYear.CreateOne)
			academic_year.PATCH("/:id/year", h.AcademicYear.EditYear)
			academic_year.PATCH("/:id/status", h.AcademicYear.EditStatus)
		}

		classroom := admin.Group("/classroom")
		{
			classroom.POST("/", h.Classroom.CreateOne)
			classroom.POST("/bulk", h.Classroom.CreateMass)
			classroom.PATCH("/:id", h.Classroom.EditOne)
		}

		school := admin.Group("/school")
		{
			school.POST("/", h.School.CreateOne)
			school.PATCH("/:id", h.School.EditOne)
		}

		student := admin.Group("/student")
		{
			student.POST("/", h.Student.CreateOne)
			student.POST("/bulk", h.Student.CreateMass)
			student.PATCH("/:student_id", h.Student.EditOne)
			student.PATCH("/:student_id/family", h.Student.EditFamilyData)
			student.PATCH("/:student_id/address", h.Student.EditAddressData)
		}

		teacher := admin.Group("/teacher")
		{
			teacher.POST("/", h.Teacher.CreateOne)
			teacher.POST("/bulk", h.Teacher.CreateMass)
			teacher.PATCH("/:teacher_id", h.Teacher.EditOne)
		}

		subject := admin.Group("/subject")
		{
			subject.POST("/", h.Subject.CreateOne)
			subject.PATCH("/:id", h.Subject.EditOne)

			subject.POST("/component", h.Subject.CreateOneSubjectComponent)
			subject.PATCH("/component/:subject_component_id", h.Subject.EditOneSubjectComponent)
		}
	}

	student := r.Group("/s")
	{
		student.Use(h.Auth.CheckStudent())

		student.GET("/detail", h.Student.GetOwnDetail)
		student.GET("/data", h.Student.GetOwnStudentData)
		student.PATCH("/data", h.Student.EditOwnData)
		student.PATCH("/data/family", h.Student.EditOwnFamilyData)
		student.PATCH("/data/address", h.Student.EditOwnAddressData)

		student.GET("/subject", h.Subject.GetAllOwnStudent)
	}

	teacher := r.Group("/t")
	{
		teacher.Use(h.Auth.CheckTeacher())

		teacher.GET("/detail", h.Teacher.GetOwnDetail)
		teacher.GET("/data", h.Teacher.GetOwnTeacherData)
		teacher.PATCH("/data", h.Teacher.EditOwnData)

		subject := teacher.Group("/subject")
		{
			subject.GET("/", h.Subject.GetAllOwnTeacher)

			subject.PATCH("/:id", h.Subject.EditOneWithValidation)

			subject.POST("/component", h.Subject.CreateOneSubjectComponentWithValidation)
			subject.PATCH("/component/:subject_component_id", h.Subject.EditOneSubjectComponentWithValidation)
		}
	}

	return r
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH,DELETE,GET,PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
