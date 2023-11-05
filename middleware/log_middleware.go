package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture relevant information for logging
		requestMethod := c.Request.Method
		requestURL := c.Request.URL.Path
		userRole := "admin" // Set the role based on the user's authentication
		timestamp := time.Now().Format("2006-01-02 15:04:05")

		// Log the incoming request
		logMessage := fmt.Sprintf("%s | %s | %s | %s", timestamp, userRole, requestMethod, requestURL)
		fmt.Println(logMessage) // You can replace this with your preferred logging mechanism

		// Continue with the request
		c.Next()
	}
}
