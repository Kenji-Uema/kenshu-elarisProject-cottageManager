package health

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type readinessChecker interface {
	Ping(ctx context.Context) error
}

type Handler interface {
	Health(c *gin.Context)
	Liveness(c *gin.Context)
	Readiness(c *gin.Context)
}

type handler struct {
	readiness readinessChecker
	timeout   time.Duration
}

func NewHandler(readiness readinessChecker) Handler {
	return &handler{
		readiness: readiness,
		timeout:   2 * time.Second,
	}
}

func (h *handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *handler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

func (h *handler) Readiness(c *gin.Context) {
	if h.readiness == nil {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
		return
	}

	ctx := c.Request.Context()
	if h.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.timeout)
		defer cancel()
	}

	if err := h.readiness.Ping(ctx); err != nil {
		slog.Error("readiness check failed", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
