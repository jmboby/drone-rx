package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/handlers"
)

type mockHealthChecker struct{ dbOK bool; natsOK bool }

func (m *mockHealthChecker) PingDB() error {
	if !m.dbOK {
		return fmt.Errorf("db down")
	}
	return nil
}

func (m *mockHealthChecker) PingNATS() error {
	if !m.natsOK {
		return fmt.Errorf("nats down")
	}
	return nil
}

func TestHealthHandler_AllHealthy(t *testing.T) {
	h := handlers.NewHealthHandler(&mockHealthChecker{dbOK: true, natsOK: true})
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
	if body["db"] != "ok" {
		t.Errorf("expected db ok, got %s", body["db"])
	}
	if body["nats"] != "ok" {
		t.Errorf("expected nats ok, got %s", body["nats"])
	}
}

func TestHealthHandler_DBDown(t *testing.T) {
	h := handlers.NewHealthHandler(&mockHealthChecker{dbOK: false, natsOK: true})
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "error" {
		t.Errorf("expected status error, got %s", body["status"])
	}
	if body["db"] != "error" {
		t.Errorf("expected db error, got %s", body["db"])
	}
}

func TestHealthHandler_NATSDown(t *testing.T) {
	h := handlers.NewHealthHandler(&mockHealthChecker{dbOK: true, natsOK: false})
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "error" {
		t.Errorf("expected status error, got %s", body["status"])
	}
	if body["nats"] != "error" {
		t.Errorf("expected nats error, got %s", body["nats"])
	}
}
