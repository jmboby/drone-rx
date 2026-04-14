# Tier 4 — Ship It on a VM (Embedded Cluster v3)

**Date:** 2026-04-14
**Status:** Approved
**Rubric items:** 4.1, 4.2, 4.3, 4.6, 4.7 + cert-manager/Cloudflare TLS follow-on

---

## Context

Tier 3 (Support It) is complete. DroneRx already has:
- KOTS Application manifest (`replicated/kots-app.yaml`) with icon and status informers
- HelmChart CR (`replicated/dronerx-chart.yaml`) with image overrides
- Embedded/external DB toggle (`postgresql.enabled`)
- License entitlement `live_tracking_enabled` gated via SDK in Go code
- v1beta2 preflight spec inside `chart/templates/_preflight.tpl` (Helm-native)
- CI pipeline that builds releases and tests on CMX k3s

## Goals

1. Enable Embedded Cluster v3 distribution of DroneRx on bare VMs
2. Support online install, in-place upgrade, and air-gap install
3. Gate `live_tracking_enabled` via KOTS `LicenseFieldValue` on the EC path (belt-and-suspenders with existing SDK runtime check)
4. Add static v1beta3 preflight spec for EC installer
5. Follow-on: Add cert-manager + Cloudflare as EC Helm extension for real TLS

## Non-goals

- KOTS Config screen (Tier 5)
- Conditional v1beta3 preflights using KOTS template functions (requires Config screen)
- Multi-node EC clusters

---

## Phase 1: EC Core

### New file: `replicated/embedded-cluster.yaml`

Minimal EmbeddedClusterConfig CRD. EC's built-in OpenEBS handles storage.

```yaml
apiVersion: embeddedcluster.replicated.com/v1beta1
kind: Config
metadata:
  name: drone-rx
spec:
  version: "3.0.0-alpha-31+k8s-1.34"
```

No Helm extensions in Phase 1. cert-manager added in Phase 2.

### New file: `replicated/preflight-v1beta3.yaml`

Static v1beta3 preflight at the release top level. Only includes always-on checks that don't need Helm templating (which is broken in EC's built-in preflight runner):

- Cluster CPU capacity (min 2 cores, warn < 4)
- Cluster memory capacity (min 4 GiB, warn < 8)
- Kubernetes version (min 1.28)
- Distribution check (fail docker-desktop, microk8s)
- Default storage class present

The existing v1beta2 preflight in `chart/templates/_preflight.tpl` is untouched and continues to work for Helm CLI installs (including the conditional external DB and Cloudflare checks).

Conditional checks (external DB, Cloudflare API) are deferred to Tier 5 when the KOTS Config screen provides `repl{{ ConfigOptionEquals }}` template functions.

### Modified: `replicated/dronerx-chart.yaml`

Add `LicenseFieldValue` for entitlement gating on the EC path:

```yaml
spec:
  values:
    api:
      liveTrackingEnabled: repl{{ LicenseFieldValue "live_tracking_enabled" }}
```

The Go API code already reads this entitlement via the SDK. The `LicenseFieldValue` seeds the same value at install time so the app gets the correct initial state on EC installs.

Add EC-sensible defaults where the HelmChart CR needs to override values.yaml defaults:
- `ingress.enabled: false` (EC admin console proxies traffic)
- `postgresql.enabled: true` (embedded DB by default on VMs)

### Modified: `replicated/kots-app.yaml`

Verify and adjust as needed:
- `spec.icon` — SVG URL from GitHub raw (already set)
- `spec.title` — "DroneRx" (already set)
- `spec.ports` — Ensure `localPort` and `applicationUrl` work with EC admin console proxy
- `spec.statusInformers` — Already tracks api and frontend deployments

### Modified: `.github/workflows/release.yaml`

Add EC-specific steps to the release pipeline:
- The release already gets created via `replicated release create` — EC builds are triggered when the channel has EC enabled
- Air-gap bundles are auto-built by Replicated on channel promotion
- Optionally add CMX VM smoke test for EC path using `replicated cluster create --distribution embedded-cluster-v3`

---

## Phase 2: cert-manager + Cloudflare TLS

After Tier 4 rubric items are validated, add real TLS support on the EC path.

### Modified: `replicated/embedded-cluster.yaml`

Add cert-manager as a Helm extension:

```yaml
spec:
  version: "3.0.0-alpha-31+k8s-1.34"
  extensions:
    helmCharts:
      - chart:
          name: cert-manager
          chartVersion: "v1.17.1"
        releaseName: cert-manager
        namespace: cert-manager
        weight: 10
        values: |
          installCRDs: true
```

Weight 10 ensures cert-manager installs before the app chart.

### Modified: `replicated/dronerx-chart.yaml`

Wire Cloudflare TLS toggle for EC path. The chart already has `ingress.tls.cloudflare.enabled` and cert-manager integration — this just makes the toggle available through the HelmChart CR values.

---

## Testing Plan

| Rubric | Test method | Pass criteria |
|--------|------------|---------------|
| 4.1 | CMX VM: `replicated cluster create --distribution embedded-cluster-v3`, install with license | All pods Running, app opens in browser via admin console proxy |
| 4.2 | Install release N, create drone delivery data, upgrade to release N+1 | Data persists, all pods Running after upgrade |
| 4.3 | Build air-gap bundle, transfer to isolated CMX VM, install with `--airgap` flag | All pods Running, app accessible, no outbound network calls |
| 4.6 | Visual check during EC install | DroneRx icon and title shown in admin console |
| 4.7 | Install with license where `live_tracking_enabled: false` | Feature returns 403; update license to `true`, feature works without redeploy |
| Phase 2 | EC install with cert-manager extension + Cloudflare enabled | Real TLS cert issued, app accessible over HTTPS |

## Known Constraints

- **EC v3 is alpha** (`3.0.0-alpha-31`) — version may need bumping during implementation
- **v1beta3 preflight + Helm templating is broken in EC** — static checks only until fixed
- **Air-gap images can't use multi-arch digests** — set digests to empty strings in extension values if needed
- **`spec.api` and `spec.storage` are immutable post-install** — get them right on first release

## Architecture Decisions

| Decision | Rationale |
|----------|-----------|
| Minimal EC config (no extensions in Phase 1) | Reduce variables for first EC install; OpenEBS built-in is sufficient |
| Belt-and-suspenders entitlement (SDK + LicenseFieldValue) | SDK does runtime checks; LicenseFieldValue seeds correct initial state on EC |
| Static v1beta3 preflight | EC's Helm templating is broken; conditional checks deferred to Tier 5 Config screen |
| cert-manager as Phase 2 | Rubric doesn't require real TLS on EC; validate core flow first |
| Existing v1beta2 preflight untouched | Still works for Helm CLI path; no reason to disrupt |
