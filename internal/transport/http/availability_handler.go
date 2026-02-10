package http

import (
	"log/slog"
	"net/http"

	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"github.com/Kenji-Uema/cottageManager/internal/transport/http/bindings"
	"github.com/gin-gonic/gin"
)

type AvailabilityHandler interface {
	GetAvailablePeriods(c *gin.Context)
	GetAvailablePeriodsByCottageType(c *gin.Context)
}

type availabilityHandler struct {
	service app.AvailabilityService
}

func NewAvailabilityHandler(service app.AvailabilityService) AvailabilityHandler {
	return &availabilityHandler{service: service}
}

func (h *availabilityHandler) GetAvailablePeriods(c *gin.Context) {
	var cottageName bindings.CottageNameURI
	var periodQuery bindings.DateRangeQuery

	if err := bindings.BindQueryAndUri(c, &cottageName, &periodQuery); err != nil {
		slog.Warn("invalid request for cottage availability", "error", err, "cottageHandler", "GetAvailablePeriods", "cottage", cottageName.Name)
		return
	}

	availablePeriods, err := h.service.GetAvailablePeriods(c.Request.Context(), cottageName.Name, bindings.ToPeriod(periodQuery.From, periodQuery.To))

	if err != nil {
		slog.Error("failed to get available periods", "error", err, "cottage", cottageName.Name)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.AvailablePeriodDTO, len(availablePeriods))
	for i, period := range availablePeriods {
		response[i] = dto.FromDomain(period)
	}

	slog.Debug("calculated availability", "cottage", cottageName.Name, "periods", len(availablePeriods))
	c.JSON(http.StatusOK, response)
}

func (h *availabilityHandler) GetAvailablePeriodsByCottageType(c *gin.Context) {
	var cottageType bindings.CottageTypeURI
	var periodQuery bindings.DateRangeQuery

	if err := bindings.BindQueryAndUri(c, &cottageType, &periodQuery); err != nil {
		slog.Warn("invalid request for cottage type availability", "error", err, "cottageHandler", "GetAvailablePeriodsByCottageType", "cottage_type", cottageType.Type)
		return
	}

	availablePeriods, err := h.service.GetAvailablePeriodsByCottageType(c.Request.Context(), cottageType.Type, bindings.ToPeriod(periodQuery.From, periodQuery.To))

	if err != nil {
		slog.Error("failed to get cottage type availability", "error", err, "cottage_type", cottageType.Type)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.Debug("calculated availability for cottage type", "cottage_type", cottageType.Type, "count", len(availablePeriods))
	c.JSON(http.StatusOK, availablePeriods)
}
