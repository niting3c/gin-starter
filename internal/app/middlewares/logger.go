package middlewares

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"starter/internal/app/utils"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/sirupsen/logrus"
)

const (
	REQUEST_ID     = "RequestID"
	REQUEST_HEADER = "X-Request-ID"
)

// LoggerMiddleware returns a Gin middleware that logs HTTP requests.
func LoggerMiddleware(logger logrus.FieldLogger, notLogged ...string) gin.HandlerFunc {
	// Get the hostname, or set it as "unknown" if an error occurs.
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Create a map to store paths that should not be logged.
	var skip map[string]struct{}
	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, p := range notLogged {
			skip[p] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Save the original path since it might be modified by other handlers.
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		dataLength := c.Writer.Size()
		reqId, _ := c.Get(REQUEST_ID)
		if dataLength < 0 {
			dataLength = 0
		}

		// Skip logging if the path is in the skip list.
		if _, ok := skip[path]; ok {
			return
		}

		// Prepare log entry fields.
		entryFields := logrus.Fields{
			"hostname":   hostname,
			"statusCode": statusCode,
			"latency":    latency, // Time taken to process the request.
			"clientIP":   clientIP,
			"method":     c.Request.Method,
			"path":       path,
			"referer":    referer,
			"dataLength": dataLength,
			"userAgent":  clientUserAgent,
			"requestId":  reqId,
		}

		// Log based on status code severity.
		logLevel := logrus.InfoLevel
		switch {
		case statusCode >= http.StatusInternalServerError:
			logLevel = logrus.ErrorLevel
		case statusCode >= http.StatusBadRequest:
			logLevel = logrus.WarnLevel
		}

		logMessage := fmt.Sprintf("[%s] - ReqId: %s - ClientIP: %s - Hostname: %s - Time: %s - Method: %s - Path: %s - StatusCode: %d - DataLength: %d - Referer: %s - UserAgent: %s - Latency: %dms",
			logLevel,
			reqId,
			clientIP,
			hostname,
			time.Now().Format(time.RFC3339),
			c.Request.Method,
			path,
			statusCode,
			dataLength,
			referer,
			clientUserAgent,
			latency)

		// Log the message with appropriate log level.
		switch logLevel {
		case logrus.ErrorLevel:
			logger.WithFields(entryFields).Error(logMessage)
		case logrus.WarnLevel:
			logger.WithFields(entryFields).Warn(logMessage)
		case logrus.DebugLevel:
			logger.WithFields(entryFields).Debug(logMessage)
		case logrus.TraceLevel:
			logger.WithFields(entryFields).Trace(logMessage)
		case logrus.PanicLevel:
			logger.WithFields(entryFields).Panic(logMessage)
		default:
			logger.WithFields(entryFields).Info(logMessage)
		}
	}
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid, _ := uuid.NewV7()
		requestID := uuid.String()
		c.Set(REQUEST_ID, requestID)
		c.Writer.Header().Set(REQUEST_HEADER, requestID)
		c.Next()
	}
}

func TimeoutMiddleware() gin.HandlerFunc {
	timeoutDuration, _ := time.ParseDuration(utils.GetEnvAsString("REQUEST_TIMEOUT_SEC", "120s"))
	//Setting request timeout
	return timeout.New(
		timeout.WithTimeout(timeoutDuration),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{"error": "Request Timeout"})
		}))
}
