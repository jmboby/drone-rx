package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthChecker is the interface used by HealthHandler to probe service dependencies.
type HealthChecker interface {
	PingDB() error
	PingNATS() error
}

// HealthHandler serves the /healthz endpoint.
type HealthHandler struct {
	checker HealthChecker
}

// NewHealthHandler constructs a HealthHandler with the given checker.
func NewHealthHandler(checker HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// ServeHTTP handles GET /healthz. It probes DB and NATS and returns JSON.
// Returns 200 when all checks pass, 503 when any check fails.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body := map[string]string{
		"status": "ok",
		"db":     "ok",
		"nats":   "ok",
	}

	healthy := true

	if err := h.checker.PingDB(); err != nil {
		body["db"] = "error"
		healthy = false
	}

	if err := h.checker.PingNATS(); err != nil {
		body["nats"] = "error"
		healthy = false
	}

	if !healthy {
		body["status"] = "error"
	}

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
