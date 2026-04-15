# Light Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a license-gated Light Mode toggle that swaps the dark navy theme to a white/light theme, controlled by a Replicated custom license field.

**Architecture:** The backend exposes a `light_mode_enabled` Boolean via `/api/license/status`, following the existing `live_tracking_enabled` pattern. The frontend reads this flag to conditionally show a sun/moon toggle in the header. CSS variable overrides on `[data-theme="light"]` handle all color swaps without changing component classes.

**Tech Stack:** Go (backend handler + config), Helm (values + ConfigMap), SvelteKit + Tailwind CSS v4 (frontend)

---

## File Structure

| File | Responsibility |
|---|---|
| `internal/config/config.go` | Add `LightModeEnabled` env var |
| `internal/handlers/license.go` | Add `LightModeEnabled` to API response |
| `internal/handlers/license_test.go` | **New** — test license handler response |
| `cmd/api/main.go` | Register `light_mode_enabled` feature override |
| `chart/values.yaml` | Add `lightModeEnabled: "false"` default |
| `chart/templates/configmap-api.yaml` | Add `LIGHT_MODE_ENABLED` env var |
| `frontend/src/lib/types.ts` | Add `light_mode_enabled` to `LicenseStatus` |
| `frontend/src/lib/stores/theme.ts` | **New** — theme store with localStorage |
| `frontend/src/lib/components/ThemeToggle.svelte` | **New** — sun/moon toggle component |
| `frontend/src/app.css` | Add `[data-theme="light"]` variable overrides |
| `frontend/src/routes/+layout.svelte` | Apply `data-theme` to body, expose license |
| `frontend/src/routes/+page.svelte` | Add ThemeToggle to header |
| `frontend/src/routes/order/+page.svelte` | Add ThemeToggle to header |
| `frontend/src/routes/order/[id]/+page.svelte` | Add ThemeToggle to header |
| `frontend/src/routes/orders/+page.svelte` | Add ThemeToggle to header |
| `frontend/src/routes/admin/+page.svelte` | Add ThemeToggle to header |

---

### Task 1: Backend — Config and License Handler

**Files:**
- Modify: `internal/config/config.go:8-29`
- Modify: `internal/handlers/license.go:20-51`
- Modify: `cmd/api/main.go:128-132`
- Create: `internal/handlers/license_test.go`

- [ ] **Step 1: Add `LightModeEnabled` to config struct and Load()**

In `internal/config/config.go`, add the field to the `Config` struct and the `Load()` function:

```go
type Config struct {
	Port                string
	DatabaseURL         string
	NATSUrl             string
	TickerInterval      int
	WebhookURL          string
	SDKUrl              string
	Namespace           string
	LiveTrackingEnabled string
	LightModeEnabled    string
}

func Load() Config {
	return Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		NATSUrl:             getEnv("NATS_URL", "nats://localhost:4222"),
		TickerInterval:      getEnvInt("TICKER_INTERVAL", 5),
		WebhookURL:          getEnv("WEBHOOK_URL", ""),
		SDKUrl:              getEnv("REPLICATED_SDK_URL", "http://drone-rx-sdk:3000"),
		Namespace:           getEnv("POD_NAMESPACE", "default"),
		LiveTrackingEnabled: getEnv("LIVE_TRACKING_ENABLED", "true"),
		LightModeEnabled:    getEnv("LIGHT_MODE_ENABLED", "false"),
	}
}
```

- [ ] **Step 2: Add `LightModeEnabled` to license handler response**

In `internal/handlers/license.go`, add the field to the response struct and populate it:

```go
type licenseStatusResponse struct {
	Valid               bool   `json:"valid"`
	Expired             bool   `json:"expired"`
	LicenseType         string `json:"license_type"`
	ExpirationDate      string `json:"expiration_date"`
	LiveTrackingEnabled bool   `json:"live_tracking_enabled"`
	LightModeEnabled    bool   `json:"light_mode_enabled"`
}
```

In the `Status` method, update the SDK-unavailable fallback (line 35-39) to include `LightModeEnabled: false`, and add the feature check in the success path:

```go
func (h *LicenseHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	info, err := h.client.GetLicenseInfo()
	if err != nil {
		json.NewEncoder(w).Encode(licenseStatusResponse{
			Valid:               true,
			Expired:             false,
			LiveTrackingEnabled: false,
			LightModeEnabled:    false,
		})
		return
	}

	liveTracking := h.client.IsFeatureEnabled("live_tracking_enabled")
	lightMode := h.client.IsFeatureEnabled("light_mode_enabled")

	json.NewEncoder(w).Encode(licenseStatusResponse{
		Valid:               !info.IsExpired(),
		Expired:             info.IsExpired(),
		LicenseType:         info.LicenseType,
		ExpirationDate:      info.ExpirationDate(),
		LiveTrackingEnabled: liveTracking,
		LightModeEnabled:    lightMode,
	})
}
```

- [ ] **Step 3: Register feature override in main.go**

In `cmd/api/main.go`, after the existing `live_tracking_enabled` override (line 130-132), add:

```go
if cfg.LightModeEnabled != "" {
	sdkClient.SetFeatureOverride("light_mode_enabled", cfg.LightModeEnabled)
}
```

- [ ] **Step 4: Write license handler test**

Create `internal/handlers/license_test.go`:

```go
package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/handlers"
	"github.com/jwilson/dronerx/internal/sdk"
)

func TestLicenseStatus_IncludesLightMode(t *testing.T) {
	// Mock SDK server that returns license info and light_mode_enabled field
	sdkSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/license/info":
			w.Write([]byte(`{
				"licenseID": "test-123",
				"channelName": "Stable",
				"licenseType": "prod",
				"entitlements": {
					"expires_at": {"title": "Expiration", "value": "2027-01-01T00:00:00Z", "valueType": "String"}
				}
			}`))
		case "/api/v1/license/fields/live_tracking_enabled":
			json.NewEncoder(w).Encode(sdk.LicenseField{Name: "live_tracking_enabled", Value: true, ValueType: "Boolean"})
		case "/api/v1/license/fields/light_mode_enabled":
			json.NewEncoder(w).Encode(sdk.LicenseField{Name: "light_mode_enabled", Value: true, ValueType: "Boolean"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer sdkSrv.Close()

	client := sdk.NewClient(sdkSrv.URL)
	handler := handlers.NewLicenseHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/api/license/status", nil)
	rr := httptest.NewRecorder()
	handler.Status(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		Valid               bool   `json:"valid"`
		Expired             bool   `json:"expired"`
		LicenseType         string `json:"license_type"`
		LiveTrackingEnabled bool   `json:"live_tracking_enabled"`
		LightModeEnabled    bool   `json:"light_mode_enabled"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp.LightModeEnabled {
		t.Error("expected light_mode_enabled to be true")
	}
	if !resp.LiveTrackingEnabled {
		t.Error("expected live_tracking_enabled to be true")
	}
	if !resp.Valid {
		t.Error("expected valid to be true")
	}
}

func TestLicenseStatus_SDKDown_DefaultsFalse(t *testing.T) {
	// SDK server that always 500s
	sdkSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer sdkSrv.Close()

	client := sdk.NewClient(sdkSrv.URL)
	handler := handlers.NewLicenseHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/api/license/status", nil)
	rr := httptest.NewRecorder()
	handler.Status(rr, req)

	var resp struct {
		LightModeEnabled bool `json:"light_mode_enabled"`
		Valid            bool `json:"valid"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.LightModeEnabled {
		t.Error("expected light_mode_enabled false when SDK down")
	}
	if !resp.Valid {
		t.Error("expected valid true when SDK down (fail open)")
	}
}
```

- [ ] **Step 5: Run backend tests**

Run: `cd /Users/jwilson/git/dronerx && go test ./internal/handlers/ -run TestLicenseStatus -v`
Expected: Both `TestLicenseStatus_IncludesLightMode` and `TestLicenseStatus_SDKDown_DefaultsFalse` PASS.

- [ ] **Step 6: Commit backend changes**

```bash
git add internal/config/config.go internal/handlers/license.go internal/handlers/license_test.go cmd/api/main.go
git commit -m "feat: add light_mode_enabled license field to API"
```

---

### Task 2: Helm Chart — Values and ConfigMap

**Files:**
- Modify: `chart/values.yaml:17`
- Modify: `chart/templates/configmap-api.yaml:13`

- [ ] **Step 1: Add `lightModeEnabled` to values.yaml**

In `chart/values.yaml`, add after line 17 (`liveTrackingEnabled: "true"`):

```yaml
  lightModeEnabled: "false"
```

- [ ] **Step 2: Add `LIGHT_MODE_ENABLED` to ConfigMap**

In `chart/templates/configmap-api.yaml`, add after line 13 (`LIVE_TRACKING_ENABLED`):

```yaml
  LIGHT_MODE_ENABLED: {{ .Values.api.lightModeEnabled | quote }}
```

- [ ] **Step 3: Lint the chart**

Run: `cd /Users/jwilson/git/dronerx && helm lint chart/`
Expected: `0 chart(s) failed` — no errors.

- [ ] **Step 4: Commit Helm changes**

```bash
git add chart/values.yaml chart/templates/configmap-api.yaml
git commit -m "feat: add light_mode_enabled to Helm chart config"
```

---

### Task 3: Frontend — Type Update and Theme Store

**Files:**
- Modify: `frontend/src/lib/types.ts:37-43`
- Create: `frontend/src/lib/stores/theme.ts`

- [ ] **Step 1: Add `light_mode_enabled` to LicenseStatus type**

In `frontend/src/lib/types.ts`, update the `LicenseStatus` interface:

```typescript
export interface LicenseStatus {
	valid: boolean;
	expired: boolean;
	license_type?: string;
	expiration_date?: string;
	live_tracking_enabled: boolean;
	light_mode_enabled: boolean;
}
```

- [ ] **Step 2: Create theme store**

Create `frontend/src/lib/stores/theme.ts`:

```typescript
import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'dark' | 'light';

const STORAGE_KEY = 'dronerx-theme';

function getInitialTheme(): Theme {
	if (browser) {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored === 'light' || stored === 'dark') return stored;
	}
	return 'dark';
}

function createThemeStore() {
	const { subscribe, update } = writable<Theme>(getInitialTheme());

	return {
		subscribe,
		toggle() {
			update((current) => {
				const next: Theme = current === 'dark' ? 'light' : 'dark';
				if (browser) {
					localStorage.setItem(STORAGE_KEY, next);
					document.body.dataset.theme = next;
				}
				return next;
			});
		},
		init() {
			// Apply current theme to body on mount
			if (browser) {
				const stored = localStorage.getItem(STORAGE_KEY);
				const theme = stored === 'light' ? 'light' : 'dark';
				document.body.dataset.theme = theme;
			}
		}
	};
}

export const theme = createThemeStore();
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/types.ts frontend/src/lib/stores/theme.ts
git commit -m "feat: add theme store and light_mode_enabled type"
```

---

### Task 4: Frontend — CSS Light Theme Overrides

**Files:**
- Modify: `frontend/src/app.css:36-54` (body and scrollbar), `frontend/src/app.css:106-110` (glass-card)

- [ ] **Step 1: Add `[data-theme="light"]` CSS overrides**

In `frontend/src/app.css`, add the light theme block immediately after the scrollbar styles (after line 54, before the `drone-pulse` keyframes):

```css
/* Light theme overrides */
[data-theme="light"] {
  --color-navy-950: #ffffff;
  --color-navy-900: #f8fafc;
  --color-navy-800: #f1f5f9;
  --color-navy-700: #e2e8f0;
  --color-navy-600: #cbd5e1;
  --color-navy-500: #94a3b8;
  --color-navy-400: #64748b;
  --color-navy-300: #475569;
  --color-navy-200: #334155;
  --color-navy-100: #1e293b;
  --color-navy-50: #0f172a;
}

[data-theme="light"] body,
body[data-theme="light"] {
  color: #1e293b;
}

[data-theme="light"] .glass-card {
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(248, 250, 252, 0.95));
  border: 1px solid rgba(226, 232, 240, 0.8);
}

[data-theme="light"] .glass-card-hover:hover {
  border-color: rgba(0, 229, 255, 0.4);
  box-shadow: 0 4px 24px rgba(0, 229, 255, 0.1), 0 1px 2px rgba(0, 0, 0, 0.05);
}

[data-theme="light"] ::-webkit-scrollbar-track {
  background: #f1f5f9;
}

[data-theme="light"] ::-webkit-scrollbar-thumb {
  background: #cbd5e1;
}

[data-theme="light"] .grid-bg {
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.15) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.15) 1px, transparent 1px);
}
```

- [ ] **Step 2: Verify no build errors**

Run: `cd /Users/jwilson/git/dronerx/frontend && npm run build`
Expected: Build succeeds with no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/app.css
git commit -m "feat: add light theme CSS variable overrides"
```

---

### Task 5: Frontend — ThemeToggle Component

**Files:**
- Create: `frontend/src/lib/components/ThemeToggle.svelte`

- [ ] **Step 1: Create the ThemeToggle component**

Create `frontend/src/lib/components/ThemeToggle.svelte`:

```svelte
<script lang="ts">
	import { theme } from '$lib/stores/theme';
</script>

<button
	onclick={() => theme.toggle()}
	class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-full border transition-all text-sm
		{$theme === 'light'
			? 'bg-amber-glow/10 border-amber-glow/30 text-amber-glow'
			: 'bg-navy-700/50 border-navy-600 text-navy-300 hover:border-navy-400 hover:text-navy-200'}"
	aria-label="Toggle {$theme === 'dark' ? 'light' : 'dark'} mode"
	title="{$theme === 'dark' ? 'Light' : 'Dark'} mode"
>
	{#if $theme === 'dark'}
		<!-- Sun icon -->
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
		</svg>
	{:else}
		<!-- Moon icon -->
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
		</svg>
	{/if}
</button>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/lib/components/ThemeToggle.svelte
git commit -m "feat: add ThemeToggle component"
```

---

### Task 6: Frontend — Layout Integration and Theme Init

**Files:**
- Modify: `frontend/src/routes/+layout.svelte:1-31`

- [ ] **Step 1: Import theme store and initialize on mount**

In `frontend/src/routes/+layout.svelte`, add the theme import and init call. The `license` state is already fetched here — expose `light_mode_enabled` via a Svelte context so child pages can access it without re-fetching.

Replace the `<script>` block with:

```svelte
<script lang="ts">
	import '../app.css';
	import { onMount, setContext } from 'svelte';
	import { getUpdates, getLicenseStatus } from '$lib/api';
	import type { UpdateInfo, LicenseStatus } from '$lib/types';
	import { theme } from '$lib/stores/theme';
	import { writable } from 'svelte/store';

	let { children } = $props();

	let latestUpdate = $state<UpdateInfo | null>(null);
	let license = $state<LicenseStatus | null>(null);
	let bannerDismissed = $state(false);

	let showUpdateBanner = $derived(latestUpdate !== null && !bannerDismissed);
	let showLicenseWarning = $derived(license !== null && (license.expired || !license.valid));

	// Expose light_mode_enabled to child pages via context
	const lightModeEnabled = writable(false);
	setContext('lightModeEnabled', lightModeEnabled);

	onMount(async () => {
		theme.init();

		try {
			const [updates, licenseStatus] = await Promise.all([
				getUpdates().catch(() => []),
				getLicenseStatus().catch(() => null),
			]);
			if (updates && updates.length > 0) {
				latestUpdate = updates[0];
			}
			if (licenseStatus) {
				license = licenseStatus;
				lightModeEnabled.set(licenseStatus.light_mode_enabled ?? false);
			}
		} catch {
			// silent — banners are non-critical
		}
	});
</script>
```

The rest of the template stays unchanged.

- [ ] **Step 2: Commit**

```bash
git add frontend/src/routes/+layout.svelte
git commit -m "feat: init theme store and expose light_mode_enabled via context"
```

---

### Task 7: Frontend — Add ThemeToggle to All Page Headers

**Files:**
- Modify: `frontend/src/routes/+page.svelte:1-53` (home header)
- Modify: `frontend/src/routes/order/+page.svelte:1-58` (place order header)
- Modify: `frontend/src/routes/order/[id]/+page.svelte:1-239` (order tracking header)
- Modify: `frontend/src/routes/orders/+page.svelte:1-71` (order history header)
- Modify: `frontend/src/routes/admin/+page.svelte:1-45` (admin header)

- [ ] **Step 1: Add toggle to home page header (`+page.svelte`)**

In `frontend/src/routes/+page.svelte`, add the imports at the top of the `<script>` block (after existing imports):

```typescript
import ThemeToggle from '$lib/components/ThemeToggle.svelte';
import { getContext } from 'svelte';
import type { Writable } from 'svelte/store';

const lightModeEnabled = getContext<Writable<boolean>>('lightModeEnabled');
```

In the header `<nav>` element (line 30-51), add the toggle after the Cart link's closing `</a>` and before the closing `</nav>`:

```svelte
			{#if $lightModeEnabled}
				<span class="text-navy-600">|</span>
				<ThemeToggle />
			{/if}
```

- [ ] **Step 2: Add toggle to place order page header**

In `frontend/src/routes/order/+page.svelte`, add imports at the top of the `<script>` block:

```typescript
import ThemeToggle from '$lib/components/ThemeToggle.svelte';
import { getContext } from 'svelte';
import type { Writable } from 'svelte/store';

const lightModeEnabled = getContext<Writable<boolean>>('lightModeEnabled');
```

In the header `<div>` (line 45-57), change the closing `</div>` of the header inner container to include the toggle. After line 56 (`<span class="text-navy-200 font-medium">Place Order</span>`), add:

```svelte
		{#if $lightModeEnabled}
			<span class="ml-auto text-navy-600">|</span>
			<ThemeToggle />
		{/if}
```

- [ ] **Step 3: Add toggle to order tracking page header**

In `frontend/src/routes/order/[id]/+page.svelte`, add imports at the top of the `<script>` block (after existing imports):

```typescript
import ThemeToggle from '$lib/components/ThemeToggle.svelte';
import { getContext } from 'svelte';
import type { Writable } from 'svelte/store';

const lightModeEnabled = getContext<Writable<boolean>>('lightModeEnabled');
```

In the header (lines 207-239), add the toggle before the closing `</div>` of the header inner container (before line 238 `</div>`), but after the tracking indicator `{/if}` on line 237:

```svelte
		{#if $lightModeEnabled}
			<span class="text-navy-600 {wsConnected || trackingEnabled === false ? '' : 'ml-auto'}">|</span>
			<ThemeToggle />
		{/if}
```

- [ ] **Step 4: Add toggle to order history page header**

In `frontend/src/routes/orders/+page.svelte`, add imports at the top of the `<script>` block:

```typescript
import ThemeToggle from '$lib/components/ThemeToggle.svelte';
import { getContext } from 'svelte';
import type { Writable } from 'svelte/store';

const lightModeEnabled = getContext<Writable<boolean>>('lightModeEnabled');
```

In the header (lines 57-71), add after line 69 (`<span class="text-navy-200 font-medium">Order History</span>`):

```svelte
		{#if $lightModeEnabled}
			<span class="ml-auto text-navy-600">|</span>
			<ThemeToggle />
		{/if}
```

- [ ] **Step 5: Add toggle to admin page header**

In `frontend/src/routes/admin/+page.svelte`, add imports at the top of the `<script>` block:

```typescript
import ThemeToggle from '$lib/components/ThemeToggle.svelte';
import { getContext } from 'svelte';
import type { Writable } from 'svelte/store';

const lightModeEnabled = getContext<Writable<boolean>>('lightModeEnabled');
```

In the header `<nav>` (lines 39-43), add the toggle after the "Back to Store" link and before the closing `</nav>`:

```svelte
			{#if $lightModeEnabled}
				<span class="text-navy-600">|</span>
				<ThemeToggle />
			{/if}
```

- [ ] **Step 6: Verify frontend builds**

Run: `cd /Users/jwilson/git/dronerx/frontend && npm run build`
Expected: Build succeeds with no errors.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/routes/+page.svelte frontend/src/routes/order/+page.svelte frontend/src/routes/order/\[id\]/+page.svelte frontend/src/routes/orders/+page.svelte frontend/src/routes/admin/+page.svelte
git commit -m "feat: add ThemeToggle to all page headers"
```

---

### Task 8: End-to-End Verification

- [ ] **Step 1: Run all backend tests**

Run: `cd /Users/jwilson/git/dronerx && go test ./... -v`
Expected: All tests pass, including the new license handler tests.

- [ ] **Step 2: Run frontend build**

Run: `cd /Users/jwilson/git/dronerx/frontend && npm run build`
Expected: Build succeeds with no errors or warnings.

- [ ] **Step 3: Start the dev server and visually test**

Run: `cd /Users/jwilson/git/dronerx/frontend && npm run dev`

Test checklist:
- Home page loads in dark mode by default
- Toggle is NOT visible (no license field set)
- Manually set `LIGHT_MODE_ENABLED=true` env var or mock the API to return `light_mode_enabled: true`
- Toggle appears in header after Cart button
- Clicking toggle switches to light mode (white backgrounds, dark text)
- Clicking again switches back to dark mode
- Refresh the page — theme preference persists from localStorage
- Navigate to other pages — toggle appears in all headers
- Cyan and amber accents remain visible and readable in light mode

- [ ] **Step 4: Lint the Helm chart**

Run: `cd /Users/jwilson/git/dronerx && helm lint chart/`
Expected: Lint passes with no errors.
