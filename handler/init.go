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
	// engine.Use(middleware.RequestLogger())

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

		r.GET("/event/:id", h.Event.GetDetailByID)
		r.GET("/event/:id/attendance", h.Attendance.GetAllForEventID)

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

		r.GET("/feedback/title", h.Feedback.GetAllUniqueFeedbackTitles)

		r.GET("/midtrans-frontend", h.Midtrans.GetMidtransFrontendCredentials)

		r.GET("/score", h.Score.GetByStudentIDAndComponentID)

		r.POST("/log", h.Log.LogActivity)
		// r.Use(h.Log.LogActivity)

		admin := r.Group("/4dm1n")
		{
			admin.Use(h.Auth.CheckAdmin())

			account := admin.Group("/account")
			{
				account.POST("/:account_id/avatar", h.Account.UploadAvatar, h.Log.LogActivity)
			}

			admin.GET("/mass_create", h.MassCreate.GetAll)
			mass_create := admin.Group("/mass_create")
			{
				mass_create.GET("/:id", h.MassCreate.GetDetailByID)
			}

			admin.POST("/academic_year", h.AcademicYear.CreateOne, h.Log.LogActivity)
			academic_year := admin.Group("/academic_year")
			{
				academic_year.PATCH("/:id/year", h.AcademicYear.EditYear, h.Log.LogActivity)
				academic_year.PATCH("/:id/status", h.AcademicYear.EditStatus, h.Log.LogActivity)
			}

			admin.POST("/classroom", h.Classroom.CreateOne, h.Log.LogActivity)
			classroom := admin.Group("/classroom")
			{
				classroom.POST("/bulk", h.Classroom.CreateMass, h.Log.LogActivity)
				classroom.PATCH("/:id", h.Classroom.EditOne, h.Log.LogActivity)
				classroom.POST("/:id/subject", h.Classroom.AssignSubjectsToClassroom, h.Log.LogActivity)
				classroom.POST("/:id/subject/delete", h.Classroom.RemoveSubjectsFromClassroom, h.Log.LogActivity)
			}

			admin.POST("/attendance", h.Attendance.CreateMany, h.Log.LogActivity)
			admin.PATCH("/attendance", h.Attendance.EditOne, h.Log.LogActivity)

			admin.POST("/event", h.Event.CreateOne, h.Log.LogActivity)
			event := admin.Group("/event")
			{
				event.GET("/by-classroom/:classroom_id", h.Event.GetAllByClassroomID)
				event.GET("/by-teacher/:teacher_id", h.Event.GetAllByTeacherID)
				event.GET("/by-student/:student_id", h.Event.GetAllByStudentID)
				event.POST("/weekly", h.Event.CreateRepeated, h.Log.LogActivity)
				event.POST("/weekly/many", h.Event.CreateManyRepeated, h.Log.LogActivity)
				event.PATCH("/:id", h.Event.EditOne, h.Log.LogActivity)
				event.DELETE("/:id", h.Event.DeleteOne, h.Log.LogActivity)
			}

			admin.GET("/event-draft", h.EventDraft.GetAll)
			admin.POST("/event-draft", h.EventDraft.CreateOrSaveOne, h.Log.LogActivity)
			admin.DELETE("/event-draft/:id", h.EventDraft.DeleteOne, h.Log.LogActivity)

			admin.POST("/school", h.School.CreateOne, h.Log.LogActivity)
			school := admin.Group("/school")
			{
				school.PATCH("/:id", h.School.EditOne, h.Log.LogActivity)
			}

			admin.POST("/student", h.Student.CreateOne, h.Log.LogActivity)
			student := admin.Group("/student")
			{
				student.POST("/bulk", h.Student.CreateMass, h.Log.LogActivity)
				student.PATCH("/:student_id", h.Student.EditOne, h.Log.LogActivity)
				student.PATCH("/:student_id/family", h.Student.EditFamilyData, h.Log.LogActivity)
				student.PATCH("/:student_id/address", h.Student.EditAddressData, h.Log.LogActivity)
			}

			admin.POST("/teacher", h.Teacher.CreateOne)
			teacher := admin.Group("/teacher")
			{
				teacher.POST("/bulk", h.Teacher.CreateMass, h.Log.LogActivity)
				teacher.PATCH("/:teacher_id", h.Teacher.EditOne, h.Log.LogActivity)
			}

			admin.POST("/subject", h.Subject.CreateOne, h.Log.LogActivity)
			subject := admin.Group("/subject")
			{
				subject.POST("/bulk", h.Subject.CreateMass, h.Log.LogActivity)
				subject.POST("/with-classroom/:classroom_id", h.Subject.CreateOneWithClassroomID, h.Log.LogActivity)
				subject.PATCH("/:id", h.Subject.EditOne, h.Log.LogActivity)
				subject.POST("/:id/classroom", h.Subject.AssignClassroomsToSubject, h.Log.LogActivity)
				subject.POST("/:id/classroom/delete", h.Subject.RemoveClassroomsFromSubject, h.Log.LogActivity)

				subject.POST("/component", h.Subject.CreateOneSubjectComponent, h.Log.LogActivity)
				subject.PATCH("/component/:subject_component_id", h.Subject.EditOneSubjectComponent, h.Log.LogActivity)

			}

			admin.GET("/feedback", h.Feedback.GetAll)
			admin.POST("/feedback", h.Feedback.CreateOne, h.Log.LogActivity)
			feedback := admin.Group("/feedback")
			{
				feedback.POST("/multiple", h.Feedback.CreateMultiple, h.Log.LogActivity)
				feedback.PATCH("/:id", h.Feedback.EditFeedback, h.Log.LogActivity)
				feedback.GET("/question", h.Feedback.GetAllFeedbackQuestions)
				feedback.POST("/question", h.Feedback.CreateOneFeedbackQuestion, h.Log.LogActivity)
				feedback.PATCH("/question/:id", h.Feedback.EditFeedbackQuestion, h.Log.LogActivity)
			}

			admin.GET("/transaction", h.Midtrans.GetAllTransactions)
			admin.GET("/transaction/update-database-job", h.Midtrans.UpdateTransactionStatusAndBill)

			admin.GET("/bill", h.Midtrans.GetAllBills)
			admin.POST("/bill", h.Midtrans.CreateOneBill, h.Log.LogActivity)
			bill := admin.Group("/bill")
			{
				bill.GET("/:id", h.Midtrans.GetBillByID)
				bill.PATCH("/:id", h.Midtrans.EditOneBill, h.Log.LogActivity)
				bill.POST("/multiple", h.Midtrans.CreateMultipleBill, h.Log.LogActivity)
			}
			admin.GET("/midtrans", h.Midtrans.GetMidtransCredentials)
			admin.PATCH("/midtrans/:school_id", h.Midtrans.SaveMidtransCredentials)

			admin.GET("/count-student", h.Student.GetStudentCount)
			admin.GET("/count-paid-bill", h.Student.GetStudentPaidBillCount)
			admin.GET("/count-unpaid-bill", h.Student.GetStudentUnpaidBillCount)

			admin.GET("/count-total-student", h.Student.CountStudents)
			admin.GET("/count-total-teacher", h.Teacher.CountTeachers)

			admin.GET("/log", h.Log.GetActivityLogs)
			admin.GET("/total-score", h.Score.GetTotalScoreForSubjectIDAndStudentID)

		}

		student := r.Group("/s")
		{
			student.Use(h.Auth.CheckStudent())

			student.GET("/detail", h.Student.GetOwnDetail)
			student.GET("/data", h.Student.GetOwnStudentData)
			student.PATCH("/data", h.Student.EditOwnData, h.Log.LogActivity)
			student.PATCH("/data/family", h.Student.EditOwnFamilyData, h.Log.LogActivity)
			student.PATCH("/data/address", h.Student.EditOwnAddressData, h.Log.LogActivity)

			student.GET("/subject", h.Subject.GetAllOwnStudent)

			student.GET("/attendance", h.Attendance.GetForOwnStudentWithClassroomIDAndSubjectID)
			student.GET("/attendance/summary", h.Attendance.GetSummaryForStudent)

			student.GET("/reward-punishment", h.RewardPunishment.GetAllOwnStudent)

			student.PATCH("/feedback", h.Feedback.EditOne, h.Log.LogActivity)
			student.PATCH("/feedback/is-taught", h.Feedback.EditIsTaughtByTeacher, h.Log.LogActivity)
			student.GET("/feedback", h.Feedback.GetAllOwnStudent)
			student.GET("/payment-link", h.PaymentLink.GetPaymentLinksByEmail)

			student.POST("/transaction", h.Midtrans.CreateTransaction, h.Log.LogActivity)
			student.GET("/transaction/update-database-job", h.Midtrans.UpdateTransactionStatusAndBill)

			student.GET("/bill", h.Midtrans.GetAllOwnBills)

			student.GET("/event", h.Event.GetAllByOwnStudentID)

			student.GET("/total-score", h.Score.GetTotalScoreForSubjectIDAndOwnStudent)
		}

		teacher := r.Group("/t")
		{
			teacher.Use(h.Auth.CheckTeacher())

			teacher.GET("/detail", h.Teacher.GetOwnDetail)
			teacher.GET("/data", h.Teacher.GetOwnTeacherData)
			teacher.PATCH("/data", h.Teacher.EditOwnData, h.Log.LogActivity)

			teacher.GET("/classroom", h.Classroom.GetAllByOwnTeacherID)
			teacher.GET("/classroom-subject", h.Teacher.GetOwnAllClassroomSubject)

			teacher.GET("/subject", h.Subject.GetAllOwnTeacher)
			subject := teacher.Group("/subject")
			{
				subject.PATCH("/:id", h.Subject.EditOneWithValidation, h.Log.LogActivity)

				subject.POST("/component", h.Subject.CreateOneSubjectComponentWithValidation, h.Log.LogActivity)
				subject.PATCH("/component/:subject_component_id", h.Subject.EditOneSubjectComponentWithValidation, h.Log.LogActivity)
				subject.DELETE("/component/:id", h.Subject.DeleteOneComponent, h.Log.LogActivity)
			}

			teacher.GET("/reward-punishment", h.RewardPunishment.GetAllRewardPunishmentForTeacher)
			teacher.POST("/reward-punishment", h.RewardPunishment.CreateOne, h.Log.LogActivity)
			reward_punishment := teacher.Group("/reward-punishment")
			{
				reward_punishment.PATCH("/:id", h.RewardPunishment.EditOne, h.Log.LogActivity)
			}

			teacher.GET("/event", h.Event.GetAllByOwnTeacherID)
			teacher.POST("/score", h.Score.SaveForStudentID)

			teacher.POST("/attendance", h.Attendance.CreateMany)
			teacher.PATCH("/attendance", h.Attendance.EditOne)

			teacher.GET("/classroom-student", h.Teacher.GetAllClassroomStudents)

			teacher.GET("/feedback", h.Feedback.GetAllOwnTeacher)
			teacher.GET("/total-score", h.Score.GetTotalScoreForSubjectIDAndStudentID)

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
