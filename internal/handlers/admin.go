package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os/exec"
	"time"
)

// AdminHandler handles administrative operations.
type AdminHandler struct {
	namespace string
	cmdName   string
	cmdArgs   []string
}

// NewAdminHandler creates an AdminHandler. The cmdName and cmdArgs parameters
// allow injecting a mock command for testing. For production use, pass
// "support-bundle" and nil.
func NewAdminHandler(namespace string, cmdName string, cmdArgs []string) *AdminHandler {
	return &AdminHandler{
		namespace: namespace,
		cmdName:   cmdName,
		cmdArgs:   cmdArgs,
	}
}

// GenerateSupportBundle handles POST /api/admin/support-bundle.
// It execs the support-bundle CLI to collect and upload diagnostics.
func (h *AdminHandler) GenerateSupportBundle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slog.Info("generating support bundle", "namespace", h.namespace)

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	args := h.cmdArgs
	if args == nil {
		args = []string{
			"--load-cluster-specs",
			"--auto-upload",
			"-n", h.namespace,
		}
	}

	cmd := exec.CommandContext(ctx, h.cmdName, args...)
	output, err := cmd.CombinedOutput()

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		slog.Error("support bundle generation failed", "error", err, "output", string(output))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Support bundle generation failed: " + err.Error(),
		})
		return
	}

	slog.Info("support bundle generated", "output", string(output))
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Support bundle generated and uploaded to Vendor Portal",
	})
}
