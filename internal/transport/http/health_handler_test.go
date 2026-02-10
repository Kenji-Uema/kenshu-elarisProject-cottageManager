package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type readinessStub struct {
	err error
}

func (r readinessStub) Ping(context.Context) error {
	return r.err
}

func TestHealth(t *testing.T) {
	setupGin()

	handler := NewHandler(nil)

	router := gin.New()
	router.GET("/healthz", handler.Health)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("unexpected status payload: %v", payload)
	}
}

func TestLiveness(t *testing.T) {
	setupGin()

	handler := NewHandler(nil)

	router := gin.New()
	router.GET("/livez", handler.Liveness)

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if payload["status"] != "alive" {
		t.Fatalf("unexpected status payload: %v", payload)
	}
}

func TestReadinessSuccess(t *testing.T) {
	setupGin()

	handler := NewHandler(readinessStub{err: nil})

	router := gin.New()
	router.GET("/readyz", handler.Readiness)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if payload["status"] != "ready" {
		t.Fatalf("unexpected status payload: %v", payload)
	}
}

func TestReadinessFailure(t *testing.T) {
	setupGin()

	handler := NewHandler(readinessStub{err: errors.New("db unavailable")})

	router := gin.New()
	router.GET("/readyz", handler.Readiness)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, resp.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if payload["status"] != "unavailable" {
		t.Fatalf("unexpected status payload: %v", payload)
	}
}
