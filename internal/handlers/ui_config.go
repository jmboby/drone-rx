package handlers

import (
	"encoding/json"
	"net/http"
)

// UIConfigHandler serves plain UI config values (not license-gated).
// Values are read from env vars at init time — no SDK calls.
type UIConfigHandler struct {
	lightModeEnabled bool
	adminLinkVisible bool
}

// NewUIConfigHandler constructs the handler from string env-var values.
// Accepts "true" or "1" as truthy; anything else is false.
func NewUIConfigHandler(lightMode, adminLink string) *UIConfigHandler {
	return &UIConfigHandler{
		lightModeEnabled: truthy(lightMode),
		adminLinkVisible: truthy(adminLink),
	}
}

type uiConfigResponse struct {
	LightModeEnabled bool `json:"light_mode_enabled"`
	AdminLinkVisible bool `json:"admin_link_visible"`
}

// Get handles GET /api/config/ui.
func (h *UIConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(uiConfigResponse{
		LightModeEnabled: h.lightModeEnabled,
		AdminLinkVisible: h.adminLinkVisible,
	})
}

func truthy(v string) bool {
	return v == "true" || v == "1"
}
