package http

import (
	"net/http"

	"github.com/Kenji-Uema/cottageManager/internal/infra/db"
	"github.com/gin-gonic/gin"
)

type ProbeHandler interface {
	Heath(c *gin.Context)
	Ready(c *gin.Context)
}

type probeHandler struct {
	mongoClient *db.Db
}

func NewProbeHandler(mongoClient *db.Db) ProbeHandler {
	return &probeHandler{mongoClient: mongoClient}
}

// Heath godoc
// @Summary Liveness probe
// @Description Returns 200 when bookingService is running
// @Tags health
// @Success 200
// @Router /healthz [get]
func (p probeHandler) Heath(c *gin.Context) {
	c.Status(http.StatusOK)
}

// Ready godoc
// @Summary Readiness probe
// @Description Returns 200 when dependencies are available
// @Tags health
// @Success 200
// @Failure 503
// @Router /readyz [get]
func (p probeHandler) Ready(c *gin.Context) {
	if err := p.mongoClient.Ping(); err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	c.Status(http.StatusOK)
}
