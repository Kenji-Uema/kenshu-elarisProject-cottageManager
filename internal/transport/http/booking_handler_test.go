package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appmocks "github.com/Kenji-Uema/cottageManager/internal/app/mocks"
	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gin-gonic/gin"
)

func TestHandler_AddBooking(t *testing.T) {
	setupGin()

	validBody := dto.RequestDto{
		GuestId:        bson.NewObjectID().Hex(),
		NumberOfGuests: 2,
		CheckInDate:    time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		CheckOutDate:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
	}

	t.Run("success returns 200", func(t *testing.T) {
		svc := appmocks.NewMockBookingService()
		h := NewBookingHandler(svc)
		r := gin.New()
		r.POST("/booking/:name", h.AddBooking)

		svc.AddBookingFunc = func(_ context.Context, _ domain.Booking) (string, error) {
			return "id123", nil
		}

		body, _ := json.Marshal(validBody)
		req := httptest.NewRequest(http.MethodPost, "/booking/A1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		svc := appmocks.NewMockBookingService()
		h := NewBookingHandler(svc)
		r := gin.New()
		r.POST("/booking/:name", h.AddBooking)

		req := httptest.NewRequest(http.MethodPost, "/booking/A1", bytes.NewBufferString("{invalid}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("ToDomain error returns 400 (invalid hex id)", func(t *testing.T) {
		svc := appmocks.NewMockBookingService()
		h := NewBookingHandler(svc)
		r := gin.New()
		r.POST("/booking/:name", h.AddBooking)

		bad := validBody
		bad.GuestId = "not-a-hex"
		body, _ := json.Marshal(bad)
		req := httptest.NewRequest(http.MethodPost, "/booking/A1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := appmocks.NewMockBookingService()
		h := NewBookingHandler(svc)
		r := gin.New()
		r.POST("/booking/:name", h.AddBooking)

		svc.AddBookingFunc = func(_ context.Context, _ domain.Booking) (string, error) {
			return "", assertAnyError()
		}

		body, _ := json.Marshal(validBody)
		req := httptest.NewRequest(http.MethodPost, "/booking/A1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestHandler_RemoveBooking(t *testing.T) {
	setupGin()

	t.Run("invalid bookingId returns 400", func(t *testing.T) {
		svc := appmocks.NewMockBookingService()
		h := NewBookingHandler(svc)
		r := gin.New()
		r.DELETE("/booking/:name/:bookingId", h.RemoveBooking)

		req := httptest.NewRequest(http.MethodDelete, "/booking/A1/nothex", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := appmocks.NewMockBookingService()
		h := NewBookingHandler(svc)
		r := gin.New()
		r.DELETE("/booking/:name/:bookingId", h.RemoveBooking)

		id := bson.NewObjectID().Hex()
		svc.RemoveBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return assertAnyError()
		}

		req := httptest.NewRequest(http.MethodDelete, "/booking/A1/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success returns 200", func(t *testing.T) {
		svc := appmocks.NewMockBookingService()
		h := NewBookingHandler(svc)
		r := gin.New()
		r.DELETE("/booking/:name/:bookingId", h.RemoveBooking)

		id := bson.NewObjectID().Hex()
		svc.RemoveBookingFunc = func(_ context.Context, _ string, _ bson.ObjectID) error {
			return nil
		}

		req := httptest.NewRequest(http.MethodDelete, "/booking/A1/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}
