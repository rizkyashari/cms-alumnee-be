package handler_auth

import (
	"github.com/fadhln/lms-be/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	CheckAuth() gin.HandlerFunc
	CheckAdmin() gin.HandlerFunc

	Login(c *gin.Context)
	Register(c *gin.Context)
	CheckEmailExist(c *gin.Context)
	PasswordRecovery(c *gin.Context)
}

type impHandler struct {
	s service.Service
}

func Init(s service.Service) AuthHandler {
	return &impHandler{
		s: s,
	}
}

func (h *impHandler) CheckAuth() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func (h *impHandler) CheckAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func (h *impHandler) Login(c *gin.Context)            {}
func (h *impHandler) Register(c *gin.Context)         {}
func (h *impHandler) CheckEmailExist(c *gin.Context)  {}
func (h *impHandler) PasswordRecovery(c *gin.Context) {}
