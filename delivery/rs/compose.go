package rs

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JSONResponse struct {
	Message string      `json:"message,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

func returnJSONResponse(c *gin.Context, message string, data interface{}, statusCode int) {
	w := c.Writer
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(JSONResponse{
		Message: message,
		Content: data,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func SuccessResponse(c *gin.Context, data interface{}, statusCode ...int) {
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	returnJSONResponse(c, "success", data, code)
}

func ErrorResponse(c *gin.Context, err error) {
	code := http.StatusBadRequest
	msg := err.Error()
	if errors.Is(err, &errmsg.ErrInternal{}) {
		code = http.StatusInternalServerError
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		code = http.StatusNoContent
		msg = "Not found"
	} else if err.Error() == "EOF" {
		code = http.StatusBadRequest
		msg = "Invalid body format"
	} else if strings.Contains(err.Error(), "deadline exceeded") {
		code = http.StatusRequestTimeout
		msg = "Request timeout"
	}

	returnJSONResponse(c, msg, nil, code)
}

func CustomResponse(c *gin.Context, data any, message string, statusCode ...int) {
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	returnJSONResponse(c, message, data, code)
}
