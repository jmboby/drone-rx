package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/handlers"
)

func TestUIConfig_Truthy(t *testing.T) {
	cases := []struct {
		name             string
		lightMode        string
		adminLink        string
		wantLight        bool
		wantAdminVisible bool
	}{
		{"both true", "true", "true", true, true},
		{"both one", "1", "1", true, true},
		{"both false", "false", "false", false, false},
		{"mixed", "true", "false", true, false},
		{"empty", "", "", false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := handlers.NewUIConfigHandler(tc.lightMode, tc.adminLink)
			req := httptest.NewRequest(http.MethodGet, "/api/config/ui", nil)
			rr := httptest.NewRecorder()
			h.Get(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rr.Code)
			}
			var resp struct {
				LightModeEnabled bool `json:"light_mode_enabled"`
				AdminLinkVisible bool `json:"admin_link_visible"`
			}
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.LightModeEnabled != tc.wantLight {
				t.Errorf("light_mode_enabled: got %v want %v", resp.LightModeEnabled, tc.wantLight)
			}
			if resp.AdminLinkVisible != tc.wantAdminVisible {
				t.Errorf("admin_link_visible: got %v want %v", resp.AdminLinkVisible, tc.wantAdminVisible)
			}
		})
	}
}
