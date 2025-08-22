package server

import (
	"github.com/changhyeonkim/pray-together/go-api-server/internal/config"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/handler/middleware"
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

// LoggerMiddleware returns a gin middleware for structured logging with slog
func LoggerMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		status := c.Writer.Status()

		// Build log fields
		fields := []any{
			"status", status,
			"method", c.Request.Method,
			"path", path,
			"ip", c.ClientIP(),
			"latency", latency.String(),
			"user_agent", c.Request.UserAgent(),
		}

		// Add request ID if exists
		if requestID, exists := c.Get(middleware.RequestIDKey); exists {
			fields = append(fields, middleware.RequestIDKey, requestID)
		}

		if raw != "" {
			fields = append(fields, "query", raw)
		}

		// Add error if exists
		if len(c.Errors) > 0 {
			fields = append(fields, "error", c.Errors.String())
		}

		// Log based on status code
		msg := "Request processed"

		switch {
		case status >= 500:
			slog.Error(msg, fields...)
		case status >= 400:
			slog.Warn(msg, fields...)
		case status >= 300:
			slog.Info(msg, fields...)
		default:
			slog.Info(msg, fields...)
		}
	}
}
