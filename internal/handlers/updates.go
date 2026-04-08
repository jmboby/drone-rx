package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jwilson/dronerx/internal/sdk"
)

// UpdatesHandler serves available application update information via the Replicated SDK.
type UpdatesHandler struct {
	client *sdk.Client
}

// NewUpdatesHandler creates an UpdatesHandler backed by the given SDK client.
func NewUpdatesHandler(client *sdk.Client) *UpdatesHandler {
	return &UpdatesHandler{client: client}
}

// Check handles GET /api/updates.
func (h *UpdatesHandler) Check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	updates, err := h.client.GetUpdates()
	if err != nil {
		// On error, return empty array
		json.NewEncoder(w).Encode([]sdk.UpdateInfo{})
		return
	}

	if updates == nil {
		updates = []sdk.UpdateInfo{}
	}

	json.NewEncoder(w).Encode(updates)
}
