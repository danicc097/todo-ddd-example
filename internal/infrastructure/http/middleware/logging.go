package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		attrs := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("ip", c.ClientIP()),
		}

		if len(c.Errors) > 0 {
			slog.ErrorContext(c.Request.Context(), "http_request_failed", append(attrs, slog.String("errors", c.Errors.String()))...)
		} else {
			slog.InfoContext(c.Request.Context(), "http_request_processed", attrs...)
		}
	}
}
