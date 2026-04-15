# Light Mode — License-Gated Theme Toggle

## Overview

A custom Replicated license field (`light_mode_enabled`, Boolean) unlocks a sun/moon theme toggle in the app header. When toggled, the app swaps from the dark navy theme to a white/light gray theme. The user's preference persists in `localStorage`. Without the license entitlement, the toggle is hidden and the app stays dark.

## Approach

**License-unlocked toggle** — the license field gates visibility of a UI toggle; the user chooses dark or light. This is more interactive than a license-only approach and better demonstrates the Replicated entitlement concept.

**Color strategy: swap backgrounds and text, keep accents.** Navy variables are overridden via a `[data-theme="light"]` CSS selector. Cyan-glow and amber-glow accents remain unchanged — they work well on both dark and light backgrounds.

## License Integration (Backend)

### Custom License Field

- **Field name:** `light_mode_enabled`
- **Type:** Boolean
- **Default:** `false`
- Created in the Replicated vendor portal as a custom license field

### API Changes

**`internal/handlers/license.go`** — Add `LightModeEnabled` to the `licenseStatusResponse` struct:

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

Populate it using `h.client.IsFeatureEnabled("light_mode_enabled")` — same pattern as `live_tracking_enabled`.

### Configuration

**`internal/config/config.go`** — Add `LightModeEnabled` env var with default `"false"` for fallback when SDK is unavailable.

**`chart/templates/configmap-api.yaml`** — Add:

```yaml
LIGHT_MODE_ENABLED: {{ .Values.api.lightModeEnabled | quote }}
```

**`chart/values.yaml`** — Add under `api`:

```yaml
lightModeEnabled: "false"
```

## Frontend — Theme Store

**New file: `frontend/src/lib/stores/theme.ts`**

- Svelte `writable` store holding `"dark"` or `"light"`
- On init, reads `localStorage` key `dronerx-theme`; defaults to `"dark"`
- On change, writes to `localStorage` and sets `document.body.dataset.theme`

## Frontend — Toggle Component

**New file: `frontend/src/lib/components/ThemeToggle.svelte`**

- Sun/moon icon pill button
- **Placement:** Header nav, far right after Cart button, separated by a subtle divider
- **Visibility:** Only rendered when `license.light_mode_enabled === true`
- Clicking toggles the theme store between `"dark"` and `"light"`

## Frontend — CSS Theming

**`frontend/src/app.css`** — Add a `[data-theme="light"]` block that overrides the navy CSS custom properties:

| Dark (current) | Variable | Light (override) |
|---|---|---|
| `#1a2035` | `--color-navy-950` | `#ffffff` |
| `#212a42` | `--color-navy-900` | `#f8fafc` |
| `#2a3450` | `--color-navy-800` | `#f1f5f9` |
| `#1c274a` | `--color-navy-700` | `#e2e8f0` |
| `#263362` | `--color-navy-600` | `#cbd5e1` |
| `#334580` | `--color-navy-500` | `#94a3b8` |
| `#4a5f9e` | `--color-navy-400` | `#64748b` |
| `#6b7fbf` | `--color-navy-300` | `#475569` |
| `#95a5d6` | `--color-navy-200` | `#334155` |
| `#c2cceb` | `--color-navy-100` | `#1e293b` |
| `#e8ecf7` | `--color-navy-50` | `#0f172a` |

Additional overrides:
- Body text color: `#1e293b`
- `.glass-card`: white/80 background with subtle gray border instead of dark gradient
- Scrollbar: light gray track and thumb

Cyan-glow (`#00e5ff`) and amber-glow (`#ffab00`) stay unchanged.

## Layout Integration

**`frontend/src/routes/+layout.svelte`** — Import theme store; apply `data-theme` attribute to body on mount and on change. The layout already fetches license status; expose `light_mode_enabled` so pages can pass it to the toggle.

**All page headers** — Each page defines its own `<header>`. The toggle must be added to all 5:
- `frontend/src/routes/+page.svelte` (home)
- `frontend/src/routes/order/+page.svelte` (place order)
- `frontend/src/routes/order/[id]/+page.svelte` (order tracking)
- `frontend/src/routes/orders/+page.svelte` (order history)
- `frontend/src/routes/admin/+page.svelte` (admin)

Each header gets the `ThemeToggle` component after the last nav element, conditionally rendered based on the license field.

**`frontend/src/lib/types.ts`** — Add `light_mode_enabled: boolean` to the `LicenseStatus` interface.

## Files Summary

| File | Change |
|---|---|
| `internal/handlers/license.go` | Add `LightModeEnabled` to response |
| `internal/config/config.go` | Add `LightModeEnabled` env var |
| `chart/templates/configmap-api.yaml` | Add `LIGHT_MODE_ENABLED` |
| `chart/values.yaml` | Add `lightModeEnabled: "false"` |
| `frontend/src/lib/types.ts` | Add `light_mode_enabled` to `LicenseStatus` |
| `frontend/src/lib/stores/theme.ts` | **New** — theme writable store with localStorage |
| `frontend/src/lib/components/ThemeToggle.svelte` | **New** — sun/moon toggle component |
| `frontend/src/app.css` | Add `[data-theme="light"]` variable overrides |
| `frontend/src/routes/+layout.svelte` | Apply data-theme, expose license to pages |
| `frontend/src/routes/+page.svelte` | Add ThemeToggle to header |
| `frontend/src/routes/order/+page.svelte` | Add ThemeToggle to header |
| `frontend/src/routes/order/[id]/+page.svelte` | Add ThemeToggle to header |
| `frontend/src/routes/orders/+page.svelte` | Add ThemeToggle to header |
| `frontend/src/routes/admin/+page.svelte` | Add ThemeToggle to header |

## What Doesn't Change

- All page components keep existing Tailwind classes (`bg-navy-950`, `text-navy-200`, etc.) — CSS variable overrides handle the swap
- Cyan/amber accent colors remain the same
- Animations remain the same
- Other pages (order, orders, admin) inherit the theme automatically via CSS variables
