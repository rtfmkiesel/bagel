package router

import (
	"bagel/internal/logger"
	"bytes"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// from https://github.com/gin-gonic/gin/issues/1363#issuecomment-577722498
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// from https://github.com/gin-gonic/gin/issues/1363#issuecomment-577722498
func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Logger is a simple logger middleware to route Gin logs to the custom logger
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// To capture the response body
		// from https://github.com/gin-gonic/gin/issues/1363#issuecomment-577722498
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		// Start timer to measure response time
		start := time.Now()

		// Process the request
		c.Next()

		// Calculate the latency
		latency := time.Since(start)

		// Get the path and query string
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery
		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		// Assemble the log message
		statusCode := c.Writer.Status()
		msg := fmt.Sprintf(
			"%s %s %d %s %s %s",
			c.Request.Method,
			path,
			statusCode,
			latency,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		// Log the request based on the status code
		if statusCode >= 500 {
			logger.ErrorF("%s: %s", msg, w.body.String()) // include the response body
		} else if statusCode >= 400 {
			logger.Warning("%s: %s", msg, w.body.String()) // include the response body
		} else {
			logger.Info(msg)
		}
	}
}

// ErrorHandler is a handlerFunc to route Gin errors to the custom logger
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error(e.Err)
			}
		}
	}
}
