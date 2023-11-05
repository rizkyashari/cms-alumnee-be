package util

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SaveRequestBody(c *gin.Context, request io.Reader) error {
	requestData, err := io.ReadAll(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return err
	}

	// Convert the request body to a string
	requestBody := string(requestData)

	// Set the request body in the context
	c.Set("RequestBody", requestBody)

	// Restore the request body so it can be read again by other parts of the application
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestData))

	return nil
}
