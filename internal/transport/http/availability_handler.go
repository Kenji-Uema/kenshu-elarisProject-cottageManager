package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/validationErrors"
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

// GetAvailablePeriods godoc
// @Summary Get availability by cottage
// @Description Return available periods for a cottage within a date range
// @Tags availability
// @Produce JSON
// @Param name path string true "Cottage name"
// @Param from query string true "Start date (YYYY-MM-DD)"
// @Param to query string true "End date (YYYY-MM-DD)"
// @Success 200 {array} dto.AvailablePeriodDTO
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cottage/{name}/available-dates [get]
func (h *availabilityHandler) GetAvailablePeriods(c *gin.Context) {
	var cottageName bindings.CottageNameURI
	var periodQuery bindings.DateRangeQuery

	if bindErr := bindings.BindQueryAndUri(c, &cottageName, &periodQuery); bindErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, bindErr)
		return
	}

	availablePeriods, err := h.service.GetAvailablePeriods(c.Request.Context(), cottageName.Name, domain.Period{CheckIn: periodQuery.From, CheckOut: periodQuery.To})

	if err != nil {
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			slog.WarnContext(c.Request.Context(), "invalid availability request", "error", err, "cottage", cottageName.Name)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var cottageNotFoundErr *appErrors.CottageNotFound
		if errors.As(err, &cottageNotFoundErr) {
			slog.WarnContext(c.Request.Context(), "cottage not found while getting availability", "error", err, "cottage", cottageName.Name)
			c.JSON(http.StatusNotFound, gin.H{"error": "Cottage not found"})
			return
		}

		slog.ErrorContext(c.Request.Context(), "failed to get available periods", "error", err, "cottage", cottageName.Name)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error happened while getting availability. Please try again later."})
		return
	}

	response := availablePeriods.ToDto()

	slog.DebugContext(c.Request.Context(), "calculated availability", "availablePeriods", availablePeriods)
	c.JSON(http.StatusOK, response)
}

// GetAvailablePeriodsByCottageType godoc
// @Summary Get availability by cottage type
// @Description Return available periods for cottages of a type within a date range
// @Tags availability
// @Produce JSON
// @Param cottageType path string true "Cottage type"
// @Param from query string true "Start date (YYYY-MM-DD)"
// @Param to query string true "End date (YYYY-MM-DD)"
// @Success 200 {array} domain.CottageAvailablePeriod
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cottage/type/{cottageType}/available-dates [get]
func (h *availabilityHandler) GetAvailablePeriodsByCottageType(c *gin.Context) {
	var cottageType bindings.CottageTypeURI
	var periodQuery bindings.DateRangeQuery

	if bindErr := bindings.BindQueryAndUri(c, &cottageType, &periodQuery); bindErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, bindErr)
		return
	}

	availablePeriods, err := h.service.GetAvailablePeriodsByCottageType(c.Request.Context(), cottageType.CottageType, domain.Period{CheckIn: periodQuery.From, CheckOut: periodQuery.To})

	if err != nil {
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			slog.WarnContext(c.Request.Context(), "invalid cottage type availability request", "error", err, "cottage_type", cottageType.CottageType)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		slog.ErrorContext(c.Request.Context(), "failed to get cottage type availability", "error", err, "cottage_type", cottageType.CottageType)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error happened while getting availability. Please try again later."})
		return
	}

	slog.DebugContext(c.Request.Context(), "calculated availability for cottage type", "cottage_type", cottageType.CottageType, "count", len(availablePeriods))
	c.JSON(http.StatusOK, availablePeriods)
}
