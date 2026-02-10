package cottage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appmocks "github.com/Kenji-Uema/cottageManager/internal/app/mocks"
	"github.com/Kenji-Uema/cottageManager/internal/domain"
	"github.com/Kenji-Uema/cottageManager/internal/domain/errors/appErrors"

	"github.com/gin-gonic/gin"
)

func setupGin() {
	gin.SetMode(gin.TestMode)
}

func TestHandler_GetAll(t *testing.T) {
	setupGin()

	t.Run("success returns 200 with cottages dto", func(t *testing.T) {
		svc := appmocks.NewMockCottageService()
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/cottages", h.GetAll)

		cottages := []domain.Cottage{{Name: "A1", View: "lux"}}
		svc.GetAllFunc = func(_ context.Context) ([]domain.Cottage, error) {
			return cottages, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/cottages", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var got []Dto
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if len(got) != 1 || got[0].Name != "A1" || got[0].Type != "lux" {
			t.Fatalf("unexpected body: %+v", got)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := appmocks.NewMockCottageService()
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/cottages", h.GetAll)

		svc.GetAllFunc = func(_ context.Context) ([]domain.Cottage, error) {
			return nil, &appErrors.UnexpectedError{}
		}

		req := httptest.NewRequest(http.MethodGet, "/cottages", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestHandler_GetByName(t *testing.T) {
	setupGin()

	t.Run("success returns 200 with dto", func(t *testing.T) {
		svc := appmocks.NewMockCottageService()
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/cottage/:name", h.GetByName)

		svc.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{Name: "A1", View: "lux"}, nil
		}

		req := httptest.NewRequest(http.MethodGet, "/cottage/A1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var got Dto
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if got.Name != "A1" || got.Type != "lux" {
			t.Fatalf("unexpected body: %+v", got)
		}
	})

	t.Run("not found returns 404", func(t *testing.T) {
		svc := appmocks.NewMockCottageService()
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/cottage/:name", h.GetByName)

		svc.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, &appErrors.CottageNotFound{}
		}

		req := httptest.NewRequest(http.MethodGet, "/cottage/A2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("unexpected error returns 500", func(t *testing.T) {
		svc := appmocks.NewMockCottageService()
		h := NewHandler(svc)
		r := gin.New()
		r.GET("/cottage/:name", h.GetByName)

		svc.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, &appErrors.UnexpectedError{}
		}

		req := httptest.NewRequest(http.MethodGet, "/cottage/A3", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}
