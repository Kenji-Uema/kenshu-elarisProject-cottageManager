package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/app/fakes"
	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gin-gonic/gin"
)

func TestHandler_AddBooking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := dto.BookingRequestDto{
		GuestId:        bson.NewObjectID().Hex(),
		NumberOfGuests: 2,
		CheckInDate:    time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		CheckOutDate:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
	}

	t.Run("success returns 200", func(t *testing.T) {
		bookingSvc := fakes.NewFakeBookingService()
		availabilitySvc := fakes.NewFakeAvailabilityService()
		h := NewBookingHandler(bookingSvc, availabilitySvc)
		r := gin.New()
		r.POST("/cottage/:name/booking", h.AddBooking)

		availabilitySvc.IsCottageAvailableFunc = func(_ context.Context, _ string, _ domain.Period) (bool, error) {
			return true, nil
		}

		bookingSvc.AddBookingFunc = func(_ context.Context, _ domain.Booking) (bson.ObjectID, error) {
			return bson.NewObjectID(), nil
		}

		body, _ := json.Marshal(validBody)
		req := httptest.NewRequest(http.MethodPost, "/cottage/A1/booking", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		bookingSvc := fakes.NewFakeBookingService()
		availabilitySvc := fakes.NewFakeAvailabilityService()
		h := NewBookingHandler(bookingSvc, availabilitySvc)
		r := gin.New()
		r.POST("/cottage/:name/booking", h.AddBooking)

		req := httptest.NewRequest(http.MethodPost, "/cottage/A1/booking", bytes.NewBufferString("{invalid}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("ToDomain error returns 400 (invalid hex id)", func(t *testing.T) {
		bookingSvc := fakes.NewFakeBookingService()
		availabilitySvc := fakes.NewFakeAvailabilityService()
		h := NewBookingHandler(bookingSvc, availabilitySvc)
		r := gin.New()
		r.POST("/cottage/:name/booking", h.AddBooking)

		bad := validBody
		bad.GuestId = "not-a-hex"
		body, _ := json.Marshal(bad)
		req := httptest.NewRequest(http.MethodPost, "/cottage/A1/booking", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("bookingService error returns 500", func(t *testing.T) {
		bookingSvc := fakes.NewFakeBookingService()
		availabilitySvc := fakes.NewFakeAvailabilityService()
		h := NewBookingHandler(bookingSvc, availabilitySvc)
		r := gin.New()
		r.POST("/cottage/:name/booking", h.AddBooking)

		availabilitySvc.IsCottageAvailableFunc = func(_ context.Context, _ string, _ domain.Period) (bool, error) {
			return true, nil
		}

		bookingSvc.AddBookingFunc = func(_ context.Context, _ domain.Booking) (bson.ObjectID, error) {
			return bson.NilObjectID, errors.New("any error")
		}

		body, _ := json.Marshal(validBody)
		req := httptest.NewRequest(http.MethodPost, "/cottage/A1/booking", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestHandler_RemoveBooking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid bookingId returns 400", func(t *testing.T) {
		bookingSvc := fakes.NewFakeBookingService()
		availabilitySvc := fakes.NewFakeAvailabilityService()
		h := NewBookingHandler(bookingSvc, availabilitySvc)
		r := gin.New()
		r.DELETE("/cottage/:name/booking/:bookingId", h.RemoveBooking)

		req := httptest.NewRequest(http.MethodDelete, "/cottage/A1/booking/nothex", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("bookingService error returns 500", func(t *testing.T) {
		bookingSvc := fakes.NewFakeBookingService()
		availabilitySvc := fakes.NewFakeAvailabilityService()
		h := NewBookingHandler(bookingSvc, availabilitySvc)
		r := gin.New()
		r.DELETE("/cottage/:name/booking/:bookingId", h.RemoveBooking)

		id := bson.NewObjectID().Hex()
		bookingSvc.RemoveBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return errors.New("any error")
		}

		req := httptest.NewRequest(http.MethodDelete, "/cottage/A1/booking/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success returns 200", func(t *testing.T) {
		bookingSvc := fakes.NewFakeBookingService()
		availabilitySvc := fakes.NewFakeAvailabilityService()
		h := NewBookingHandler(bookingSvc, availabilitySvc)
		r := gin.New()
		r.DELETE("/cottage/:name/booking/:bookingId", h.RemoveBooking)

		id := bson.NewObjectID().Hex()
		bookingSvc.RemoveBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodDelete, "/cottage/A1/booking/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}
