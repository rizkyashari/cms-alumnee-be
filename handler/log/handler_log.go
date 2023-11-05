package handler_log

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/service"
	"github.com/fadhln/lms-be/util"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LogHandler interface {
	LogActivity(c *gin.Context)
	GetActivityLogs(c *gin.Context)
}

type impHandler struct {
	s            service.Service
	activityLogs []string
}

func Init(s service.Service) LogHandler {
	return &impHandler{
		s:            s,
		activityLogs: []string{},
	}
}

func (h *impHandler) LogActivity(c *gin.Context) {

	gotRequestBody := c.MustGet("RequestBody")
	fmt.Println("gotRequestBody: ", gotRequestBody)
	requestBody, ok := gotRequestBody.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert requestBody to string"})
		return
	}

	gotResponse := c.MustGet("Response")
	fmt.Println(gotResponse)
	var responseData string

	if responseStr, ok := gotResponse.(string); ok {
		responseData = responseStr
	}

	gotAccount, _ := h.s.Auth().GetOwnAccountDetail(c)
	requestMethod := c.Request.Method
	requestURL := c.Request.URL.Path
	userRole := gotAccount.AccountType
	userEmail := gotAccount.Email
	userName := gotAccount.Name
	timestamp := time.Now()

	responseWriter := newResponseBodyCapture(c.Writer)

	c.Writer = responseWriter

	c.Next()

	responseStatus := c.Writer.Status()
	responseStatusDescription := http.StatusText(responseStatus)

	logRequest := model.LogData{
		ID:                        uuid.New(),
		CreatedAt:                 timestamp,
		UserRole:                  userRole,
		UserName:                  userName,
		UserEmail:                 userEmail,
		RequestMethod:             requestMethod,
		RequestURL:                requestURL,
		RequestBody:               requestBody,
		ResponseData:              responseData,
		ResponseStatus:            responseStatus,
		ResponseStatusDescription: responseStatusDescription,
	}

	err := h.s.Log().LogActivity(c, logRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log activity"})
		return
	}

	logRequest.ResponseData = responseData

	err = h.s.Log().UpdateLogActivity(c, logRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update log activity"})
		return
	}
}

func (h *impHandler) GetActivityLogs(c *gin.Context) {
	limit, page, sortBy, sortOrder, err := util.ParseQuery(c)
	if err != nil {
		rs.ErrorResponse(c, errmsg.ErrRequestParamsInvalid)
		return
	}

	searchUserRole := c.DefaultQuery("user_role", "")
	userRoleNum := 0

	if searchUserRole != "" {
		num, err := strconv.Atoi(searchUserRole)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			userRoleNum = num
		}
	}

	filters := model.LogData{
		UserRole: userRoleNum,
	}

	params := rq.PaginationParams[model.LogData]{
		Limit:     limit,
		Page:      page,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Data:      filters,
	}

	// Retrieve activity logs from the database using the service
	logs, err := h.s.Log().GetActivityLogs(c, &params)
	if err != nil {
		// Handle the error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	rs.SuccessResponse(c, logs, http.StatusOK)
}

type responseBodyCapture struct {
	gin.ResponseWriter
	bodyBuffer *bytes.Buffer
}

func newResponseBodyCapture(w gin.ResponseWriter) responseBodyCapture {
	return responseBodyCapture{
		ResponseWriter: w,
		bodyBuffer:     bytes.NewBuffer(nil),
	}
}
