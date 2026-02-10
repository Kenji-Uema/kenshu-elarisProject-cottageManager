package booking

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handler interface {
	AddBooking(c *gin.Context)
	RemoveBooking(c *gin.Context)
}

type handler struct {
	service app.BookingService
}

func NewHandler(service app.BookingService) Handler {
	return &handler{service: service}
}

func (h *handler) AddBooking(c *gin.Context) {
	cottageNameURI := c.Param("name")
	var bookingRequest RequestDto

	if err := c.ShouldBindJSON(&bookingRequest); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Warn("invalid booking payload", "error", err, "cottage", cottageNameURI)
		return
	}

	booking, err := bookingRequest.ToDomain(cottageNameURI)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		slog.Warn("failed to convert booking payload", "error", err, "cottage", cottageNameURI)
		return
	}

	bookingId, err := h.service.AddBooking(c.Request.Context(), booking)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("failed to add booking", "error", err, "cottage", cottageNameURI)
		return
	}

	slog.Info("booking added", "cottage", cottageNameURI, "booking_id", bookingId, "period_start", booking.StayPeriod.Start, "period_end", booking.StayPeriod.End)
	c.JSON(http.StatusOK, ConfirmationDto{BookingId: bookingId})
}

func (h *handler) RemoveBooking(c *gin.Context) {
	cottageName := c.Param("name")
	bookingIdHex := c.Param("bookingId")

	bookingId, err := bson.ObjectIDFromHex(bookingIdHex)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid hex: %v", err)})
		slog.Warn("invalid booking id provided", "error", err, "cottage", cottageName, "booking_id", bookingIdHex)
		return
	}

	err = h.service.RemoveBooking(c.Request.Context(), cottageName, bookingId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		slog.Error("failed to remove booking", "error", err, "cottage", cottageName, "booking_id", bookingIdHex)
		return
	}

	slog.Info("booking removed", "cottage", cottageName, "booking_id", bookingIdHex)
	c.JSON(http.StatusOK, gin.H{"message": "Booking removed successfully"})
}
