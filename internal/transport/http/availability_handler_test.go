package http

import (
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

	"github.com/gin-gonic/gin"
)

func TestAvailabilityHandler_GetAvailablePeriods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200 with periods", func(t *testing.T) {
		svc := fakes.NewFakeAvailabilityService()
		h := NewAvailabilityHandler(svc)
		r := gin.New()
		r.GET("/cottage/:name/available-dates", h.GetAvailablePeriods)

		from := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)
		expected := domain.CottageAvailablePeriod{Name: "A1", Periods: []domain.Period{{CheckIn: from, CheckOut: to}}}
		svc.GetAvailablePeriodsFunc = func(_ context.Context, _ string, _ domain.Period) (domain.CottageAvailablePeriod, error) {
			return expected, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/cottage/A1/available-dates?from=2025-09-01&to=2025-09-30", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
		}
		var got dto.AvailablePeriodDTO
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if got.Name != "A1" || len(got.Periods) != 1 || !got.Periods[0].CheckIn.Equal(from) || !got.Periods[0].CheckOut.Equal(to) {
			t.Fatalf("unexpected body: %+v", got)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := fakes.NewFakeAvailabilityService()
		h := NewAvailabilityHandler(svc)
		r := gin.New()
		r.GET("/cottage/:name/available-dates", h.GetAvailablePeriods)

		svc.GetAvailablePeriodsFunc = func(_ context.Context, _ string, _ domain.Period) (domain.CottageAvailablePeriod, error) {
			return domain.CottageAvailablePeriod{}, errors.New("any error")
		}

		req := httptest.NewRequest(http.MethodGet, "/cottage/A1/available-dates?from=2025-09-01&to=2025-09-30", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("binding error returns 400 (missing query)", func(t *testing.T) {
		svc := fakes.NewFakeAvailabilityService()
		h := NewAvailabilityHandler(svc)
		r := gin.New()
		r.GET("/cottage/:name/available-dates", h.GetAvailablePeriods)

		// Missing 'to' query param
		req := httptest.NewRequest(http.MethodGet, "/cottage/A1/available-dates?from=2025-09-01", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}

func TestAvailabilityHandler_GetAvailablePeriodsByCottageType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success returns 200 with available periods per cottage", func(t *testing.T) {
		svc := fakes.NewFakeAvailabilityService()
		h := NewAvailabilityHandler(svc)
		r := gin.New()
		r.GET("/cottage/type/:cottageType/available-dates", h.GetAvailablePeriodsByCottageType)

		from := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)
		expected := []domain.CottageAvailablePeriod{{Name: "A1", Periods: []domain.Period{{CheckIn: from, CheckOut: to}}}}
		svc.GetAvailablePeriodsByCottageTypeFunc = func(_ context.Context, _ string, _ domain.Period) ([]domain.CottageAvailablePeriod, error) {
			return expected, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/cottage/type/lux/available-dates?from=2025-09-01&to=2025-09-30", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var got []domain.CottageAvailablePeriod
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if len(got) != 1 || got[0].Name != "A1" || len(got[0].Periods) != 1 || !got[0].Periods[0].CheckIn.Equal(from) {
			t.Fatalf("unexpected body: %+v", got)
		}
	})

	t.Run("binding error returns 400 (from>to)", func(t *testing.T) {
		svc := fakes.NewFakeAvailabilityService()
		h := NewAvailabilityHandler(svc)
		r := gin.New()
		r.GET("/cottage/type/:cottageType/available-dates", h.GetAvailablePeriodsByCottageType)

		// from after to
		req := httptest.NewRequest(http.MethodGet, "/cottage/type/lux/available-dates?from=2025-10-01&to=2025-09-01", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}
