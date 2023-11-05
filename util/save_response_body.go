package util

import (
	"github.com/gin-gonic/gin"
)

func SaveResponseBody(c *gin.Context, errorResponse interface{}, successResponse interface{}) {
	if errorResponse != nil {
		// Handle error response
		if responseStr, ok := errorResponse.(string); ok {
			// Error response is a plain text string, store it as is
			c.Set("Response", responseStr)
		}
	} else if successResponse != nil {
		// Handle success response
		if responseStr, ok := successResponse.(string); ok {
			// Success response is a plain text string, store it as is
			c.Set("Response", responseStr)
		}
	}
}
