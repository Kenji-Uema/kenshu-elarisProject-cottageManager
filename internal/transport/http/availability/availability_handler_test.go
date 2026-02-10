package availability

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appmocks "github.com/Kenji-Uema/cottageManager/internal/app/mocks"
	"github.com/Kenji-Uema/cottageManager/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
)

func setupGin() {
	gin.SetMode(gin.TestMode)
}

func TestAvailabilityHandler_GetAvailablePeriods(t *testing.T) {
	setupGin()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success returns 200 with periods", func(t *testing.T) {
		svc := appmocks.NewMockAvailabilityService(ctrl)
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/cottage/:name/available-dates", h.GetAvailablePeriods)

		from := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)
		expected := []domain.Period{{Start: from, End: to}}
		svc.EXPECT().GetAvailablePeriods(gomock.Any(), "A1", gomock.Any()).Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/cottage/A1/available-dates?from=2025-09-01&to=2025-09-30", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
		}
		var got []AvailablePeriodDTO
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if len(got) != 1 || !got[0].From.Equal(from) || !got[0].To.Equal(to) {
			t.Fatalf("unexpected body: %+v", got)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := appmocks.NewMockAvailabilityService(ctrl)
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/availability/:name", h.GetAvailablePeriods)

		svc.EXPECT().GetAvailablePeriods(gomock.Any(), "A1", gomock.Any()).Return(nil, assertAnyError())

		req := httptest.NewRequest(http.MethodGet, "/availability/A1?from=2025-09-01&to=2025-09-30", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("binding error returns 400 (missing query)", func(t *testing.T) {
		svc := appmocks.NewMockAvailabilityService(ctrl)
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/availability/:name", h.GetAvailablePeriods)

		// Missing 'to' query param
		req := httptest.NewRequest(http.MethodGet, "/availability/A1?from=2025-09-01", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestAvailabilityHandler_GetAvailablePeriodsByCottageType(t *testing.T) {
	setupGin()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success returns 200 with available periods per cottage", func(t *testing.T) {
		svc := appmocks.NewMockAvailabilityService(ctrl)
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/availability/type/:type", h.GetAvailablePeriodsByCottageType)

		from := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)
		expected := []domain.CottageAvailablePeriod{{Name: "A1", Periods: []domain.Period{{Start: from, End: to}}}}
		svc.EXPECT().GetAvailablePeriodsByCottageType(gomock.Any(), "lux", gomock.Any()).Return(expected, nil)

		req := httptest.NewRequest(http.MethodGet, "/availability/type/lux?from=2025-09-01&to=2025-09-30", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var got []domain.CottageAvailablePeriod
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if len(got) != 1 || got[0].Name != "A1" || len(got[0].Periods) != 1 || !got[0].Periods[0].Start.Equal(from) {
			t.Fatalf("unexpected body: %+v", got)
		}
	})

	t.Run("binding error returns 400 (from>to)", func(t *testing.T) {
		svc := appmocks.NewMockAvailabilityService(ctrl)
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/availability/type/:type", h.GetAvailablePeriodsByCottageType)

		// from after to
		req := httptest.NewRequest(http.MethodGet, "/availability/type/lux?from=2025-10-01&to=2025-09-01", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

// helper matcher that returns an error for gomock expectation without caring about message
func assertAnyError() error { return assertError("any error") }

type assertError string

func (e assertError) Error() string { return string(e) }
