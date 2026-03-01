package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/validationErrors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BookingHandler interface {
	AddBooking(c *gin.Context)
	RemoveBooking(c *gin.Context)
}

type bookingHandler struct {
	bookingService      app.BookingService
	availabilityService app.AvailabilityService
}

func NewBookingHandler(service app.BookingService, availabilityService app.AvailabilityService) BookingHandler {
	return &bookingHandler{bookingService: service, availabilityService: availabilityService}
}

// AddBooking godoc
// @Summary Create booking
// @Description Register a booking for a cottage
// @Tags bookings
// @Accept json
// @Produce json
// @Param name path string true "Cottage name"
// @Param request body dto.BookingRequestDto true "IsValidBooking request"
// @Success 200 {object} dto.ConfirmationDto
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cottage/{name}/booking [post]
func (h *bookingHandler) AddBooking(c *gin.Context) {
	cottageNameURI := c.Param("name")
	var bookingRequest dto.BookingRequestDto

	if err := c.ShouldBindJSON(&bookingRequest); err != nil {
		slog.WarnContext(c.Request.Context(), "invalid booking payload", "error", err, "cottage", cottageNameURI)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	booking, err := domain.NewBookingFromDto(bookingRequest, cottageNameURI)

	if err != nil {
		slog.WarnContext(c.Request.Context(), "failed to convert booking payload", "error", err, "cottage", cottageNameURI)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isCottageFree, err := h.availabilityService.IsCottageAvailable(c.Request.Context(), booking.CottageName, booking.StayPeriod)

	if err != nil {
		slog.Error("failed to validate cottage availability", "error", err, "cottage", booking.CottageName)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !isCottageFree {
		slog.Warn("cottage not available for requested period", "cottage", booking.CottageName, "check_in", booking.StayPeriod.CheckIn, "check_out", booking.StayPeriod.CheckOut)
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Cottage not available for requested period"})
		return
	}

	bookingId, err := h.bookingService.AddBooking(c.Request.Context(), booking)

	if err != nil {
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			slog.WarnContext(c.Request.Context(), "invalid booking payload", "error", err, "cottage", cottageNameURI)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var cottageNotFoundErr *appErrors.CottageNotFound
		if errors.As(err, &cottageNotFoundErr) {
			slog.WarnContext(c.Request.Context(), "cottage not found while booking", "error", err, "cottage", cottageNameURI)
			c.JSON(http.StatusNotFound, gin.H{"error": "Cottage not found"})
			return
		}

		var cottageNotAvailableErr *appErrors.CottageNotAvailableError
		if errors.As(err, &cottageNotAvailableErr) {
			slog.WarnContext(c.Request.Context(), "cottage not available for booking", "error", err, "cottage", cottageNameURI)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		slog.ErrorContext(c.Request.Context(), "failed to add booking", "error", err, "cottage", cottageNameURI)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.InfoContext(c.Request.Context(), "booking added", "cottage", cottageNameURI, "booking_id", bookingId, "period", booking.StayPeriod)

	c.JSON(http.StatusOK, dto.ConfirmationDto{
		Message:   "Thank you for choosing us. Your booking registered, soon you will receive your invoice",
		BookingId: bookingId.Hex(),
		Info: dto.ConfirmationInfoDto{
			CottageName:    cottageNameURI,
			NumberOfGuests: bookingRequest.NumberOfGuests,
			CheckInDate:    bookingRequest.CheckInDate,
			CheckOutDate:   bookingRequest.CheckOutDate,
		},
	})
}

// RemoveBooking godoc
// @Summary Delete booking
// @Description Cancel a booking for a cottage
// @Tags bookings
// @Produce json
// @Param name path string true "Cottage name"
// @Param bookingId path string true "IsValidBooking ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /cottage/{name}/booking/{bookingId} [delete]
func (h *bookingHandler) RemoveBooking(c *gin.Context) {
	cottageName := c.Param("name")
	bookingIdHex := c.Param("bookingId")

	bookingId, err := bson.ObjectIDFromHex(bookingIdHex)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid hex: %v", err)})
		slog.WarnContext(c.Request.Context(), "invalid booking id provided", "error", err, "cottage", cottageName, "booking_id", bookingIdHex)
		return
	}

	err = h.bookingService.RemoveBooking(c.Request.Context(), cottageName, bookingId)

	if err != nil {
		var validationErr *validationErrors.ErrValidationConstrain
		if errors.As(err, &validationErr) {
			slog.WarnContext(c.Request.Context(), "invalid remove booking request", "error", err, "cottage", cottageName, "booking_id", bookingIdHex)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var bookingNotFoundErr *appErrors.BookingNotFound
		if errors.As(err, &bookingNotFoundErr) {
			slog.WarnContext(c.Request.Context(), "booking not found", "error", err, "cottage", cottageName, "booking_id", bookingIdHex)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		var cottageNotFoundErr *appErrors.CottageNotFound
		if errors.As(err, &cottageNotFoundErr) {
			slog.WarnContext(c.Request.Context(), "cottage not found while removing booking", "error", err, "cottage", cottageName, "booking_id", bookingIdHex)
			c.JSON(http.StatusNotFound, gin.H{"error": "Cottage not found"})
			return
		}

		slog.ErrorContext(c.Request.Context(), "failed to remove booking", "error", err, "cottage", cottageName, "booking_id", bookingIdHex)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.InfoContext(c.Request.Context(), "booking removed", "cottage", cottageName, "booking_id", bookingIdHex)
	c.JSON(http.StatusOK, gin.H{"message": "IsValidBooking removed successfully"})
}
