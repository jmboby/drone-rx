package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminHandler_GenerateSupportBundle(t *testing.T) {
	t.Run("returns ok when command succeeds", func(t *testing.T) {
		h := NewAdminHandler("default", "echo", []string{"bundle-generated"})

		req := httptest.NewRequest(http.MethodPost, "/api/admin/support-bundle", nil)
		w := httptest.NewRecorder()

		h.GenerateSupportBundle(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var body map[string]string
		json.NewDecoder(w.Body).Decode(&body)
		if body["status"] != "ok" {
			t.Fatalf("expected status ok, got %s", body["status"])
		}
	})

	t.Run("returns error when command fails", func(t *testing.T) {
		h := NewAdminHandler("default", "false", nil)

		req := httptest.NewRequest(http.MethodPost, "/api/admin/support-bundle", nil)
		w := httptest.NewRecorder()

		h.GenerateSupportBundle(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}

		var body map[string]string
		json.NewDecoder(w.Body).Decode(&body)
		if body["status"] != "error" {
			t.Fatalf("expected status error, got %s", body["status"])
		}
	})

	t.Run("rejects non-POST methods", func(t *testing.T) {
		h := NewAdminHandler("default", "echo", []string{"ok"})

		req := httptest.NewRequest(http.MethodGet, "/api/admin/support-bundle", nil)
		w := httptest.NewRecorder()

		h.GenerateSupportBundle(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", w.Code)
		}
	})
}
