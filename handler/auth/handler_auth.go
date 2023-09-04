package handler_auth

import (
	"net/http"
	"strings"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	auth_util "github.com/fadhln/lms-be/util/auth"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	CheckAuth() gin.HandlerFunc
	CheckAdmin() gin.HandlerFunc
	CheckStudent() gin.HandlerFunc
	CheckTeacher() gin.HandlerFunc

	GetOwnAccountDetail(c *gin.Context)

	Login(c *gin.Context)
	Register(c *gin.Context)
	CheckEmailExist(c *gin.Context)
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
	return func(c *gin.Context) {
		authCheck := c.Request.Header["Authorization"]
		if len(authCheck) < 1 {
			rs.ErrorResponse(c, errmsg.ErrRequestHeaderInvalid)
			c.Abort()
			return
		}

		authString := authCheck[0]
		tokenCheck := strings.Split(authString, " ")
		if tokenCheck[0] != "Bearer" || len(tokenCheck) < 1 {
			rs.ErrorResponse(c, errmsg.ErrRequestHeaderInvalid)
			c.Abort()
			return
		}

		if len(tokenCheck) != 2 {
			rs.ErrorResponse(c, errmsg.ErrRequestHeaderInvalid)
			c.Abort()
			return
		}
		token := tokenCheck[1]

		email, accountType, err := auth_util.CheckToken(token)
		if err != nil {
			rs.ErrorResponse(c, err)
			c.Abort()
			return
		}

		res, err := h.s.Account().GetDetailWithAccType(c,
			&rq.EmailAndAccTypeRequest{
				Email:       email,
				AccountType: accountType},
		)

		if err != nil {
			rs.ErrorResponse(c, err)
			c.Abort()
			return
		}

		c.Set("account", res)
		c.Next()
	}
}

func (h *impHandler) CheckAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		gotAccount, err := util.GetAccountContext(c)
		if err != nil {
			rs.ErrorResponse(c, err)
			c.Abort()
			return
		}

		if gotAccount.AccountType != constants.ACCOUNT_ADMIN {
			rs.ErrorResponse(c, &errmsg.ErrUserIsNot{FieldName: "Admin"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *impHandler) CheckStudent() gin.HandlerFunc {
	return func(c *gin.Context) {
		gotAccount, err := util.GetAccountContext(c)
		if err != nil {
			rs.ErrorResponse(c, err)
			c.Abort()
			return
		}

		if gotAccount.AccountType != constants.ACCOUNT_STUDENT {
			rs.ErrorResponse(c, &errmsg.ErrUserIsNot{FieldName: "Student"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *impHandler) CheckTeacher() gin.HandlerFunc {
	return func(c *gin.Context) {
		gotAccount, err := util.GetAccountContext(c)
		if err != nil {
			rs.ErrorResponse(c, err)
			c.Abort()
			return
		}

		if gotAccount.AccountType != constants.ACCOUNT_TEACHER {
			rs.ErrorResponse(c, &errmsg.ErrUserIsNot{FieldName: "Teacher"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *impHandler) Login(c *gin.Context) {
	var req rq.LoginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	res, err := h.s.Auth().Login(c, &req)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, *res, http.StatusOK)
}

func (h *impHandler) Register(c *gin.Context) {
	var req rq.RegisterRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	res, err := h.s.Auth().Register(c, &req)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, *res, http.StatusCreated)

}

// TODO: write function
func (h *impHandler) CheckEmailExist(c *gin.Context) {}

func (h *impHandler) GetOwnAccountDetail(c *gin.Context) {
	res, err := h.s.Auth().GetOwnAccountDetail(c)
	if err != nil {
		rs.ErrorResponse(c, err)
		return
	}

	rs.SuccessResponse(c, *res, http.StatusOK)
}
