# Tier 3 Phase 2: Admin Page + Support Bundle Generation — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an `/admin` page with a "Generate Support Bundle" button that collects diagnostics and uploads them to the Vendor Portal via the Replicated SDK.

**Architecture:** The Go API gets a new handler that execs the `support-bundle` CLI as a subprocess. The CLI discovers specs from Kubernetes Secrets, collects data, and auto-uploads to the SDK. The frontend gets a new `/admin` route with a button, confirmation, and spinner. Helm templates add RBAC and the `POD_NAMESPACE` env var.

**Tech Stack:** Go (os/exec), SvelteKit (Svelte 5 runes), Helm templates, troubleshoot `support-bundle` CLI

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `Dockerfile.api` | Modify | Add `support-bundle` CLI binary |
| `internal/config/config.go` | Modify | Add `Namespace` field |
| `internal/handlers/admin.go` | Create | Support bundle generation handler |
| `internal/handlers/admin_test.go` | Create | Unit test for handler |
| `cmd/api/main.go` | Modify | Register admin route, pass namespace |
| `chart/templates/rbac.yaml` | Create | ServiceAccount, Role, ClusterRole, bindings |
| `chart/templates/api-deployment.yaml` | Modify | Add serviceAccountName, POD_NAMESPACE env |
| `frontend/src/lib/api.ts` | Modify | Add `generateSupportBundle` function |
| `frontend/src/lib/types.ts` | Modify | Add `SupportBundleResponse` type |
| `frontend/src/routes/admin/+page.svelte` | Create | Admin page with support bundle button |

---

### Task 1: Add `support-bundle` CLI to the API Docker image

**Files:**
- Modify: `Dockerfile.api`

- [ ] **Step 1: Add a download stage and copy the binary**

Replace the full contents of `Dockerfile.api` with:

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api/

FROM alpine:3.19 AS support-bundle
RUN wget -qO /tmp/support-bundle.tar.gz https://github.com/replicatedhq/troubleshoot/releases/download/v0.121.1/support-bundle_linux_amd64.tar.gz \
    && tar -xzf /tmp/support-bundle.tar.gz -C /usr/local/bin support-bundle \
    && rm /tmp/support-bundle.tar.gz

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /api /api
COPY --from=support-bundle /usr/local/bin/support-bundle /usr/local/bin/support-bundle
EXPOSE 8080
ENTRYPOINT ["/api"]
```

- [ ] **Step 2: Verify the image builds locally**

Run: `docker build -f Dockerfile.api -t dronerx-api:test --platform linux/amd64 .`
Expected: Build succeeds.

Run: `docker run --rm dronerx-api:test support-bundle version 2>&1 | head -1`
Expected: Shows a version string (e.g., `Replicated Troubleshoot ...`)

- [ ] **Step 3: Commit**

```bash
git add Dockerfile.api
git commit -m "feat: add support-bundle CLI to API container image"
```

---

### Task 2: Add Namespace to config and create admin handler

**Files:**
- Modify: `internal/config/config.go:8-15`
- Create: `internal/handlers/admin.go`
- Create: `internal/handlers/admin_test.go`

- [ ] **Step 1: Add Namespace field to config**

In `internal/config/config.go`, add `Namespace` to the `Config` struct and `Load` function.

Add to the struct (after `SDKUrl`):
```go
	Namespace      string
```

Add to the `Load` return (after `SDKUrl` line):
```go
		Namespace:      getEnv("POD_NAMESPACE", "default"),
```

- [ ] **Step 2: Write the test for the admin handler**

Create `internal/handlers/admin_test.go`:

```go
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
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/handlers/ -run TestAdminHandler -v`
Expected: Fails — `NewAdminHandler` undefined.

- [ ] **Step 4: Implement the admin handler**

Create `internal/handlers/admin.go`:

```go
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
```

- [ ] **Step 5: Run the tests**

Run: `go test ./internal/handlers/ -run TestAdminHandler -v`
Expected: All 3 tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/config/config.go internal/handlers/admin.go internal/handlers/admin_test.go
git commit -m "feat: add admin handler for support bundle generation

- AdminHandler execs support-bundle CLI with --load-cluster-specs --auto-upload
- Configurable command for testability
- 120s timeout, structured JSON response
- POD_NAMESPACE from config"
```

---

### Task 3: Register the admin route in main.go

**Files:**
- Modify: `cmd/api/main.go:146-164`

- [ ] **Step 1: Add the admin handler creation after the updates handler (line 152)**

After the line `updatesHandler := handlers.NewUpdatesHandler(sdkClient)`, add:

```go
	adminHandler := handlers.NewAdminHandler(cfg.Namespace, "support-bundle", nil)
```

- [ ] **Step 2: Register the route after the updates route (line 164)**

After the line `mux.HandleFunc("GET /api/updates", updatesHandler.Check)`, add:

```go
	mux.HandleFunc("POST /api/admin/support-bundle", adminHandler.GenerateSupportBundle)
```

- [ ] **Step 3: Verify it compiles**

Run: `go build ./cmd/api/`
Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat: register POST /api/admin/support-bundle route"
```

---

### Task 4: Add RBAC and POD_NAMESPACE to Helm templates

**Files:**
- Create: `chart/templates/rbac.yaml`
- Modify: `chart/templates/api-deployment.yaml`

- [ ] **Step 1: Create the RBAC template**

Create `chart/templates/rbac.yaml`:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "dronerx.fullname" . }}-api
  labels:
    {{- include "dronerx.api.labels" . | nindent 4 }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ include "dronerx.fullname" . }}-api
  labels:
    {{- include "dronerx.api.labels" . | nindent 4 }}
rules:
  - apiGroups: [""]
    resources: ["secrets", "configmaps", "pods", "pods/log", "services", "events"]
    verbs: ["get", "list"]
  - apiGroups: ["apps"]
    resources: ["deployments", "statefulsets", "replicasets"]
    verbs: ["get", "list"]
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{ include "dronerx.fullname" . }}-api
  labels:
    {{- include "dronerx.api.labels" . | nindent 4 }}
rules:
  - apiGroups: [""]
    resources: ["nodes", "namespaces"]
    verbs: ["get", "list"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ include "dronerx.fullname" . }}-api
  labels:
    {{- include "dronerx.api.labels" . | nindent 4 }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ include "dronerx.fullname" . }}-api
subjects:
  - kind: ServiceAccount
    name: {{ include "dronerx.fullname" . }}-api
    namespace: {{ .Release.Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{ include "dronerx.fullname" . }}-api
  labels:
    {{- include "dronerx.api.labels" . | nindent 4 }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ include "dronerx.fullname" . }}-api
subjects:
  - kind: ServiceAccount
    name: {{ include "dronerx.fullname" . }}-api
    namespace: {{ .Release.Namespace }}
```

- [ ] **Step 2: Add serviceAccountName and POD_NAMESPACE to the API deployment**

In `chart/templates/api-deployment.yaml`, add `serviceAccountName` to the pod spec. After the `imagePullSecrets` block (line 18), add:

```yaml
      serviceAccountName: {{ include "dronerx.fullname" . }}-api
```

Add `POD_NAMESPACE` to the env block. After the `DATABASE_URL` env var (line 55), add:

```yaml
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
```

- [ ] **Step 3: Verify helm lint and template**

Run: `helm lint ./chart`
Expected: Pass.

Run: `helm template drone-rx ./chart 2>&1 | grep -c "kind: ServiceAccount\|kind: Role\|kind: ClusterRole\|kind: RoleBinding\|kind: ClusterRoleBinding"`
Expected: `5` (or more if SDK has its own)

Run: `helm template drone-rx ./chart 2>&1 | grep "POD_NAMESPACE"`
Expected: Shows the env var.

Run: `helm template drone-rx ./chart 2>&1 | grep "serviceAccountName"`
Expected: Shows `drone-rx-api`.

- [ ] **Step 4: Commit**

```bash
git add chart/templates/rbac.yaml chart/templates/api-deployment.yaml
git commit -m "feat: add RBAC for support bundle collection and POD_NAMESPACE env

- ServiceAccount, Role, ClusterRole with bindings for API pod
- Role: read secrets, pods, logs, deployments, statefulsets in namespace
- ClusterRole: read nodes, storageclasses, namespaces
- POD_NAMESPACE injected via downward API"
```

---

### Task 5: Add frontend admin page

**Files:**
- Modify: `frontend/src/lib/api.ts`
- Modify: `frontend/src/lib/types.ts`
- Create: `frontend/src/routes/admin/+page.svelte`

- [ ] **Step 1: Add the SupportBundleResponse type**

In `frontend/src/lib/types.ts`, add at the end of the file:

```typescript

export interface SupportBundleResponse {
	status: string;
	message: string;
}
```

- [ ] **Step 2: Add the API function**

In `frontend/src/lib/api.ts`, add the import for `SupportBundleResponse` to the import line:

Change:
```typescript
import type { Medicine, Order, CreateOrderRequest, LicenseStatus, UpdateInfo } from './types';
```
to:
```typescript
import type { Medicine, Order, CreateOrderRequest, LicenseStatus, UpdateInfo, SupportBundleResponse } from './types';
```

Add at the end of the file:

```typescript

export async function generateSupportBundle(): Promise<SupportBundleResponse> {
	return fetchJSON<SupportBundleResponse>(`${BASE_URL}/admin/support-bundle`, {
		method: 'POST',
	});
}
```

- [ ] **Step 3: Create the admin page**

Create `frontend/src/routes/admin/+page.svelte`:

```svelte
<script lang="ts">
	import { generateSupportBundle } from '$lib/api';
	import DroneIcon from '$lib/components/DroneIcon.svelte';

	let loading = $state(false);
	let result = $state<{ status: string; message: string } | null>(null);
	let showConfirm = $state(false);

	async function handleGenerate() {
		showConfirm = false;
		loading = true;
		result = null;

		try {
			const response = await generateSupportBundle();
			result = response;
		} catch (err) {
			result = {
				status: 'error',
				message: err instanceof Error ? err.message : 'An unexpected error occurred',
			};
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Admin — DroneRx</title>
</svelte:head>

<header class="sticky top-0 z-20 border-b border-navy-700/60 bg-navy-900/80 backdrop-blur-xl">
	<div class="max-w-6xl mx-auto px-4 sm:px-6 py-3.5 flex items-center justify-between">
		<div class="flex items-center gap-2.5">
			<span class="text-cyan-glow"><DroneIcon size="w-7 h-7" /></span>
			<span class="text-xl font-bold tracking-tight text-white">DroneRx</span>
			<span class="text-xs text-navy-300 hidden sm:inline ml-1 font-medium">Admin</span>
		</div>
		<nav class="flex items-center gap-4">
			<a href="/" class="text-sm font-medium text-navy-200 hover:text-cyan-glow transition-colors">
				Back to Store
			</a>
		</nav>
	</div>
</header>

<main class="max-w-2xl mx-auto px-4 sm:px-6 py-12">
	<h1 class="text-2xl font-bold text-white mb-2">Admin</h1>
	<p class="text-navy-300 text-sm mb-8">Operational tools for DroneRx administrators.</p>

	<!-- Support Bundle Section -->
	<div class="glass-card rounded-xl border border-navy-700/60 p-6">
		<h2 class="text-lg font-semibold text-white mb-1">Support Bundle</h2>
		<p class="text-navy-300 text-sm mb-5">
			Collect diagnostic data from this cluster and upload it to the Vendor Portal for troubleshooting.
		</p>

		{#if loading}
			<div class="flex items-center gap-3 text-cyan-400">
				<svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
					<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
				</svg>
				<span class="text-sm font-medium">Generating support bundle... This may take a minute.</span>
			</div>
		{:else if showConfirm}
			<div class="bg-amber-500/10 border border-amber-500/30 rounded-lg p-4 mb-4">
				<p class="text-amber-200 text-sm mb-3">
					This will collect diagnostic data from this cluster and upload it to the vendor. Continue?
				</p>
				<div class="flex gap-3">
					<button
						onclick={handleGenerate}
						class="bg-cyan-glow/15 hover:bg-cyan-glow/25 text-cyan-glow text-sm font-semibold px-4 py-2 rounded-lg border border-cyan-glow/30 transition-all"
					>
						Yes, generate
					</button>
					<button
						onclick={() => { showConfirm = false; }}
						class="text-navy-300 hover:text-navy-100 text-sm font-medium px-4 py-2 rounded-lg border border-navy-600 transition-all"
					>
						Cancel
					</button>
				</div>
			</div>
		{:else}
			<button
				onclick={() => { showConfirm = true; }}
				class="bg-cyan-glow/15 hover:bg-cyan-glow/25 text-cyan-glow text-sm font-semibold px-4 py-2 rounded-lg border border-cyan-glow/30 transition-all"
			>
				Generate Support Bundle
			</button>
		{/if}

		{#if result}
			<div class="mt-4 rounded-lg p-4 {result.status === 'ok' ? 'bg-emerald-500/10 border border-emerald-500/30' : 'bg-red-500/10 border border-red-500/30'}">
				<p class="text-sm {result.status === 'ok' ? 'text-emerald-300' : 'text-red-300'}">
					{result.message}
				</p>
			</div>
		{/if}
	</div>
</main>
```

- [ ] **Step 4: Verify the frontend builds**

Run: `cd frontend && npm run build`
Expected: Build succeeds.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/types.ts frontend/src/lib/api.ts frontend/src/routes/admin/+page.svelte
git commit -m "feat: add /admin page with support bundle generation button

- Confirmation dialog before triggering
- Spinner during collection
- Success/error feedback
- Consistent glass-card styling"
```

---

### Task 6: Full integration verification

- [ ] **Step 1: Run all Go tests**

Run: `go test ./... -v`
Expected: All tests pass including new admin handler tests.

- [ ] **Step 2: Run frontend build**

Run: `cd frontend && npm run build`
Expected: Build succeeds.

- [ ] **Step 3: Verify Helm templates**

Run: `helm lint ./chart`
Expected: Pass.

Run: `helm template drone-rx ./chart 2>&1 | grep "serviceAccountName.*api"`
Expected: `serviceAccountName: drone-rx-api`

Run: `helm template drone-rx ./chart 2>&1 | grep "POD_NAMESPACE" -A3`
Expected: Shows the downward API fieldRef.

Run: `helm template drone-rx ./chart 2>&1 | grep -c "kind: Role"`
Expected: At least `2` (Role + ClusterRole).

- [ ] **Step 4: Verify RBAC permissions are sufficient**

Run: `helm template drone-rx ./chart --show-only templates/rbac.yaml 2>&1 | grep "resources:" -A1`
Expected: Shows secrets, pods, pods/log, deployments, nodes, storageclasses.
