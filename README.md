# DroneRx

Medicine delivery by drone. Patients browse medicines, place orders, and track deliveries in real-time as drones fly from the local chemist to their door.

Built as a [Replicated Bootcamp](https://docs.replicated.com) exercise covering Tiers 0–7.

## Architecture

| Component | Technology |
|-----------|-----------|
| Backend | Go (REST API + WebSocket) |
| Frontend | SvelteKit + TypeScript + Tailwind CSS |
| Database | PostgreSQL via [CloudNativePG](https://cloudnative-pg.io/) |
| Messaging | [NATS](https://nats.io/) (real-time order status events) |
| Distribution | Replicated (Helm + Embedded Cluster) |

```
┌──────────────┐     ┌──────────────┐
│   SvelteKit  │────▶│   Go API     │
│   Frontend   │     │  :8080       │
│   :3000      │     └──────┬───────┘
└──────────────┘            │
                    ┌───────┴───────┐
                    │               │
              ┌─────▼─────┐  ┌─────▼─────┐
              │ PostgreSQL │  │   NATS    │
              │ (CNPG)    │  │           │
              └───────────┘  └───────────┘
```

## Features

| Feature | Description |
|---------|------------|
| Order placement | Browse medicines, add to cart, submit order |
| Delivery ETA | Estimated delivery time based on order status |
| Live tracking | Real-time WebSocket updates via NATS (license-gated) |
| Webhook notifications | HTTP POST on delivery completion |

Orders auto-advance through 4 statuses: `placed → preparing → in-flight → delivered`

## Local Development

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose

### Option A: Docker Compose (recommended)

```bash
docker compose up --build
```

Open http://localhost:3000

### Option B: Run services individually

```bash
# Start Postgres and NATS
docker compose up -d postgres nats

# Run the Go API
DATABASE_URL="postgres://dronerx:dronerx@localhost:5432/dronerx?sslmode=disable" \
NATS_URL="nats://localhost:4222" \
go run ./cmd/api/

# Run the frontend (separate terminal)
cd frontend && npm install && npm run dev
```

Open http://localhost:5173

## Helm Chart

### Install on an existing cluster

```bash
# Install CloudNativePG operator first
helm repo add cnpg https://cloudnative-pg.github.io/charts
helm install cnpg-operator cnpg/cloudnative-pg -n cnpg-system --create-namespace --wait

# Install DroneRx
helm dependency build chart/
helm install dronerx chart/ \
  --namespace dronerx \
  --create-namespace \
  --set cloudnativepg.enabled=false \
  --set api.image.tag=0.1.0 \
  --set frontend.image.tag=0.1.0
```

### Install with bundled operator

```bash
helm dependency build chart/
helm install dronerx chart/ \
  --namespace dronerx \
  --create-namespace \
  --set api.image.tag=0.1.0 \
  --set frontend.image.tag=0.1.0 \
  --timeout 5m
```

### Configuration

| Value | Default | Description |
|-------|---------|------------|
| `api.tickerInterval` | `10` | Seconds between order status transitions |
| `api.webhookURL` | `""` | Webhook URL for delivery notifications |
| `service.type` | `ClusterIP` | Service type (ClusterIP/NodePort/LoadBalancer) |
| `ingress.enabled` | `false` | Enable ingress |
| `ingress.tls.mode` | `self-signed` | TLS mode: `auto`, `manual`, or `self-signed` |
| `postgresql.enabled` | `true` | Deploy embedded PostgreSQL (CNPG) |
| `externalDatabase.host` | `""` | External DB host (when postgresql.enabled=false) |

## Makefile

```bash
make build           # Build Docker images
make lint            # Run Go vet, svelte-check, helm lint
make test            # Run Go and frontend tests
make clean           # Clean build artifacts
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|------------|
| `PORT` | `8080` | API server port |
| `DATABASE_URL` | — | PostgreSQL connection string (required) |
| `NATS_URL` | `nats://localhost:4222` | NATS server URL |
| `TICKER_INTERVAL` | `10` | Seconds between status transitions |
| `WEBHOOK_URL` | `""` | Delivery notification webhook URL |

## Bootcamp Progress

- [x] **Tier 0** — Build It
- [ ] **Tier 1** — Automate It (CI/CD)
- [ ] **Tier 2** — Ship It with Helm (SDK, metrics, license gating)
- [ ] **Tier 3** — Support It (preflight, support bundle)
- [ ] **Tier 4** — Ship It on a VM (Embedded Cluster)
- [ ] **Tier 5** — Config Screen
- [ ] **Tier 6** — Deliver It (Enterprise Portal)
- [ ] **Tier 7** — Operationalize It
