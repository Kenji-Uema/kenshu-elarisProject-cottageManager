package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// RequestLogger adds structured request logging to every HTTP call handled by gin.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		spanCtx := trace.SpanFromContext(c.Request.Context()).SpanContext()
		fields := []any{"method", c.Request.Method, "path", path, "client_ip", c.ClientIP()}
		if spanCtx.IsValid() {
			fields = append(fields, "trace_id", spanCtx.TraceID().String())
		}

		slog.Debug("request started", fields...)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				errorFields := append([]any{}, fields...)
				errorFields = append(errorFields, "status", status, "duration", latency, "error", err.Error())
				slog.Error("request failed", errorFields...)
			}
			return
		}

		completeFields := append([]any{}, fields...)
		completeFields = append(completeFields, "status", status, "duration", latency)
		slog.Info("request completed", completeFields...)
	}
}
