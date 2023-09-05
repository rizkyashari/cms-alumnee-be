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

	r.GET("/teacher", h.Teacher.GetAll)
	r.GET("/school", h.School.GetAll)

	r.Use(h.Auth.CheckAuth())

	r.GET("/account-detail", h.Auth.GetOwnAccountDetail)

	r.GET("/academic_year", h.AcademicYear.GetAll)
	r.GET("/academic_year/:id", h.AcademicYear.GetDetailByID)
	r.GET("/school/:id", h.School.GetDetailByID)
	r.GET("/teacher/:account_id", h.Teacher.GetDetailByAccountID)
	r.GET("/teacher/data/:teacher_id", h.Teacher.GetTeacherDataByTeacherID)

	admin := r.Group("/4dm1n")
	{
		admin.Use(h.Auth.CheckAdmin())

		academic_year := admin.Group("/academic_year")
		{
			academic_year.POST("/", h.AcademicYear.CreateOne)
			academic_year.PATCH("/:id/year", h.AcademicYear.EditYear)
			academic_year.PATCH("/:id/status", h.AcademicYear.EditStatus)
		}

		school := admin.Group("/school")
		{
			school.POST("/", h.School.CreateOne)
			school.PATCH("/:id", h.School.EditOne)
		}

		teacher := admin.Group("/teacher")
		{
			teacher.POST("/", h.Teacher.CreateOne)
			teacher.POST("/bulk", h.Teacher.CreateMass)
			teacher.PATCH("/:id", h.Teacher.EditOne)
		}
	}

	teacher := r.Group("/teacher")
	{
		teacher.Use(h.Auth.CheckTeacher())

		teacher.GET("/detail", h.Teacher.GetOwnDetail)
		teacher.GET("/data", h.Teacher.GetOwnTeacherData)
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
