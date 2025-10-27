package booking

import (
	"bytes"
	appmocks "cottageManager/internal/app/mocks"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupGin() {
	gin.SetMode(gin.TestMode)
}

func TestHandler_AddBooking(t *testing.T) {
	setupGin()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validBody := RequestDto{
		GuestId:        primitive.NewObjectID().Hex(),
		NumberOfGuests: 2,
		CheckInDate:    time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		CheckOutDate:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
	}

	t.Run("success returns 200", func(t *testing.T) {
		svc := appmocks.NewMockBookingService(ctrl)
		h := NewHandler(svc)
		r := gin.New()
		r.POST("/booking/:name", h.AddBooking)

		svc.EXPECT().AddBooking(gomock.Any(), gomock.Any()).Return("id123", nil)

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
		svc := appmocks.NewMockBookingService(ctrl)
		h := NewHandler(svc)
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
		svc := appmocks.NewMockBookingService(ctrl)
		h := NewHandler(svc)
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
		svc := appmocks.NewMockBookingService(ctrl)
		h := NewHandler(svc)
		r := gin.New()
		r.POST("/booking/:name", h.AddBooking)

		svc.EXPECT().AddBooking(gomock.Any(), gomock.Any()).Return("", assertAnyError())

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("invalid bookingId returns 400", func(t *testing.T) {
		svc := appmocks.NewMockBookingService(ctrl)
		h := NewHandler(svc)
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
		svc := appmocks.NewMockBookingService(ctrl)
		h := NewHandler(svc)
		r := gin.New()
		r.DELETE("/booking/:name/:bookingId", h.RemoveBooking)

		id := primitive.NewObjectID().Hex()
		svc.EXPECT().RemoveBooking(gomock.Any(), "A1", gomock.Any()).Return(assertAnyError())

		req := httptest.NewRequest(http.MethodDelete, "/booking/A1/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success returns 200", func(t *testing.T) {
		svc := appmocks.NewMockBookingService(ctrl)
		h := NewHandler(svc)
		r := gin.New()
		r.DELETE("/booking/:name/:bookingId", h.RemoveBooking)

		id := primitive.NewObjectID().Hex()
		svc.EXPECT().RemoveBooking(gomock.Any(), "A1", gomock.Any()).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/booking/A1/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}

// re-use helper from availability tests
type assertError string

func (e assertError) Error() string { return string(e) }

func assertAnyError() error { return assertError("any error") }
