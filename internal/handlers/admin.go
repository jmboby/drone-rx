package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"time"
)

// AdminHandler handles administrative operations.
type AdminHandler struct {
	namespace string
	sdkURL    string
	cmdName   string
	cmdArgs   []string
}

// NewAdminHandler creates an AdminHandler. The cmdName and cmdArgs parameters
// allow injecting a mock command for testing. For production use, pass
// "sh" and nil (the handler builds a shell script that collects and uploads).
func NewAdminHandler(namespace, sdkURL, cmdName string, cmdArgs []string) *AdminHandler {
	return &AdminHandler{
		namespace: namespace,
		sdkURL:    sdkURL,
		cmdName:   cmdName,
		cmdArgs:   cmdArgs,
	}
}

// GenerateSupportBundle handles POST /api/admin/support-bundle.
// It collects a support bundle via the CLI and uploads it to the SDK.
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
		// Two-step: collect bundle, then upload to local SDK endpoint.
		// --auto-upload targets replicated.app which requires a license ID.
		// Instead, we POST the tarball directly to the SDK's upload endpoint.
		script := fmt.Sprintf(`set -e
BUNDLE=$(support-bundle --load-cluster-specs -n %s -o /tmp/support-bundle.tar.gz 2>&1 && echo /tmp/support-bundle.tar.gz)
wget -q --header="Content-Type: application/gzip" --post-file=/tmp/support-bundle.tar.gz -O - %s/api/v1/supportbundle
rm -f /tmp/support-bundle.tar.gz`,
			h.namespace, h.sdkURL)
		args = []string{"-c", script}
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

	slog.Info("support bundle generated and uploaded", "output", string(output))
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Support bundle generated and uploaded to Vendor Portal",
	})
}
