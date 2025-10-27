package availability

import (
	"cottageManager/internal/app"
	"cottageManager/internal/transport/http/common"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	GetAvailablePeriods(c *gin.Context)
	GetAvailablePeriodsByCottageType(c *gin.Context)
}

type availabilityHandler struct {
	service app.AvailabilityService
}

func NewHandler(service app.AvailabilityService) Handler {
	return &availabilityHandler{service: service}
}

func (h *availabilityHandler) GetAvailablePeriods(c *gin.Context) {
	var cottageName common.CottageNameURI
	var periodQuery common.DateRangeQuery

	if err := common.BindQueryAndUri(c, &cottageName, &periodQuery); err != nil {
		slog.Warn("invalid request for cottage availability", "error", err, "handler", "GetAvailablePeriods", "cottage", cottageName.Name)
		return
	}

	availablePeriods, err := h.service.GetAvailablePeriods(c.Request.Context(), cottageName.Name, common.ToPeriod(periodQuery.From, periodQuery.To))

	if err != nil {
		slog.Error("failed to get available periods", "error", err, "cottage", cottageName.Name)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]AvailablePeriodDTO, len(availablePeriods))
	for i, period := range availablePeriods {
		response[i] = fromDomain(period)
	}

	slog.Debug("calculated availability", "cottage", cottageName.Name, "periods", len(availablePeriods))
	c.JSON(http.StatusOK, response)
}

func (h *availabilityHandler) GetAvailablePeriodsByCottageType(c *gin.Context) {
	var cottageType common.CottageTypeURI
	var periodQuery common.DateRangeQuery

	if err := common.BindQueryAndUri(c, &cottageType, &periodQuery); err != nil {
		slog.Warn("invalid request for cottage type availability", "error", err, "handler", "GetAvailablePeriodsByCottageType", "cottage_type", cottageType.Type)
		return
	}

	availablePeriods, err := h.service.GetAvailablePeriodsByCottageType(c.Request.Context(), cottageType.Type, common.ToPeriod(periodQuery.From, periodQuery.To))

	if err != nil {
		slog.Error("failed to get cottage type availability", "error", err, "cottage_type", cottageType.Type)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.Debug("calculated availability for cottage type", "cottage_type", cottageType.Type, "count", len(availablePeriods))
	c.JSON(http.StatusOK, availablePeriods)
}
