package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/validationErrors"

	"github.com/gin-gonic/gin"
)

type CottageHandler interface {
	GetAll(c *gin.Context)
	GetByName(c *gin.Context)
	GetByView(c *gin.Context)
}

type cottageHandler struct {
	service app.CottageService
}

func NewCottageHandler(service app.CottageService) CottageHandler {
	return &cottageHandler{service: service}
}

// GetAll godoc
// @Summary List cottages
// @Description Return all cottages with details
// @Tags cottages
// @Produce JSON
// @Success 200 {array} dto.Cottage
// @Failure 500 {object} map[string]string
// @Router /cottages [get]
func (h *cottageHandler) GetAll(c *gin.Context) {
	cottages, err := h.service.GetAll(c.Request.Context())

	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to retrieve cottages", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error happened while retrieving cottages. Please try again later."})
		return
	}

	cottagesDto := make([]dto.Cottage, len(cottages))
	for i, cottage := range cottages {
		cottagesDto[i] = cottage.ToDto()
	}

	c.JSON(http.StatusOK, cottagesDto)
}

// GetByName godoc
// @Summary Get cottage by name
// @Description Return a single cottage by its name
// @Tags cottages
// @Produce JSON
// @Param name path string true "Cottage name"
// @Success 200 {object} dto.Cottage
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cottage/{name} [get]
func (h *cottageHandler) GetByName(c *gin.Context) {
	cottageName := c.Param("name")

	cottage, err := h.service.GetByName(c.Request.Context(), cottageName)

	if err != nil {
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			slog.WarnContext(c.Request.Context(), "invalid cottage name", "error", err, "cottage", cottageName)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var cottageNotFound *appErrors.CottageNotFound
		if errors.As(err, &cottageNotFound) {
			slog.WarnContext(c.Request.Context(), "cottage not found", "error", err, "cottage", cottageName)
			c.JSON(404, gin.H{"error": "Cottage not found"})
			return
		}

		slog.ErrorContext(c.Request.Context(), "failed to retrieve cottage", "error", err, "cottage", cottageName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error happened while retrieving cottages. Please try again later."})
		return
	}

	c.JSON(http.StatusOK, cottage.ToDto())
}

func (h *cottageHandler) GetByView(c *gin.Context) {
	view := c.Param("view")

	cottages, err := h.service.GetByView(c.Request.Context(), view)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "failed to retrieve cottages by view", "error", err, "view", view)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error happened while retrieving cottages. Please try again later."})
		return
	}

	cottagesDto := make([]dto.Cottage, len(cottages))
	for i, cottage := range cottages {
		cottagesDto[i] = cottage.ToDto()
	}

	c.JSON(http.StatusOK, cottagesDto)
}
