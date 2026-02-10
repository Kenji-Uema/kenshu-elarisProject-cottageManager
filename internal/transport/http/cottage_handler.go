package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"

	"github.com/gin-gonic/gin"
)

type CottageHandler interface {
	GetAll(c *gin.Context)
	GetByName(c *gin.Context)
}

type cottageHandler struct {
	service app.CottageService
}

func NewCottageHandler(service app.CottageService) CottageHandler {
	return &cottageHandler{service: service}
}

func (h *cottageHandler) GetAll(c *gin.Context) {
	cottages, err := h.service.GetAll(c.Request.Context())

	if err != nil {
		slog.Error("failed to retrieve cottages", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error happened while retrieving cottages. Please try again later."})
		return
	}

	cottagesDto := make([]dto.Dto, len(cottages))
	for i, cottage := range cottages {
		cottagesDto[i] = dto.FromCottageDomainToDto(cottage)
	}

	slog.Info("cottages retrieved", "count", len(cottagesDto))
	c.JSON(http.StatusOK, cottagesDto)

}

func (h *cottageHandler) GetByName(c *gin.Context) {
	cottageName := c.Param("name")

	cottage, err := h.service.GetByName(c.Request.Context(), cottageName)

	if err != nil {
		var cottageNotFound *appErrors.CottageNotFound

		if errors.As(err, &cottageNotFound) {
			slog.Warn("cottage not found", "error", err, "cottage", cottageName)
			c.JSON(404, gin.H{"error": "Cottage not found"})
			return
		}

		slog.Error("failed to retrieve cottage", "error", err, "cottage", cottageName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected error happened while retrieving cottages. Please try again later."})
		return
	}

	slog.Info("cottage retrieved", "cottage", cottageName)
	c.JSON(http.StatusOK, dto.FromCottageDomainToDto(cottage))
}
