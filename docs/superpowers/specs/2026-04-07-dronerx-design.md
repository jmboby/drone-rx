# DroneRx — Medicine Delivery by Drone

**Date:** 2026-04-07
**Status:** Approved
**Scope:** Replicated distribution exercise — satisfies Tiers 0–7 of the Replicated Rubric

---

## Overview

DroneRx is a lightweight medicine-by-drone delivery application. Patients browse medicines from their local chemist, place orders, receive delivery time estimates, and (with a premium license) track deliveries in real-time via WebSocket. Pharmacy fulfilment is simulated — a backend state machine auto-advances orders through statuses on a configurable timer.

The application is designed to satisfy all tiers of the Replicated Rubric (Tiers 0–7).

---

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go (REST API + WebSocket) |
| Frontend | SvelteKit + TypeScript + Tailwind CSS |
| Database | PostgreSQL via CloudNativePG (official chart) |
| Messaging | NATS (nats-io/nats official chart) |
| Container registry | GHCR (private) |
| CI/CD | GitHub Actions + Replicated GitHub Actions |
| Replicated SDK | Subchart, renamed to `dronerx-sdk` |

**Constraints:**
- No Bitnami charts
- No authentication (not required by rubric)
- cert-manager is an EC extension / cluster prerequisite, not a subchart

---

## Repo Structure

```
dronerx/
├── cmd/api/              # Go backend entry point
├── internal/             # Go packages (handlers, models, state machine, nats)
├── frontend/             # SvelteKit + TypeScript app
├── chart/                # Helm chart
│   ├── Chart.yaml        # Subcharts: cloudnative-pg, nats
│   ├── values.yaml
│   ├── values.schema.json
│   └── templates/
│       ├── _preflight.tpl
│       ├── _supportbundle.tpl
│       └── ...
├── replicated/           # EC v3 config, KOTS manifests, branding
├── release/              # Replicated release manifests (no scripts)
├── Dockerfile.api        # Go multi-stage build
├── Dockerfile.frontend   # SvelteKit build + Node runtime
├── Makefile              # All automation targets
└── .github/workflows/    # CI/CD (calls make targets)
```

---

## Features

| Feature | Type | License-gated? |
|---------|------|----------------|
| Order placement | Core | No |
| Delivery time estimation | Core | No |
| Live delivery tracking (WebSocket) | Premium | Yes — `live_tracking_enabled` license field |
| Delivery notifications (webhook) | Configurable | No — webhook URL set via Config Screen; fires on `delivered` status |

---

## Data Model

### Postgres Tables

| Table | Key Columns |
|-------|------------|
| `medicines` | id (uuid), name, description, price, in_stock (bool), category |
| `orders` | id (uuid), patient_name, address, status (enum), estimated_delivery (timestamp), created_at, updated_at |
| `order_items` | id (uuid), order_id (fk), medicine_id (fk), quantity |

### Order State Machine

**Statuses:** `placed → preparing → in-flight → delivered`

- A ticker goroutine auto-advances each order through statuses on a configurable interval (e.g., 30s per transition)
- Each transition: update Postgres, then publish event to NATS
- Orders only advance forward; `delivered` is terminal
- ~10 medicines pre-loaded via DB migration/init job for immediate demoability

### NATS Events

- Subject pattern: `orders.<order_id>.status`
- Payload: `{"order_id": "...", "status": "in-flight", "estimated_delivery": "...", "updated_at": "..."}`
- Go API subscribes on behalf of connected WebSocket clients
- License check at WebSocket connection time — no valid license = connection refused, frontend falls back to polling

---

## API Design

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/api/medicines` | List available medicines |
| `GET` | `/api/medicines/:id` | Get medicine details |
| `POST` | `/api/orders` | Place an order |
| `GET` | `/api/orders/:id` | Get order status + ETA |
| `GET` | `/api/orders` | List patient's orders |
| `WS` | `/api/orders/:id/track` | Live tracking stream (license-gated) |
| `GET` | `/healthz` | Structured health check (DB + NATS connectivity) |

**Health endpoint:** Returns `{"status": "ok", "db": "ok", "nats": "ok"}` — serves Tier 0.4 and Tier 3.3.

**ETA calculation:** Mock — base time per status minus elapsed time. Returned in the order response.

**WebSocket license gate:** Go API checks license entitlement via SDK before upgrading connection. No valid license = HTTP 403. Frontend falls back to polling `GET /api/orders/:id`.

---

## Frontend

| Route | Purpose |
|-------|---------|
| `/` | Landing page — browse medicines by category |
| `/order` | Order form — selected items, patient name, address, submit |
| `/order/[id]` | Order status — current state, ETA, live tracking if licensed |
| `/orders` | Order history — list past orders by patient name |

**Style:** Modern, sleek — clean typography, generous whitespace, subtle animations on status progression. Medical/pharmacy colour palette (calming blues/greens). Built with Tailwind CSS. Frontend design skill will be used during implementation for polish.

**No auth UI** — order history looked up by patient name input.

---

## Helm Chart & Kubernetes

### Deployments

| Component | Template | Notes |
|-----------|----------|-------|
| Go API | Deployment + Service | Liveness + readiness on `/healthz`, resource requests/limits, init container waits for DB |
| SvelteKit frontend | Deployment + Service | Liveness + readiness, resource requests/limits |
| CloudNativePG | Subchart | Embedded by default, BYO opt-in via `postgresql.enabled: false` + `externalDatabase.*` |
| NATS | Subchart | nats-io/nats official chart, embedded by default |
| Ingress | Optional, off by default | 3 TLS modes: auto (cert-manager), manual (user Secret), self-signed (generated) |

### Service Type

Configurable via values: ClusterIP (default) / NodePort / LoadBalancer.

### Preflight & Support Bundle

Both embedded as Helm template helpers in `chart/templates/`:

**Preflight checks (all 5 mandatory):**
1. External DB connectivity (conditional — only when BYO configured)
2. Notification webhook URL reachability (when configured)
3. Cluster resource check (CPU + memory)
4. Kubernetes version check
5. Distribution check (fail on docker-desktop, microk8s)

**Support bundle:**
- Log collectors per component (app, CloudNativePG, NATS) with `maxLines`/`maxAge`
- HTTP collector on `/healthz` + textAnalyze for pass/fail
- Status analyzers: deploymentStatus, statefulsetStatus
- textAnalyze for known app failure pattern (regex on logs)
- storageClass + nodeResources analyzers
- Upload to Vendor Portal via SDK from app UI

---

## Replicated Integration

### SDK (Tier 2)

- Deployed as subchart, renamed to `dronerx-sdk`
- All images proxied through custom domain
- Custom metrics: orders placed, orders delivered, avg delivery time
- License entitlement: `live_tracking_enabled` boolean field
- Update available banner with version info
- License validity enforcement (expiry/invalid → warning/block)
- Instance reporting — all services healthy

### CI/CD (Tier 1)

| Workflow | Trigger | Steps |
|----------|---------|-------|
| PR | Pull request | `make build`, `make lint`, create Replicated release from `.replicated`, test via CMX |
| Release | Merge to main | `make build`, push to GHCR, create release, promote to Unstable |
| Promote | Manual | Promote to Beta or Stable |

- Scoped Replicated RBAC policy for CI service account
- Email notification on Stable promotion only
- Follows docs at docs.replicated.com/vendor/ci-workflows-github-actions

### Makefile Targets

| Target | Purpose |
|--------|---------|
| `make build` | Build both Docker images |
| `make lint` | helm lint + Go lint + frontend lint |
| `make release` | Create a Replicated release |
| `make promote` | Promote release to a channel |
| `make test` | Run tests |

---

## Embedded Cluster & Config Screen (Tiers 4–5)

### EC v3 Install

- EC config in `replicated/` defines k0s cluster
- cert-manager as EC extension
- App icon + name set in Application CR
- License entitlement gates live tracking via KOTS `LicenseFieldValue` in HelmChart CR

### Config Screen

| Config Item | Type | Conditional | Notes |
|-------------|------|-------------|-------|
| External DB toggle | `select_one` | Reveals host, port, credentials when external | Tier 5.0 |
| Embedded DB password | `password` | Visible when embedded selected | Auto-generated, survives upgrade (Tier 5.2) |
| App domain | `text` | Always visible | Hostname for Ingress + cert CN |
| TLS mode | `select_one` | auto / manual / self-signed | Tier 0.5 |
| TLS email | `text` | Visible when TLS = auto | Let's Encrypt email, regex validated |
| Live tracking toggle | `bool` | Gated by license field | Tier 4.7 / 5.1 |
| Notification webhook URL | `text` | Always visible | Regex validated URL (Tier 5.3) |
| Service type | `select_one` | Always visible | ClusterIP / NodePort / LoadBalancer |

All items have `help_text` (Tier 5.4).

---

## Enterprise Portal v2 (Tier 6)

- Custom branding: logo, favicon, title, primary/secondary colours
- Custom email sender from vendor domain
- Security center with CVE visibility
- GitHub app integration for custom setup instructions
- Auto-generated Helm chart reference in `toc.yaml` (1 field intentionally undocumented)
- Terraform modules gated by license field
- Self-serve sign-up URL
- End-to-end install instructions for both Helm + EC paths
- Upgrade instructions for both paths

---

## Operationalize (Tier 7)

- Email + webhook notifications on account activity events
- CVE security posture assessment + reduction strategy
- Container image signing with Cosign
- Zero outbound verification under CMX network policy in air-gap
