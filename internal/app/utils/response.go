package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RespondJSON is a generic utility to send JSON responses
func RespondJSON(c *gin.Context, statusCode int, payload interface{}) {
	logrus.WithFields(logrus.Fields{
		"status": statusCode,
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	}).Info("Responding with JSON")
	c.JSON(statusCode, payload)
	c.Done()
}

// ErrorResponse sends error messages
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	requestId, exists := c.Get("RequestID")
	if exists {
		logrus.WithFields(logrus.Fields{
			"status":    statusCode,
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"RequestId": requestId,
			"message":   message,
		}).Error("Error response")
	} else {
		logrus.WithFields(logrus.Fields{
			"status":  statusCode,
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"message": message,
		}).Error("Error response")
	}
	c.AbortWithStatusJSON(statusCode, gin.H{"error": message})
}

type ErrorMessage struct {
	StatusCode int
	Message    string `json:"error"`
}

func (e *ErrorMessage) Error() string {
	return e.Message
}

func NewValidationErrorMessage(message string) *ErrorMessage {
	return &ErrorMessage{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

func NewErrorMessage(message string) *ErrorMessage {
	return &ErrorMessage{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}
