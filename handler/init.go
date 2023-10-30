package handler

import (
	"github.com/fadhln/lms-be/database"
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/service"
	"github.com/gin-gonic/gin"
)

func InitRouter(server *database.RepoServer) *gin.Engine {
	engine := gin.Default()
	engine.RedirectTrailingSlash = false

	repository := repo.SetupRepo(server)
	service := service.SetupService(repository)
	handler := SetupHandler(service)

	h := handler

	r := engine.Group("/api")
	{
		r.Static("/file", "file")

		r.Use(CORSMiddleware())
		r.POST("/login", h.Auth.Login)

		// Register for public is not exist
		// r.POST("/register", h.Auth.Register)

		r.GET("/academic_year", h.AcademicYear.GetAll)
		r.GET("/academic_year/:id", h.AcademicYear.GetDetailByID)

		r.GET("/classroom", h.Classroom.GetAll)
		r.GET("/classroom/by-subject/:subject_id", h.Classroom.GetAllBySubjectID)
		r.GET("/classroom/notin-subject/:subject_id", h.Classroom.GetAllByNotInSubjectID)
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
		r.GET("/subject/by-classroom/:classroom_id", h.Subject.GetAllByClassroomID)
		r.GET("/subject/notin-classroom/:classroom_id", h.Subject.GetAllNotInClassroomID)
		r.GET("/subject/:id", h.Subject.GetDetailByID)
		r.GET("/subject/:id/component", h.Subject.GetAllSubjectComponentBySubjectID)
		r.GET("/subject-component/:subject_component_id", h.Subject.GetSubjectComponentDetailByID)

		r.Use(h.Auth.CheckAuth())
		r.GET("/account-detail", h.Auth.GetOwnAccountDetail)
		r.POST("/avatar", h.Account.UploadOwnAvatar)

		r.GET("/reward-punishment", h.RewardPunishment.GetAll)

		r.GET("/feedback", h.Feedback.GetAll)
		r.GET("/feedback/:id", h.Feedback.GetDetailByID)

		r.GET("/midtrans-frontend", h.Midtrans.GetMidtransFrontendCredentials)

		admin := r.Group("/4dm1n")
		{
			admin.Use(h.Auth.CheckAdmin())

			account := admin.Group("/account")
			{
				account.POST("/:account_id/avatar", h.Account.UploadAvatar)
			}

			admin.GET("/mass_create", h.MassCreate.GetAll)
			mass_create := admin.Group("/mass_create")
			{
				mass_create.GET("/:id", h.MassCreate.GetDetailByID)
			}

			admin.POST("/academic_year", h.AcademicYear.CreateOne)
			academic_year := admin.Group("/academic_year")
			{
				academic_year.PATCH("/:id/year", h.AcademicYear.EditYear)
				academic_year.PATCH("/:id/status", h.AcademicYear.EditStatus)
			}

			admin.POST("/classroom", h.Classroom.CreateOne)
			classroom := admin.Group("/classroom")
			{
				classroom.POST("/bulk", h.Classroom.CreateMass)
				classroom.PATCH("/:id", h.Classroom.EditOne)
				classroom.POST("/:id/subject", h.Classroom.AssignSubjectsToClassroom)
				classroom.POST("/:id/subject/delete", h.Classroom.RemoveSubjectsFromClassroom)
			}

			admin.POST("/event", h.Event.CreateOne)
			event := admin.Group("/event")
			{
				event.GET("/by-classroom/:classroom_id", h.Event.GetAllByClassroomID)
				event.GET("/by-teacher/:teacher_id", h.Event.GetAllByTeacherID)
				event.GET("/by-student/:student_id", h.Event.GetAllByStudentID)
				event.POST("/weekly", h.Event.CreateRepeated)
				event.POST("/weekly/many", h.Event.CreateManyRepeated)
				event.PATCH("/:id", h.Event.EditOne)
				event.DELETE("/:id", h.Event.DeleteOne)
			}

			admin.POST("/school", h.School.CreateOne)
			school := admin.Group("/school")
			{
				school.PATCH("/:id", h.School.EditOne)
			}

			admin.POST("/student", h.Student.CreateOne)
			student := admin.Group("/student")
			{
				student.POST("/bulk", h.Student.CreateMass)
				student.PATCH("/:student_id", h.Student.EditOne)
				student.PATCH("/:student_id/family", h.Student.EditFamilyData)
				student.PATCH("/:student_id/address", h.Student.EditAddressData)
			}

			admin.POST("/teacher", h.Teacher.CreateOne)
			teacher := admin.Group("/teacher")
			{
				teacher.POST("/bulk", h.Teacher.CreateMass)
				teacher.PATCH("/:teacher_id", h.Teacher.EditOne)
			}

			admin.POST("/subject", h.Subject.CreateOne)
			subject := admin.Group("/subject")
			{
				subject.POST("/bulk", h.Subject.CreateMass)
				subject.POST("/with-classroom/:classroom_id", h.Subject.CreateOneWithClassroomID)
				subject.PATCH("/:id", h.Subject.EditOne)
				subject.POST("/:id/classroom", h.Subject.AssignClassroomsToSubject)
				subject.POST("/:id/classroom/delete", h.Subject.RemoveClassroomsFromSubject)

				subject.POST("/component", h.Subject.CreateOneSubjectComponent)
				subject.PATCH("/component/:subject_component_id", h.Subject.EditOneSubjectComponent)

			}

			admin.GET("/feedback", h.Feedback.GetAll)
			admin.POST("/feedback", h.Feedback.CreateOne)
			feedback := admin.Group("/feedback")
			{
				feedback.POST("/multiple", h.Feedback.CreateMultiple)
				feedback.PATCH("/:id", h.Feedback.EditFeedback)
				feedback.GET("/question", h.Feedback.GetAllFeedbackQuestions)
				feedback.POST("/question", h.Feedback.CreateOneFeedbackQuestion)
				feedback.PATCH("/question/:id", h.Feedback.EditFeedbackQuestion)
			}

			admin.GET("/bill", h.Midtrans.GetAllBills)
			admin.POST("/bill", h.Midtrans.CreateOneBill)
			bill := admin.Group("/bill")
			{
				bill.GET("/:id", h.Midtrans.GetBillByID)
				bill.PATCH("/:id", h.Midtrans.EditOneBill)
				bill.POST("/multiple", h.Midtrans.CreateMultipleBill)
			}
			admin.GET("/midtrans", h.Midtrans.GetMidtransCredentials)
			admin.PATCH("/midtrans", h.Midtrans.SaveMidtransCredentials)
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

			student.GET("/reward-punishment", h.RewardPunishment.GetAllOwnStudent)

			student.PATCH("/feedback", h.Feedback.EditOne)
			student.GET("/feedback", h.Feedback.GetAllOwnStudent)
			student.GET("/payment-link", h.PaymentLink.GetPaymentLinksByEmail)

			student.POST("/transaction", h.Midtrans.CreateTransaction)
			student.GET("/bill", h.Midtrans.GetAllOwnBills)
		}

		teacher := r.Group("/t")
		{
			teacher.Use(h.Auth.CheckTeacher())

			teacher.GET("/detail", h.Teacher.GetOwnDetail)
			teacher.GET("/data", h.Teacher.GetOwnTeacherData)
			teacher.PATCH("/data", h.Teacher.EditOwnData)

			teacher.GET("/classroom", h.Classroom.GetAllByOwnTeacherID)
			teacher.GET("/classroom-subject", h.Teacher.GetOwnAllClassroomSubject)

			teacher.GET("/subject", h.Subject.GetAllOwnTeacher)
			subject := teacher.Group("/subject")
			{
				subject.PATCH("/:id", h.Subject.EditOneWithValidation)

				subject.POST("/component", h.Subject.CreateOneSubjectComponentWithValidation)
				subject.PATCH("/component/:subject_component_id", h.Subject.EditOneSubjectComponentWithValidation)
			}

			teacher.GET("/reward-punishment", h.RewardPunishment.GetAllRewardPunishmentForTeacher)
			teacher.POST("/reward-punishment", h.RewardPunishment.CreateOne)
			reward_punishment := teacher.Group("/reward-punishment")
			{
				reward_punishment.PATCH("/:id", h.RewardPunishment.EditOne)
			}

			teacher.GET("/classroom-student", h.Teacher.GetAllClassroomStudents)

			teacher.GET("/feedback", h.Feedback.GetAllOwnTeacher)
		}
	}

	return engine
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH,DELETE,GET,PUT,OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
