package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jwilson/dronerx/internal/sdk"
)

// LicenseHandler serves license status information via the Replicated SDK.
type LicenseHandler struct {
	client *sdk.Client
}

// NewLicenseHandler creates a LicenseHandler backed by the given SDK client.
func NewLicenseHandler(client *sdk.Client) *LicenseHandler {
	return &LicenseHandler{client: client}
}

type licenseStatusResponse struct {
	Valid               bool   `json:"valid"`
	Expired             bool   `json:"expired"`
	LicenseType         string `json:"license_type"`
	ExpirationDate      string `json:"expiration_date"`
	LiveTrackingEnabled bool   `json:"live_tracking_enabled"`
}

// Status handles GET /api/license/status.
func (h *LicenseHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	info, err := h.client.GetLicenseInfo()
	if err != nil {
		// SDK unavailable — fail open for validity, disable features
		json.NewEncoder(w).Encode(licenseStatusResponse{
			Valid:               true,
			Expired:             false,
			LiveTrackingEnabled: false,
		})
		return
	}

	liveTracking := h.client.IsFeatureEnabled("live_tracking_enabled")

	json.NewEncoder(w).Encode(licenseStatusResponse{
		Valid:               true,
		Expired:             info.IsExpired,
		LicenseType:         info.LicenseType,
		ExpirationDate:      info.ExpirationDate,
		LiveTrackingEnabled: liveTracking,
	})
}
