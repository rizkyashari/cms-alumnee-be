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
	r.POST("/register", h.Auth.Register)

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
