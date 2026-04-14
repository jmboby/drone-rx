package sdk_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/sdk"
)

func TestClient_GetLicenseField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/license/fields/live_tracking_enabled" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sdk.LicenseField{
			Name:      "live_tracking_enabled",
			Value:     "true",
			ValueType: "Boolean",
		})
	}))
	defer srv.Close()

	client := sdk.NewClient(srv.URL)
	field, err := client.GetLicenseField("live_tracking_enabled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field.Name != "live_tracking_enabled" {
		t.Errorf("expected name live_tracking_enabled, got %s", field.Name)
	}
	if field.Value != "true" {
		t.Errorf("expected value true, got %s", field.Value)
	}
	if field.ValueType != "Boolean" {
		t.Errorf("expected ValueType Boolean, got %s", field.ValueType)
	}
}

func TestClient_GetLicenseInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/license/info" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"licenseID": "abc123",
			"channelName": "Stable",
			"licenseType": "prod",
			"entitlements": {
				"expires_at": {"title": "Expiration", "value": "2027-01-01T00:00:00Z", "valueType": "String"}
			}
		}`))
	}))
	defer srv.Close()

	client := sdk.NewClient(srv.URL)
	info, err := client.GetLicenseInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.LicenseID != "abc123" {
		t.Errorf("expected LicenseID abc123, got %s", info.LicenseID)
	}
	if info.ChannelName != "Stable" {
		t.Errorf("expected ChannelName Stable, got %s", info.ChannelName)
	}
	if info.LicenseType != "prod" {
		t.Errorf("expected LicenseType prod, got %s", info.LicenseType)
	}
	if info.IsExpired() {
		t.Error("expected not expired (expires 2027)")
	}
	if info.ExpirationDate() != "2027-01-01T00:00:00Z" {
		t.Errorf("expected expiration date, got %s", info.ExpirationDate())
	}
}

func TestClient_LicenseInfo_Expired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"licenseID": "expired123",
			"licenseType": "trial",
			"entitlements": {
				"expires_at": {"title": "Expiration", "value": "2020-01-01T00:00:00Z", "valueType": "String"}
			}
		}`))
	}))
	defer srv.Close()

	client := sdk.NewClient(srv.URL)
	info, err := client.GetLicenseInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.IsExpired() {
		t.Error("expected expired (expires 2020)")
	}
}

func TestClient_GetUpdates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/app/updates" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]sdk.UpdateInfo{
			{VersionLabel: "1.2.0", CreatedAt: "2026-03-01", ReleaseNotes: "Bug fixes"},
		})
	}))
	defer srv.Close()

	client := sdk.NewClient(srv.URL)
	updates, err := client.GetUpdates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0].VersionLabel != "1.2.0" {
		t.Errorf("expected VersionLabel 1.2.0, got %s", updates[0].VersionLabel)
	}
}

func TestClient_SendMetrics(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/app/custom-metrics" {
			http.NotFound(w, r)
			return
		}
		called = true
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		if _, ok := body["data"]; !ok {
			http.Error(w, "missing data key", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := sdk.NewClient(srv.URL)
	err := client.SendMetrics(map[string]interface{}{"orders_placed": 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected server to be called")
	}
}

func TestIsFeatureEnabled_EnvFallback(t *testing.T) {
	// SDK server that returns 500 (unreachable)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := sdk.NewClient(srv.URL)

	// No env override — SDK fails, returns false (fail closed)
	got := client.IsFeatureEnabled("live_tracking_enabled")
	if got != false {
		t.Errorf("expected false when SDK down and no override, got %v", got)
	}

	// With env override — SDK fails but env says enabled
	client.SetFeatureOverride("live_tracking_enabled", "true")
	got = client.IsFeatureEnabled("live_tracking_enabled")
	if got != true {
		t.Errorf("expected true with env override, got %v", got)
	}

	// With env override disabled
	client.SetFeatureOverride("live_tracking_enabled", "false")
	got = client.IsFeatureEnabled("live_tracking_enabled")
	if got != false {
		t.Errorf("expected false with env override false, got %v", got)
	}
}

func TestClient_SDKUnavailable_FailsGracefully(t *testing.T) {
	// Point at a server that is immediately closed — nothing listening.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close before any request

	client := sdk.NewClient(srv.URL)

	// SendMetrics must return nil (fail silently)
	if err := client.SendMetrics(map[string]interface{}{"key": "val"}); err != nil {
		t.Errorf("SendMetrics should return nil on error, got %v", err)
	}

	// GetLicenseInfo must return an error (not nil)
	_, err := client.GetLicenseInfo()
	if err == nil {
		t.Error("expected GetLicenseInfo to return error when SDK unavailable")
	}
}
