package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/handlers"
	"github.com/jwilson/dronerx/internal/sdk"
)

func TestLicenseStatus_IncludesLightMode(t *testing.T) {
	sdkSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/license/info":
			w.Write([]byte(`{
				"licenseID": "test-123",
				"channelName": "Stable",
				"licenseType": "prod",
				"entitlements": {
					"expires_at": {"title": "Expiration", "value": "2027-01-01T00:00:00Z", "valueType": "String"}
				}
			}`))
		case "/api/v1/license/fields/live_tracking_enabled":
			json.NewEncoder(w).Encode(sdk.LicenseField{Name: "live_tracking_enabled", Value: true, ValueType: "Boolean"})
		case "/api/v1/license/fields/light_mode_enabled":
			json.NewEncoder(w).Encode(sdk.LicenseField{Name: "light_mode_enabled", Value: true, ValueType: "Boolean"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer sdkSrv.Close()

	client := sdk.NewClient(sdkSrv.URL)
	handler := handlers.NewLicenseHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/api/license/status", nil)
	rr := httptest.NewRecorder()
	handler.Status(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		Valid               bool   `json:"valid"`
		Expired             bool   `json:"expired"`
		LicenseType         string `json:"license_type"`
		LiveTrackingEnabled bool   `json:"live_tracking_enabled"`
		LightModeEnabled    bool   `json:"light_mode_enabled"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp.LightModeEnabled {
		t.Error("expected light_mode_enabled to be true")
	}
	if !resp.LiveTrackingEnabled {
		t.Error("expected live_tracking_enabled to be true")
	}
	if !resp.Valid {
		t.Error("expected valid to be true")
	}
}

func TestLicenseStatus_SDKDown_DefaultsFalse(t *testing.T) {
	sdkSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer sdkSrv.Close()

	client := sdk.NewClient(sdkSrv.URL)
	handler := handlers.NewLicenseHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/api/license/status", nil)
	rr := httptest.NewRecorder()
	handler.Status(rr, req)

	var resp struct {
		LightModeEnabled bool `json:"light_mode_enabled"`
		Valid            bool `json:"valid"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.LightModeEnabled {
		t.Error("expected light_mode_enabled false when SDK down")
	}
	if !resp.Valid {
		t.Error("expected valid true when SDK down (fail open)")
	}
}
