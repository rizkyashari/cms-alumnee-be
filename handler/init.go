package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// repository := repo.SetupRepo(db)
	// service := service.SetupService(repository)
	// handler := SetupHandler(service)

	// h := handler

	return r

}
