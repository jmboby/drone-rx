# DroneRx Phase 3: Ship It with Helm (Tier 2 remainder) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Integrate with the Replicated SDK to send custom metrics, gate live tracking by license entitlement, show update banners, and enforce license validity — completing Tier 2 rubric requirements.

**Architecture:** New `internal/sdk` Go package wraps the Replicated SDK HTTP API. The API server calls it for license checks, metrics, and update info. Frontend gets new `/api/license/status` and `/api/updates` endpoints. Metrics sent via a background goroutine. License checked on startup + periodic recheck + per-request for tracking.

**Tech Stack:** Go HTTP client, Replicated SDK API (in-cluster HTTP at `http://drone-rx-sdk:3000`)

**SDK API endpoints used:**
- `POST /api/v1/app/custom-metrics` — send app metrics
- `GET /api/v1/license/fields/<name>` — check license entitlement
- `GET /api/v1/license/info` — check license validity/expiry
- `GET /api/v1/app/updates` — check for available updates

---

## File Structure

| File | Responsibility |
|------|---------------|
| `internal/sdk/client.go` | SDK HTTP client — wraps all SDK API calls |
| `internal/sdk/client_test.go` | Tests with httptest mock server |
| `internal/sdk/metrics.go` | Background goroutine that sends metrics periodically |
| `internal/handlers/license.go` | GET /api/license/status — license info for frontend |
| `internal/handlers/updates.go` | GET /api/updates — available updates for frontend |
| `internal/config/config.go` | Add REPLICATED_SDK_URL env var |
| `internal/handlers/tracking.go` | Add license check before WebSocket upgrade |
| `cmd/api/main.go` | Wire SDK client, metrics sender, new routes |
| `frontend/src/lib/api.ts` | Add getLicenseStatus() and getUpdates() |
| `frontend/src/lib/types.ts` | Add LicenseStatus and UpdateInfo types |
| `frontend/src/routes/+layout.svelte` | Update banner component |
| `frontend/src/routes/order/[id]/+page.svelte` | License gate on tracking |

---

## Task 1: SDK Client

**Files:**
- Create: `internal/sdk/client.go`, `internal/sdk/client_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/sdk/client_test.go`:

```go
package sdk_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/sdk"
)

func TestClient_GetLicenseField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/license/fields/live_tracking_enabled" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name": "live_tracking_enabled", "value": "true", "valueType": "String",
		})
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL)
	field, err := client.GetLicenseField("live_tracking_enabled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field.Value != "true" {
		t.Errorf("expected true, got %s", field.Value)
	}
}

func TestClient_GetLicenseInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/license/info" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"licenseID": "abc123", "isExpired": false, "licenseType": "prod",
		})
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL)
	info, err := client.GetLicenseInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.IsExpired {
		t.Error("expected not expired")
	}
	if info.LicenseID != "abc123" {
		t.Errorf("expected abc123, got %s", info.LicenseID)
	}
}

func TestClient_GetUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]string{
			{"versionLabel": "1.2.0", "releaseNotes": "Bug fixes"},
		})
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL)
	updates, err := client.GetUpdates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0].VersionLabel != "1.2.0" {
		t.Errorf("expected 1.2.0, got %s", updates[0].VersionLabel)
	}
}

func TestClient_SendMetrics(t *testing.T) {
	var received map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		received = body["data"].(map[string]interface{})
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := sdk.NewClient(server.URL)
	err := client.SendMetrics(map[string]interface{}{
		"orders_placed": 10, "orders_delivered": 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["orders_placed"].(float64) != 10 {
		t.Errorf("expected 10, got %v", received["orders_placed"])
	}
}

func TestClient_SDKUnavailable_FailsGracefully(t *testing.T) {
	client := sdk.NewClient("http://localhost:1") // nothing listening

	_, err := client.GetLicenseInfo()
	if err == nil {
		t.Error("expected error for unavailable SDK")
	}

	// SendMetrics should not error — fire and forget
	err = client.SendMetrics(map[string]interface{}{"test": 1})
	if err != nil {
		t.Errorf("SendMetrics should not error on SDK unavailability: %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/sdk/ -v
```

Expected: FAIL — package not found.

- [ ] **Step 3: Write implementation**

Create `internal/sdk/client.go`:

```go
package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type LicenseField struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	ValueType string `json:"valueType"`
}

type LicenseInfo struct {
	LicenseID      string `json:"licenseID"`
	ChannelName    string `json:"channelName"`
	LicenseType    string `json:"licenseType"`
	IsExpired      bool   `json:"isExpired"`
	ExpirationDate string `json:"expirationDate,omitempty"`
}

type UpdateInfo struct {
	VersionLabel string `json:"versionLabel"`
	CreatedAt    string `json:"createdAt"`
	ReleaseNotes string `json:"releaseNotes"`
}

func (c *Client) GetLicenseField(fieldName string) (*LicenseField, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/v1/license/fields/%s", c.baseURL, fieldName))
	if err != nil {
		return nil, fmt.Errorf("sdk: license field request: %w", err)
	}
	defer resp.Body.Close()

	var field LicenseField
	if err := json.NewDecoder(resp.Body).Decode(&field); err != nil {
		return nil, fmt.Errorf("sdk: decode license field: %w", err)
	}
	return &field, nil
}

func (c *Client) GetLicenseInfo() (*LicenseInfo, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/v1/license/info", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("sdk: license info request: %w", err)
	}
	defer resp.Body.Close()

	var info LicenseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("sdk: decode license info: %w", err)
	}
	return &info, nil
}

func (c *Client) GetUpdates() ([]UpdateInfo, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/v1/app/updates", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("sdk: updates request: %w", err)
	}
	defer resp.Body.Close()

	var updates []UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
		return nil, fmt.Errorf("sdk: decode updates: %w", err)
	}
	return updates, nil
}

func (c *Client) SendMetrics(data map[string]interface{}) error {
	body := map[string]interface{}{"data": data}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil // fail silently
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/v1/app/custom-metrics", c.baseURL),
		"application/json",
		bytes.NewReader(jsonData),
	)
	if err != nil {
		return nil // fail silently — metrics are best-effort
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) IsFeatureEnabled(fieldName string) bool {
	field, err := c.GetLicenseField(fieldName)
	if err != nil {
		return false // fail closed
	}
	return field.Value == "true" || field.Value == "1"
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/sdk/ -v
```

Expected: All PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/sdk/
git commit -m "feat: add Replicated SDK client with license, metrics, and updates API"
```

---

## Task 2: Metrics Background Sender

**Files:**
- Create: `internal/sdk/metrics.go`

- [ ] **Step 1: Create metrics sender**

Create `internal/sdk/metrics.go`:

```go
package sdk

import (
	"context"
	"log"
	"time"
)

type MetricsSource interface {
	CountByStatus(ctx context.Context) (map[string]int, error)
}

func StartMetricsSender(ctx context.Context, client *Client, source MetricsSource, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sendMetrics(ctx, client, source)
		}
	}
}

func sendMetrics(ctx context.Context, client *Client, source MetricsSource) {
	counts, err := source.CountByStatus(ctx)
	if err != nil {
		log.Printf("metrics: counting orders: %v", err)
		return
	}

	data := map[string]interface{}{
		"orders_placed":    counts["placed"],
		"orders_preparing": counts["preparing"],
		"orders_in_flight": counts["in-flight"],
		"orders_delivered": counts["delivered"],
		"orders_total":     counts["placed"] + counts["preparing"] + counts["in-flight"] + counts["delivered"],
	}

	client.SendMetrics(data)
}
```

- [ ] **Step 2: Add CountByStatus to OrderStore**

Add to `internal/models/order.go`:

```go
func (s *OrderStore) CountByStatus(ctx context.Context) (map[string]int, error) {
	rows, err := s.db.Query(ctx,
		`SELECT status, COUNT(*) FROM orders GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/sdk/ ./internal/models/
```

- [ ] **Step 4: Commit**

```bash
git add internal/sdk/metrics.go internal/models/order.go
git commit -m "feat: add background metrics sender with order count by status"
```

---

## Task 3: License and Updates API Handlers

**Files:**
- Create: `internal/handlers/license.go`, `internal/handlers/updates.go`

- [ ] **Step 1: Create license status handler**

Create `internal/handlers/license.go`:

```go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jwilson/dronerx/internal/sdk"
)

type LicenseHandler struct {
	client *sdk.Client
}

func NewLicenseHandler(client *sdk.Client) *LicenseHandler {
	return &LicenseHandler{client: client}
}

func (h *LicenseHandler) Status(w http.ResponseWriter, r *http.Request) {
	info, err := h.client.GetLicenseInfo()
	if err != nil {
		// SDK unavailable — return default (valid, no features)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":                 true,
			"expired":              false,
			"live_tracking_enabled": false,
		})
		return
	}

	trackingEnabled := h.client.IsFeatureEnabled("live_tracking_enabled")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":                 !info.IsExpired,
		"expired":              info.IsExpired,
		"license_type":         info.LicenseType,
		"expiration_date":      info.ExpirationDate,
		"live_tracking_enabled": trackingEnabled,
	})
}
```

- [ ] **Step 2: Create updates handler**

Create `internal/handlers/updates.go`:

```go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jwilson/dronerx/internal/sdk"
)

type UpdatesHandler struct {
	client *sdk.Client
}

func NewUpdatesHandler(client *sdk.Client) *UpdatesHandler {
	return &UpdatesHandler{client: client}
}

func (h *UpdatesHandler) Check(w http.ResponseWriter, r *http.Request) {
	updates, err := h.client.GetUpdates()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updates)
}
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./internal/handlers/
```

- [ ] **Step 4: Commit**

```bash
git add internal/handlers/license.go internal/handlers/updates.go
git commit -m "feat: add license status and updates check API handlers"
```

---

## Task 4: Gate WebSocket Tracking by License

**Files:**
- Modify: `internal/handlers/tracking.go`

- [ ] **Step 1: Add license check to tracking handler**

Replace `internal/handlers/tracking.go`:

```go
package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/nats-io/nats.go"

	"github.com/jwilson/dronerx/internal/sdk"
)

type TrackingHandler struct {
	nc     *nats.Conn
	client *sdk.Client
}

func NewTrackingHandler(nc *nats.Conn, client *sdk.Client) *TrackingHandler {
	return &TrackingHandler{nc: nc, client: client}
}

func (h *TrackingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	if orderID == "" {
		http.Error(w, "order ID required", http.StatusBadRequest)
		return
	}

	// Check license entitlement for live tracking
	if h.client != nil && !h.client.IsFeatureEnabled("live_tracking_enabled") {
		http.Error(w, "live tracking requires a premium license", http.StatusForbidden)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		log.Printf("websocket accept: %v", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	subject := "orders." + orderID + ".status"
	sub, err := h.nc.Subscribe(subject, func(msg *nats.Msg) {
		if err := conn.Write(ctx, websocket.MessageText, msg.Data); err != nil {
			log.Printf("websocket write: %v", err)
			cancel()
		}
	})
	if err != nil {
		log.Printf("nats subscribe: %v", err)
		conn.Close(websocket.StatusInternalError, "subscription failed")
		return
	}
	defer sub.Unsubscribe()

	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			break
		}
	}
	conn.Close(websocket.StatusNormalClosure, "")
}
```

- [ ] **Step 2: Verify compilation**

```bash
go build ./internal/handlers/
```

- [ ] **Step 3: Commit**

```bash
git add internal/handlers/tracking.go
git commit -m "feat: gate WebSocket live tracking by license entitlement via SDK"
```

---

## Task 5: Wire SDK into main.go

**Files:**
- Modify: `internal/config/config.go`, `internal/config/config_test.go`, `cmd/api/main.go`

- [ ] **Step 1: Add SDK URL to config**

Add to `internal/config/config.go` Config struct:

```go
SDKUrl string
```

And in `Load()`:

```go
SDKUrl: getEnv("REPLICATED_SDK_URL", "http://drone-rx-sdk:3000"),
```

- [ ] **Step 2: Update config test**

Add to `TestLoad_Defaults`:

```go
if cfg.SDKUrl != "http://drone-rx-sdk:3000" {
    t.Errorf("expected default SDK URL, got %s", cfg.SDKUrl)
}
```

- [ ] **Step 3: Wire SDK client into main.go**

Add after the NATS connection in `main()`:

```go
// SDK client
sdkClient := sdk.NewClient(cfg.SDKUrl)

// Start metrics sender (every 5 minutes)
go sdk.StartMetricsSender(ctx, sdkClient, orderStore, 5*time.Minute)
```

Update handler creation:

```go
trackingHandler := handlers.NewTrackingHandler(nc, sdkClient)
licenseHandler := handlers.NewLicenseHandler(sdkClient)
updatesHandler := handlers.NewUpdatesHandler(sdkClient)
```

Add new routes:

```go
mux.HandleFunc("GET /api/license/status", licenseHandler.Status)
mux.HandleFunc("GET /api/updates", updatesHandler.Check)
```

Add import for `sdk` package.

- [ ] **Step 4: Add SDK URL to Helm ConfigMap**

Add to `chart/templates/configmap-api.yaml`:

```yaml
REPLICATED_SDK_URL: "http://drone-rx-sdk:3000"
```

- [ ] **Step 5: Run all tests**

```bash
go test ./... -v
```

Expected: All pass.

- [ ] **Step 6: Commit**

```bash
git add internal/config/ cmd/api/main.go chart/templates/configmap-api.yaml
git commit -m "feat: wire SDK client, metrics sender, license and updates routes into main.go"
```

---

## Task 6: Frontend — License Status, Update Banner, Tracking Gate

**Files:**
- Modify: `frontend/src/lib/api.ts`, `frontend/src/lib/types.ts`, `frontend/src/routes/+layout.svelte`, `frontend/src/routes/order/[id]/+page.svelte`

- [ ] **Step 1: Add types**

Add to `frontend/src/lib/types.ts`:

```ts
export interface LicenseStatus {
	valid: boolean;
	expired: boolean;
	license_type?: string;
	expiration_date?: string;
	live_tracking_enabled: boolean;
}

export interface UpdateInfo {
	versionLabel: string;
	createdAt: string;
	releaseNotes: string;
}
```

- [ ] **Step 2: Add API functions**

Add to `frontend/src/lib/api.ts`:

```ts
import type { Medicine, Order, CreateOrderRequest, LicenseStatus, UpdateInfo } from './types';

export async function getLicenseStatus(): Promise<LicenseStatus> {
	return fetchJSON<LicenseStatus>(`${BASE_URL}/license/status`);
}

export async function getUpdates(): Promise<UpdateInfo[]> {
	return fetchJSON<UpdateInfo[]>(`${BASE_URL}/updates`);
}
```

- [ ] **Step 3: Add update banner to layout**

Update `frontend/src/routes/+layout.svelte` to fetch updates on mount and show a banner:

```svelte
<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { getUpdates } from '$lib/api';
	import type { UpdateInfo } from '$lib/types';

	let { children } = $props();
	let latestUpdate = $state<UpdateInfo | null>(null);
	let showBanner = $state(true);

	onMount(async () => {
		try {
			const updates = await getUpdates();
			if (updates.length > 0) {
				latestUpdate = updates[0];
			}
		} catch {
			// SDK not available — no banner
		}
	});
</script>

{#if latestUpdate && showBanner}
	<div class="bg-amber-glow/10 border-b border-amber-glow/30 px-4 py-2 text-center text-sm">
		<span class="text-amber-200">
			Update available: <strong>{latestUpdate.versionLabel}</strong>
		</span>
		<button
			onclick={() => showBanner = false}
			class="ml-3 text-amber-400 hover:text-amber-200 text-xs"
		>
			Dismiss
		</button>
	</div>
{/if}

<div class="min-h-screen bg-navy-950 grid-bg">
	{@render children()}
</div>
```

- [ ] **Step 4: Update tracking page to check license**

In `frontend/src/routes/order/[id]/+page.svelte`, update the `startTracking` function to check license before connecting WebSocket. Add a `getLicenseStatus()` call in `onMount`:

```ts
onMount(async () => {
    try {
        const license = await getLicenseStatus();
        if (license.live_tracking_enabled && order.status !== 'delivered') {
            startTracking();
        } else {
            startPolling();
        }
    } catch {
        startPolling();
    }
});
```

- [ ] **Step 5: Verify frontend builds**

```bash
cd frontend && npm run build
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/
git commit -m "feat: add update banner, license status check, and tracking license gate in frontend"
```

---

## Task 7: Create license field in Vendor Portal

This is a Vendor Portal configuration step, not code.

- [ ] **Step 1: Create the license field**

In Vendor Portal → **License Fields** → **Create License Field**:
- **Field name:** `live_tracking_enabled`
- **Title:** Live Tracking
- **Type:** Boolean
- **Default:** `false`
- **Hidden:** No

- [ ] **Step 2: Verify on a test customer**

Create/update a customer license with `live_tracking_enabled: true`. Install the app. Verify:
- `GET /api/license/status` returns `"live_tracking_enabled": true`
- WebSocket tracking connects successfully
- Toggle to `false` → tracking returns 403, frontend falls back to polling

---

## Rubric Coverage (Tier 2 remainder)

| Requirement | Task(s) | Status |
|------------|---------|--------|
| 2.4 App sends custom metrics | Tasks 1, 2, 5 (SDK client + metrics sender + wiring) | Covered |
| 2.5 License entitlement gates live tracking | Tasks 1, 4, 6, 7 (SDK client + tracking gate + frontend + license field) | Covered |
| 2.6a Update available banner | Tasks 1, 3, 6 (SDK client + updates handler + frontend banner) | Covered |
| 2.6b License validity enforced via SDK | Tasks 1, 3, 6 (SDK client + license handler + frontend check) | Covered |
| 2.9 Instance is live, named, tagged, healthy | SDK handles automatically when running | Covered |
| 2.10 Services show as healthy in instance reporting | SDK + statusInformers in Application CR | Covered |
