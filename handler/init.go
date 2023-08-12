package handler

import (
	"github.com/fadhln/lms-be/repo"
	"github.com/fadhln/lms-be/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	repository := repo.SetupRepo(db)
	service := service.SetupService(repository)
	handler := SetupHandler(service)

	h := handler

	r.POST("/login", h.Auth.Login)
	r.POST("/register", h.Auth.Register)

	return r
}
