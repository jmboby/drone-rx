# DroneRx Phase 1: Build It (Tier 0) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a working medicine-by-drone delivery app with Go backend, SvelteKit frontend, CloudNativePG, NATS, Helm chart with 2 subcharts, TLS (3 modes), health endpoint, init container, and seed data — satisfying all Tier 0 rubric requirements.

**Architecture:** Monorepo with a Go REST API + WebSocket backend, SvelteKit + TypeScript + Tailwind CSS frontend, CloudNativePG for persistence, NATS for real-time event streaming. Single Helm chart with CloudNativePG and NATS as subcharts. Order state machine auto-advances on a timer, publishing events to NATS for live tracking via WebSocket.

**Tech Stack:** Go 1.22+, SvelteKit 2, TypeScript, Tailwind CSS 4, PostgreSQL (CloudNativePG), NATS, Helm 3, Docker

**Spec:** `docs/superpowers/specs/2026-04-07-dronerx-design.md`

---

## File Structure

### Go Backend

| File | Responsibility |
|------|---------------|
| `cmd/api/main.go` | Entry point — wires dependencies, starts HTTP server |
| `internal/config/config.go` | Environment variable parsing into typed config struct |
| `internal/database/database.go` | Postgres connection pool via pgx |
| `internal/database/migrate.go` | Run SQL migrations on startup |
| `internal/database/migrations/001_create_medicines.up.sql` | Medicines table + seed data |
| `internal/database/migrations/001_create_medicines.down.sql` | Drop medicines table |
| `internal/database/migrations/002_create_orders.up.sql` | Orders + order_items tables |
| `internal/database/migrations/002_create_orders.down.sql` | Drop orders tables |
| `internal/models/medicine.go` | Medicine struct + DB queries |
| `internal/models/order.go` | Order/OrderItem structs, status enum, DB queries |
| `internal/models/eta.go` | ETA calculation logic |
| `internal/handlers/medicine.go` | HTTP handlers for /api/medicines |
| `internal/handlers/order.go` | HTTP handlers for /api/orders |
| `internal/handlers/health.go` | GET /healthz handler |
| `internal/handlers/tracking.go` | WebSocket handler for live tracking |
| `internal/events/publisher.go` | NATS event publishing |
| `internal/events/subscriber.go` | NATS subscription for WebSocket relay |
| `internal/statemachine/ticker.go` | Goroutine that advances order statuses |
| `internal/webhook/notifier.go` | HTTP POST to webhook URL on delivery |

### Go Tests

| File | Tests |
|------|-------|
| `internal/config/config_test.go` | Config parsing from env vars |
| `internal/models/eta_test.go` | ETA calculation |
| `internal/models/medicine_test.go` | Medicine DB queries (integration) |
| `internal/models/order_test.go` | Order DB queries + status transitions (integration) |
| `internal/handlers/medicine_test.go` | Medicine HTTP handler responses |
| `internal/handlers/order_test.go` | Order HTTP handler responses |
| `internal/handlers/health_test.go` | Health endpoint responses |
| `internal/statemachine/ticker_test.go` | State machine advances orders correctly |
| `internal/webhook/notifier_test.go` | Webhook fires on delivery |

### Frontend

| File | Responsibility |
|------|---------------|
| `frontend/src/routes/+page.svelte` | Landing page — medicine browsing |
| `frontend/src/routes/+page.ts` | Load medicines from API |
| `frontend/src/routes/order/+page.svelte` | Order form |
| `frontend/src/routes/order/+page.ts` | Load cart state |
| `frontend/src/routes/order/[id]/+page.svelte` | Order status + live tracking |
| `frontend/src/routes/order/[id]/+page.ts` | Load order by ID |
| `frontend/src/routes/orders/+page.svelte` | Order history |
| `frontend/src/routes/orders/+page.ts` | Load orders by patient name |
| `frontend/src/lib/api.ts` | API client (fetch wrapper) |
| `frontend/src/lib/types.ts` | Shared TypeScript types |
| `frontend/src/lib/stores/cart.ts` | Cart store (Svelte store) |
| `frontend/src/lib/components/MedicineCard.svelte` | Medicine display card |
| `frontend/src/lib/components/StatusTracker.svelte` | Visual 4-step status progression |
| `frontend/src/app.css` | Tailwind base styles |

### Helm Chart

| File | Responsibility |
|------|---------------|
| `chart/Chart.yaml` | Chart metadata + subchart dependencies |
| `chart/values.yaml` | Default values |
| `chart/values.schema.json` | JSON schema for values validation |
| `chart/templates/_helpers.tpl` | Template helpers (names, labels, selectors) |
| `chart/templates/api-deployment.yaml` | Go API deployment with init container |
| `chart/templates/api-service.yaml` | Go API service |
| `chart/templates/frontend-deployment.yaml` | SvelteKit deployment |
| `chart/templates/frontend-service.yaml` | SvelteKit service |
| `chart/templates/ingress.yaml` | Optional ingress with TLS modes |
| `chart/templates/postgres-cluster.yaml` | CloudNativePG Cluster CR |
| `chart/templates/self-signed-cert-job.yaml` | Job to generate self-signed cert |
| `chart/templates/configmap-api.yaml` | API environment config |
| `chart/templates/secret-db.yaml` | DB credentials secret |
| `chart/templates/_preflight.tpl` | Preflight spec (placeholder for Tier 3) |
| `chart/templates/_supportbundle.tpl` | Support bundle spec (placeholder for Tier 3) |

### Root Files

| File | Responsibility |
|------|---------------|
| `Dockerfile.api` | Multi-stage Go build |
| `Dockerfile.frontend` | SvelteKit build + Node runtime |
| `Makefile` | Build, lint, test targets |
| `go.mod` | Go module definition |
| `.gitignore` | Ignore patterns |

---

## Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`, `go.sum`, `.gitignore`, `cmd/api/main.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/jwilson/git/dronerx
go mod init github.com/jwilson/dronerx
```

- [ ] **Step 2: Create .gitignore**

Create `.gitignore`:

```gitignore
# Go
/bin/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Frontend
frontend/node_modules/
frontend/.svelte-kit/
frontend/build/

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local
```

- [ ] **Step 3: Create minimal main.go**

Create `cmd/api/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 4: Verify it compiles and runs**

```bash
cd /Users/jwilson/git/dronerx
go run ./cmd/api/
# In another terminal: curl http://localhost:8080/healthz
# Expected: {"status":"ok"}
```

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum .gitignore cmd/api/main.go
git commit -m "feat: initialize Go project with minimal health endpoint"
```

---

## Task 2: Configuration

**Files:**
- Create: `internal/config/config.go`, `internal/config/config_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/config/config_test.go`:

```go
package config_test

import (
	"testing"

	"github.com/jwilson/dronerx/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "" {
		t.Errorf("expected empty DatabaseURL, got %s", cfg.DatabaseURL)
	}
	if cfg.NATSUrl != "nats://localhost:4222" {
		t.Errorf("expected default NATS URL, got %s", cfg.NATSUrl)
	}
	if cfg.TickerInterval != 30 {
		t.Errorf("expected default ticker interval 30, got %d", cfg.TickerInterval)
	}
	if cfg.WebhookURL != "" {
		t.Errorf("expected empty WebhookURL, got %s", cfg.WebhookURL)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/dronerx")
	t.Setenv("NATS_URL", "nats://nats:4222")
	t.Setenv("TICKER_INTERVAL", "10")
	t.Setenv("WEBHOOK_URL", "https://example.com/hook")

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://localhost:5432/dronerx" {
		t.Errorf("expected DatabaseURL from env, got %s", cfg.DatabaseURL)
	}
	if cfg.NATSUrl != "nats://nats:4222" {
		t.Errorf("expected NATS URL from env, got %s", cfg.NATSUrl)
	}
	if cfg.TickerInterval != 10 {
		t.Errorf("expected ticker interval 10, got %d", cfg.TickerInterval)
	}
	if cfg.WebhookURL != "https://example.com/hook" {
		t.Errorf("expected WebhookURL from env, got %s", cfg.WebhookURL)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/jwilson/git/dronerx
go test ./internal/config/ -v
```

Expected: FAIL — `package config not found` or similar.

- [ ] **Step 3: Write implementation**

Create `internal/config/config.go`:

```go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	DatabaseURL    string
	NATSUrl        string
	TickerInterval int
	WebhookURL     string
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		NATSUrl:        getEnv("NATS_URL", "nats://localhost:4222"),
		TickerInterval: getEnvInt("TICKER_INTERVAL", 30),
		WebhookURL:     getEnv("WEBHOOK_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/config/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config package with env var parsing"
```

---

## Task 3: Database Connection

**Files:**
- Create: `internal/database/database.go`
- Dependencies: `github.com/jackc/pgx/v5`

- [ ] **Step 1: Add pgx dependency**

```bash
cd /Users/jwilson/git/dronerx
go get github.com/jackc/pgx/v5/pgxpool
```

- [ ] **Step 2: Write database connection package**

Create `internal/database/database.go`:

```go
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing database URL: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/database/
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add internal/database/database.go go.mod go.sum
git commit -m "feat: add database connection package with pgx pool"
```

---

## Task 4: Database Migrations

**Files:**
- Create: `internal/database/migrate.go`, `internal/database/migrations/001_create_medicines.up.sql`, `internal/database/migrations/001_create_medicines.down.sql`, `internal/database/migrations/002_create_orders.up.sql`, `internal/database/migrations/002_create_orders.down.sql`
- Dependencies: `github.com/golang-migrate/migrate/v4`

- [ ] **Step 1: Add migrate dependency**

```bash
cd /Users/jwilson/git/dronerx
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/pgx/v5
go get github.com/golang-migrate/migrate/v4/source/iofs
```

- [ ] **Step 2: Create medicines migration (up)**

Create `internal/database/migrations/001_create_medicines.up.sql`:

```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE medicines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    in_stock BOOLEAN NOT NULL DEFAULT true,
    category TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO medicines (name, description, price, in_stock, category) VALUES
    ('Paracetamol 500mg', 'Pain relief and fever reduction tablets', 4.99, true, 'Pain Relief'),
    ('Ibuprofen 200mg', 'Anti-inflammatory pain relief', 5.49, true, 'Pain Relief'),
    ('Amoxicillin 250mg', 'Broad-spectrum antibiotic capsules', 8.99, true, 'Antibiotics'),
    ('Cetirizine 10mg', 'Non-drowsy antihistamine for allergies', 6.29, true, 'Allergy'),
    ('Omeprazole 20mg', 'Acid reflux and heartburn relief', 7.49, true, 'Digestive'),
    ('Loratadine 10mg', 'Allergy relief tablets', 5.99, true, 'Allergy'),
    ('Vitamin D3 1000IU', 'Daily vitamin D supplement', 3.99, true, 'Supplements'),
    ('Zinc 25mg', 'Immune support supplement', 4.49, true, 'Supplements'),
    ('Salbutamol Inhaler', 'Reliever inhaler for asthma', 12.99, true, 'Respiratory'),
    ('Throat Lozenges', 'Soothing lozenges for sore throat', 3.49, true, 'Respiratory');
```

- [ ] **Step 3: Create medicines migration (down)**

Create `internal/database/migrations/001_create_medicines.down.sql`:

```sql
DROP TABLE IF EXISTS medicines;
```

- [ ] **Step 4: Create orders migration (up)**

Create `internal/database/migrations/002_create_orders.up.sql`:

```sql
CREATE TYPE order_status AS ENUM ('placed', 'preparing', 'in-flight', 'delivered');

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_name TEXT NOT NULL,
    address TEXT NOT NULL,
    status order_status NOT NULL DEFAULT 'placed',
    estimated_delivery TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    medicine_id UUID NOT NULL REFERENCES medicines(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0)
);

CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_patient_name ON orders(patient_name);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
```

- [ ] **Step 5: Create orders migration (down)**

Create `internal/database/migrations/002_create_orders.down.sql`:

```sql
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS order_status;
```

- [ ] **Step 6: Write migration runner**

Create `internal/database/migrate.go`:

```go
package database

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Migrate(databaseURL string) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("creating migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, "pgx5://"+stripScheme(databaseURL))
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

func stripScheme(url string) string {
	for i := 0; i < len(url); i++ {
		if url[i] == '/' && i+1 < len(url) && url[i+1] == '/' {
			return url[i+2:]
		}
	}
	return url
}
```

- [ ] **Step 7: Verify it compiles**

```bash
go build ./internal/database/
```

Expected: No errors.

- [ ] **Step 8: Commit**

```bash
git add internal/database/ go.mod go.sum
git commit -m "feat: add database migrations with medicines seed data and orders schema"
```

---

## Task 5: Medicine Model + Queries

**Files:**
- Create: `internal/models/medicine.go`, `internal/models/medicine_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/models/medicine_test.go`:

```go
package models_test

import (
	"testing"

	"github.com/jwilson/dronerx/internal/models"
)

func TestMedicineFields(t *testing.T) {
	m := models.Medicine{
		ID:          "test-id",
		Name:        "Paracetamol",
		Description: "Pain relief",
		Price:       4.99,
		InStock:     true,
		Category:    "Pain Relief",
	}

	if m.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", m.ID)
	}
	if m.Name != "Paracetamol" {
		t.Errorf("expected Name Paracetamol, got %s", m.Name)
	}
	if m.Price != 4.99 {
		t.Errorf("expected Price 4.99, got %f", m.Price)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/models/ -v -run TestMedicineFields
```

Expected: FAIL — `package models not found`.

- [ ] **Step 3: Write implementation**

Create `internal/models/medicine.go`:

```go
package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Medicine struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	InStock     bool    `json:"in_stock"`
	Category    string  `json:"category"`
}

type MedicineStore struct {
	db *pgxpool.Pool
}

func NewMedicineStore(db *pgxpool.Pool) *MedicineStore {
	return &MedicineStore{db: db}
}

func (s *MedicineStore) List(ctx context.Context) ([]Medicine, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, name, description, price, in_stock, category
		 FROM medicines
		 WHERE in_stock = true
		 ORDER BY category, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medicines []Medicine
	for rows.Next() {
		var m Medicine
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Price, &m.InStock, &m.Category); err != nil {
			return nil, err
		}
		medicines = append(medicines, m)
	}
	return medicines, rows.Err()
}

func (s *MedicineStore) GetByID(ctx context.Context, id string) (*Medicine, error) {
	var m Medicine
	err := s.db.QueryRow(ctx,
		`SELECT id, name, description, price, in_stock, category
		 FROM medicines
		 WHERE id = $1`, id).
		Scan(&m.ID, &m.Name, &m.Description, &m.Price, &m.InStock, &m.Category)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/models/ -v -run TestMedicineFields
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/models/
git commit -m "feat: add medicine model with list and get-by-id queries"
```

---

## Task 6: Order Model + Status Enum

**Files:**
- Create: `internal/models/order.go`, `internal/models/order_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/models/order_test.go`:

```go
package models_test

import (
	"testing"

	"github.com/jwilson/dronerx/internal/models"
)

func TestOrderStatusProgression(t *testing.T) {
	tests := []struct {
		current  models.OrderStatus
		expected models.OrderStatus
		terminal bool
	}{
		{models.StatusPlaced, models.StatusPreparing, false},
		{models.StatusPreparing, models.StatusInFlight, false},
		{models.StatusInFlight, models.StatusDelivered, false},
		{models.StatusDelivered, "", true},
	}

	for _, tt := range tests {
		next, isTerminal := tt.current.Next()
		if isTerminal != tt.terminal {
			t.Errorf("status %s: expected terminal=%v, got %v", tt.current, tt.terminal, isTerminal)
		}
		if !tt.terminal && next != tt.expected {
			t.Errorf("status %s: expected next=%s, got %s", tt.current, tt.expected, next)
		}
	}
}

func TestOrderStatusIsValid(t *testing.T) {
	if !models.StatusPlaced.IsValid() {
		t.Error("placed should be valid")
	}
	if models.OrderStatus("invalid").IsValid() {
		t.Error("invalid should not be valid")
	}
}

func TestCreateOrderRequest(t *testing.T) {
	req := models.CreateOrderRequest{
		PatientName: "John",
		Address:     "123 Main St",
		Items: []models.OrderItemRequest{
			{MedicineID: "med-1", Quantity: 2},
		},
	}

	if err := req.Validate(); err != nil {
		t.Errorf("expected valid request, got error: %v", err)
	}
}

func TestCreateOrderRequest_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  models.CreateOrderRequest
	}{
		{"empty name", models.CreateOrderRequest{Address: "123 St", Items: []models.OrderItemRequest{{MedicineID: "m1", Quantity: 1}}}},
		{"empty address", models.CreateOrderRequest{PatientName: "John", Items: []models.OrderItemRequest{{MedicineID: "m1", Quantity: 1}}}},
		{"no items", models.CreateOrderRequest{PatientName: "John", Address: "123 St"}},
		{"zero quantity", models.CreateOrderRequest{PatientName: "John", Address: "123 St", Items: []models.OrderItemRequest{{MedicineID: "m1", Quantity: 0}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/models/ -v -run "TestOrderStatus|TestCreateOrder"
```

Expected: FAIL — types not defined.

- [ ] **Step 3: Write implementation**

Create `internal/models/order.go`:

```go
package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStatus string

const (
	StatusPlaced    OrderStatus = "placed"
	StatusPreparing OrderStatus = "preparing"
	StatusInFlight  OrderStatus = "in-flight"
	StatusDelivered OrderStatus = "delivered"
)

var statusOrder = []OrderStatus{StatusPlaced, StatusPreparing, StatusInFlight, StatusDelivered}

func (s OrderStatus) Next() (OrderStatus, bool) {
	for i, status := range statusOrder {
		if status == s && i+1 < len(statusOrder) {
			return statusOrder[i+1], false
		}
	}
	return "", true
}

func (s OrderStatus) IsValid() bool {
	for _, status := range statusOrder {
		if status == s {
			return true
		}
	}
	return false
}

type Order struct {
	ID                string      `json:"id"`
	PatientName       string      `json:"patient_name"`
	Address           string      `json:"address"`
	Status            OrderStatus `json:"status"`
	EstimatedDelivery *time.Time  `json:"estimated_delivery"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	Items             []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID         string  `json:"id"`
	OrderID    string  `json:"order_id"`
	MedicineID string  `json:"medicine_id"`
	Quantity   int     `json:"quantity"`
	Name       string  `json:"name,omitempty"`
	Price      float64 `json:"price,omitempty"`
}

type CreateOrderRequest struct {
	PatientName string             `json:"patient_name"`
	Address     string             `json:"address"`
	Items       []OrderItemRequest `json:"items"`
}

type OrderItemRequest struct {
	MedicineID string `json:"medicine_id"`
	Quantity   int    `json:"quantity"`
}

func (r *CreateOrderRequest) Validate() error {
	if r.PatientName == "" {
		return fmt.Errorf("patient_name is required")
	}
	if r.Address == "" {
		return fmt.Errorf("address is required")
	}
	if len(r.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}
	for i, item := range r.Items {
		if item.MedicineID == "" {
			return fmt.Errorf("item %d: medicine_id is required", i)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item %d: quantity must be positive", i)
		}
	}
	return nil
}

type OrderStore struct {
	db *pgxpool.Pool
}

func NewOrderStore(db *pgxpool.Pool) *OrderStore {
	return &OrderStore{db: db}
}

func (s *OrderStore) Create(ctx context.Context, req CreateOrderRequest, estimatedDelivery time.Time) (*Order, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var order Order
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (patient_name, address, status, estimated_delivery)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, patient_name, address, status, estimated_delivery, created_at, updated_at`,
		req.PatientName, req.Address, StatusPlaced, estimatedDelivery).
		Scan(&order.ID, &order.PatientName, &order.Address, &order.Status,
			&order.EstimatedDelivery, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	for _, item := range req.Items {
		var oi OrderItem
		err = tx.QueryRow(ctx,
			`INSERT INTO order_items (order_id, medicine_id, quantity)
			 VALUES ($1, $2, $3)
			 RETURNING id, order_id, medicine_id, quantity`,
			order.ID, item.MedicineID, item.Quantity).
			Scan(&oi.ID, &oi.OrderID, &oi.MedicineID, &oi.Quantity)
		if err != nil {
			return nil, fmt.Errorf("insert order item: %w", err)
		}
		order.Items = append(order.Items, oi)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &order, nil
}

func (s *OrderStore) GetByID(ctx context.Context, id string) (*Order, error) {
	var order Order
	err := s.db.QueryRow(ctx,
		`SELECT id, patient_name, address, status, estimated_delivery, created_at, updated_at
		 FROM orders WHERE id = $1`, id).
		Scan(&order.ID, &order.PatientName, &order.Address, &order.Status,
			&order.EstimatedDelivery, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx,
		`SELECT oi.id, oi.order_id, oi.medicine_id, oi.quantity, m.name, m.price
		 FROM order_items oi
		 JOIN medicines m ON m.id = oi.medicine_id
		 WHERE oi.order_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.MedicineID, &item.Quantity, &item.Name, &item.Price); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, rows.Err()
}

func (s *OrderStore) ListByPatient(ctx context.Context, patientName string) ([]Order, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, patient_name, address, status, estimated_delivery, created_at, updated_at
		 FROM orders
		 WHERE patient_name = $1
		 ORDER BY created_at DESC`, patientName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.PatientName, &o.Address, &o.Status,
			&o.EstimatedDelivery, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (s *OrderStore) ListByStatus(ctx context.Context, status OrderStatus) ([]Order, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, patient_name, address, status, estimated_delivery, created_at, updated_at
		 FROM orders
		 WHERE status = $1
		 ORDER BY created_at ASC`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.PatientName, &o.Address, &o.Status,
			&o.EstimatedDelivery, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (s *OrderStore) AdvanceStatus(ctx context.Context, id string) (*Order, error) {
	var current OrderStatus
	err := s.db.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1`, id).Scan(&current)
	if err != nil {
		return nil, fmt.Errorf("get current status: %w", err)
	}

	next, terminal := current.Next()
	if terminal {
		return nil, fmt.Errorf("order %s is already in terminal status %s", id, current)
	}

	var order Order
	err = s.db.QueryRow(ctx,
		`UPDATE orders SET status = $1, updated_at = now()
		 WHERE id = $2
		 RETURNING id, patient_name, address, status, estimated_delivery, created_at, updated_at`,
		next, id).
		Scan(&order.ID, &order.PatientName, &order.Address, &order.Status,
			&order.EstimatedDelivery, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("advance status: %w", err)
	}

	return &order, nil
}

func (s *OrderStore) ListNonTerminal(ctx context.Context) ([]Order, error) {
	return s.listByStatuses(ctx, StatusPlaced, StatusPreparing, StatusInFlight)
}

func (s *OrderStore) listByStatuses(ctx context.Context, statuses ...OrderStatus) ([]Order, error) {
	args := make([]interface{}, len(statuses))
	placeholders := ""
	for i, s := range statuses {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = s
	}

	rows, err := s.db.Query(ctx,
		fmt.Sprintf(`SELECT id, patient_name, address, status, estimated_delivery, created_at, updated_at
		 FROM orders
		 WHERE status IN (%s)
		 ORDER BY created_at ASC`, placeholders), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.PatientName, &o.Address, &o.Status,
			&o.EstimatedDelivery, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

// Needed for pgx to scan OrderStatus from postgres enum
func (s *OrderStatus) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		*s = OrderStatus(v)
		return nil
	case []byte:
		*s = OrderStatus(string(v))
		return nil
	}
	return fmt.Errorf("cannot scan %T into OrderStatus", src)
}

func (s OrderStatus) Value() (interface{}, error) {
	return string(s), nil
}

// Needed for pgx v5
func (s OrderStatus) TextValue() (pgx.TextValue, error) {
	return pgx.TextValue{String: string(s), Valid: true}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/models/ -v -run "TestOrderStatus|TestCreateOrder"
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/models/
git commit -m "feat: add order model with status progression, validation, and DB queries"
```

---

## Task 7: ETA Calculation

**Files:**
- Create: `internal/models/eta.go`, `internal/models/eta_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/models/eta_test.go`:

```go
package models_test

import (
	"testing"
	"time"

	"github.com/jwilson/dronerx/internal/models"
)

func TestCalculateETA_NewOrder(t *testing.T) {
	now := time.Now()
	eta := models.CalculateETA(now, 30)

	// 4 transitions × 30s = 120s total
	expectedDuration := 4 * 30 * time.Second
	expected := now.Add(expectedDuration)

	diff := eta.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected ETA near %v, got %v (diff: %v)", expected, eta, diff)
	}
}

func TestCalculateRemainingETA(t *testing.T) {
	tests := []struct {
		name     string
		status   models.OrderStatus
		interval int
		minSecs  float64
		maxSecs  float64
	}{
		{"placed has 3 transitions left", models.StatusPlaced, 30, 80, 100},
		{"preparing has 2 transitions left", models.StatusPreparing, 30, 50, 70},
		{"in-flight has 1 transition left", models.StatusInFlight, 30, 20, 40},
		{"delivered has 0 remaining", models.StatusDelivered, 30, -1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remaining := models.RemainingETA(tt.status, time.Now(), tt.interval)
			secs := remaining.Seconds()
			if secs < tt.minSecs || secs > tt.maxSecs {
				t.Errorf("expected remaining between %v-%vs, got %vs", tt.minSecs, tt.maxSecs, secs)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/models/ -v -run "TestCalculateETA|TestCalculateRemaining"
```

Expected: FAIL — functions not defined.

- [ ] **Step 3: Write implementation**

Create `internal/models/eta.go`:

```go
package models

import "time"

// transitionsRemaining returns how many status transitions remain before delivery.
func transitionsRemaining(status OrderStatus) int {
	switch status {
	case StatusPlaced:
		return 3
	case StatusPreparing:
		return 2
	case StatusInFlight:
		return 1
	case StatusDelivered:
		return 0
	default:
		return 0
	}
}

// CalculateETA returns the estimated delivery time for a new order.
// intervalSecs is the seconds between each status transition.
func CalculateETA(orderCreated time.Time, intervalSecs int) time.Time {
	totalTransitions := 4 // placed -> preparing -> in-flight -> delivered
	totalDuration := time.Duration(totalTransitions*intervalSecs) * time.Second
	return orderCreated.Add(totalDuration)
}

// RemainingETA returns the estimated time remaining until delivery.
func RemainingETA(status OrderStatus, updatedAt time.Time, intervalSecs int) time.Duration {
	remaining := transitionsRemaining(status)
	if remaining == 0 {
		return 0
	}
	totalRemaining := time.Duration(remaining*intervalSecs) * time.Second
	elapsed := time.Since(updatedAt)
	eta := totalRemaining - elapsed
	if eta < 0 {
		return 0
	}
	return eta
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/models/ -v -run "TestCalculateETA|TestCalculateRemaining"
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/models/eta.go internal/models/eta_test.go
git commit -m "feat: add ETA calculation for order delivery estimation"
```

---

## Task 8: Health Handler

**Files:**
- Create: `internal/handlers/health.go`, `internal/handlers/health_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/handlers/health_test.go`:

```go
package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/handlers"
)

type mockHealthChecker struct {
	dbOK   bool
	natsOK bool
}

func (m *mockHealthChecker) PingDB() error {
	if !m.dbOK {
		return fmt.Errorf("db down")
	}
	return nil
}

func (m *mockHealthChecker) PingNATS() error {
	if !m.natsOK {
		return fmt.Errorf("nats down")
	}
	return nil
}

func TestHealthHandler_AllHealthy(t *testing.T) {
	h := handlers.NewHealthHandler(&mockHealthChecker{dbOK: true, natsOK: true})

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)

	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
	if body["db"] != "ok" {
		t.Errorf("expected db ok, got %s", body["db"])
	}
	if body["nats"] != "ok" {
		t.Errorf("expected nats ok, got %s", body["nats"])
	}
}

func TestHealthHandler_DBDown(t *testing.T) {
	h := handlers.NewHealthHandler(&mockHealthChecker{dbOK: false, natsOK: true})

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)

	if body["status"] != "error" {
		t.Errorf("expected status error, got %s", body["status"])
	}
	if body["db"] != "error" {
		t.Errorf("expected db error, got %s", body["db"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/handlers/ -v -run TestHealth
```

Expected: FAIL — package not found.

- [ ] **Step 3: Write implementation**

Create `internal/handlers/health.go`:

```go
package handlers

import (
	"encoding/json"
	"net/http"
)

type HealthChecker interface {
	PingDB() error
	PingNATS() error
}

type HealthHandler struct {
	checker HealthChecker
}

func NewHealthHandler(checker HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbStatus := "ok"
	natsStatus := "ok"
	overall := "ok"
	statusCode := http.StatusOK

	if err := h.checker.PingDB(); err != nil {
		dbStatus = "error"
		overall = "error"
		statusCode = http.StatusServiceUnavailable
	}

	if err := h.checker.PingNATS(); err != nil {
		natsStatus = "error"
		overall = "error"
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"status": overall,
		"db":     dbStatus,
		"nats":   natsStatus,
	})
}
```

- [ ] **Step 4: Add missing import to test file**

Update `internal/handlers/health_test.go` — add `"fmt"` to the imports block:

```go
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/handlers"
)
```

- [ ] **Step 5: Run test to verify it passes**

```bash
go test ./internal/handlers/ -v -run TestHealth
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/handlers/
git commit -m "feat: add health endpoint with DB and NATS connectivity checks"
```

---

## Task 9: Medicine Handlers

**Files:**
- Create: `internal/handlers/medicine.go`, `internal/handlers/medicine_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/handlers/medicine_test.go`:

```go
package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/handlers"
	"github.com/jwilson/dronerx/internal/models"
)

type mockMedicineStore struct {
	medicines []models.Medicine
}

func (m *mockMedicineStore) List(ctx context.Context) ([]models.Medicine, error) {
	return m.medicines, nil
}

func (m *mockMedicineStore) GetByID(ctx context.Context, id string) (*models.Medicine, error) {
	for _, med := range m.medicines {
		if med.ID == id {
			return &med, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func TestListMedicines(t *testing.T) {
	store := &mockMedicineStore{
		medicines: []models.Medicine{
			{ID: "1", Name: "Paracetamol", Price: 4.99, InStock: true, Category: "Pain Relief"},
		},
	}
	h := handlers.NewMedicineHandler(store)

	req := httptest.NewRequest("GET", "/api/medicines", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body []models.Medicine
	json.NewDecoder(rec.Body).Decode(&body)

	if len(body) != 1 {
		t.Fatalf("expected 1 medicine, got %d", len(body))
	}
	if body[0].Name != "Paracetamol" {
		t.Errorf("expected Paracetamol, got %s", body[0].Name)
	}
}

func TestGetMedicine(t *testing.T) {
	store := &mockMedicineStore{
		medicines: []models.Medicine{
			{ID: "med-1", Name: "Ibuprofen", Price: 5.49, InStock: true, Category: "Pain Relief"},
		},
	}
	h := handlers.NewMedicineHandler(store)

	req := httptest.NewRequest("GET", "/api/medicines/med-1", nil)
	req.SetPathValue("id", "med-1")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body models.Medicine
	json.NewDecoder(rec.Body).Decode(&body)

	if body.Name != "Ibuprofen" {
		t.Errorf("expected Ibuprofen, got %s", body.Name)
	}
}

func TestGetMedicine_NotFound(t *testing.T) {
	store := &mockMedicineStore{medicines: []models.Medicine{}}
	h := handlers.NewMedicineHandler(store)

	req := httptest.NewRequest("GET", "/api/medicines/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/handlers/ -v -run "TestListMedicines|TestGetMedicine"
```

Expected: FAIL — `NewMedicineHandler` not defined.

- [ ] **Step 3: Write implementation**

Create `internal/handlers/medicine.go`:

```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jwilson/dronerx/internal/models"
)

type MedicineLister interface {
	List(ctx context.Context) ([]models.Medicine, error)
	GetByID(ctx context.Context, id string) (*models.Medicine, error)
}

type MedicineHandler struct {
	store MedicineLister
}

func NewMedicineHandler(store MedicineLister) *MedicineHandler {
	return &MedicineHandler{store: store}
}

func (h *MedicineHandler) List(w http.ResponseWriter, r *http.Request) {
	medicines, err := h.store.List(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medicines)
}

func (h *MedicineHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	medicine, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "medicine not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medicine)
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/handlers/ -v -run "TestListMedicines|TestGetMedicine"
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/medicine.go internal/handlers/medicine_test.go
git commit -m "feat: add medicine list and get-by-id HTTP handlers"
```

---

## Task 10: Order Handlers

**Files:**
- Create: `internal/handlers/order.go`, `internal/handlers/order_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/handlers/order_test.go`:

```go
package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jwilson/dronerx/internal/handlers"
	"github.com/jwilson/dronerx/internal/models"
)

type mockOrderStore struct {
	orders []models.Order
}

func (m *mockOrderStore) Create(ctx context.Context, req models.CreateOrderRequest, eta time.Time) (*models.Order, error) {
	order := &models.Order{
		ID:                "order-1",
		PatientName:       req.PatientName,
		Address:           req.Address,
		Status:            models.StatusPlaced,
		EstimatedDelivery: &eta,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	return order, nil
}

func (m *mockOrderStore) GetByID(ctx context.Context, id string) (*models.Order, error) {
	for _, o := range m.orders {
		if o.ID == id {
			return &o, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockOrderStore) ListByPatient(ctx context.Context, name string) ([]models.Order, error) {
	var result []models.Order
	for _, o := range m.orders {
		if o.PatientName == name {
			result = append(result, o)
		}
	}
	return result, nil
}

func TestCreateOrder(t *testing.T) {
	store := &mockOrderStore{}
	h := handlers.NewOrderHandler(store, 30)

	body := `{"patient_name":"Alice","address":"42 High St","items":[{"medicine_id":"med-1","quantity":2}]}`
	req := httptest.NewRequest("POST", "/api/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var order models.Order
	json.NewDecoder(rec.Body).Decode(&order)

	if order.PatientName != "Alice" {
		t.Errorf("expected Alice, got %s", order.PatientName)
	}
	if order.Status != models.StatusPlaced {
		t.Errorf("expected placed, got %s", order.Status)
	}
}

func TestCreateOrder_InvalidBody(t *testing.T) {
	store := &mockOrderStore{}
	h := handlers.NewOrderHandler(store, 30)

	body := `{"patient_name":"","address":"","items":[]}`
	req := httptest.NewRequest("POST", "/api/orders", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetOrder(t *testing.T) {
	now := time.Now()
	eta := now.Add(2 * time.Minute)
	store := &mockOrderStore{
		orders: []models.Order{
			{ID: "order-1", PatientName: "Bob", Status: models.StatusPreparing, EstimatedDelivery: &eta, UpdatedAt: now},
		},
	}
	h := handlers.NewOrderHandler(store, 30)

	req := httptest.NewRequest("GET", "/api/orders/order-1", nil)
	req.SetPathValue("id", "order-1")
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&body)

	if body["patient_name"] != "Bob" {
		t.Errorf("expected Bob, got %v", body["patient_name"])
	}
}

func TestListOrders(t *testing.T) {
	store := &mockOrderStore{
		orders: []models.Order{
			{ID: "o1", PatientName: "Alice", Status: models.StatusPlaced},
			{ID: "o2", PatientName: "Alice", Status: models.StatusDelivered},
			{ID: "o3", PatientName: "Bob", Status: models.StatusPlaced},
		},
	}
	h := handlers.NewOrderHandler(store, 30)

	req := httptest.NewRequest("GET", "/api/orders?patient_name=Alice", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body []models.Order
	json.NewDecoder(rec.Body).Decode(&body)

	if len(body) != 2 {
		t.Errorf("expected 2 orders for Alice, got %d", len(body))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/handlers/ -v -run "TestCreateOrder|TestGetOrder|TestListOrders"
```

Expected: FAIL — `NewOrderHandler` not defined.

- [ ] **Step 3: Write implementation**

Create `internal/handlers/order.go`:

```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jwilson/dronerx/internal/models"
)

type OrderStorer interface {
	Create(ctx context.Context, req models.CreateOrderRequest, eta time.Time) (*models.Order, error)
	GetByID(ctx context.Context, id string) (*models.Order, error)
	ListByPatient(ctx context.Context, name string) ([]models.Order, error)
}

type OrderHandler struct {
	store          OrderStorer
	tickerInterval int
}

func NewOrderHandler(store OrderStorer, tickerInterval int) *OrderHandler {
	return &OrderHandler{store: store, tickerInterval: tickerInterval}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	eta := models.CalculateETA(time.Now(), h.tickerInterval)
	order, err := h.store.Create(r.Context(), req, eta)
	if err != nil {
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	order, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	type OrderResponse struct {
		models.Order
		RemainingETA float64 `json:"remaining_eta_seconds"`
	}

	resp := OrderResponse{
		Order:        *order,
		RemainingETA: models.RemainingETA(order.Status, order.UpdatedAt, h.tickerInterval).Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	patientName := r.URL.Query().Get("patient_name")
	if patientName == "" {
		http.Error(w, "patient_name query parameter required", http.StatusBadRequest)
		return
	}

	orders, err := h.store.ListByPatient(r.Context(), patientName)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/handlers/ -v -run "TestCreateOrder|TestGetOrder|TestListOrders"
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/order.go internal/handlers/order_test.go
git commit -m "feat: add order create, get, and list HTTP handlers"
```

---

## Task 11: NATS Events Publisher

**Files:**
- Create: `internal/events/publisher.go`
- Dependencies: `github.com/nats-io/nats.go`

- [ ] **Step 1: Add NATS dependency**

```bash
cd /Users/jwilson/git/dronerx
go get github.com/nats-io/nats.go
```

- [ ] **Step 2: Write publisher**

Create `internal/events/publisher.go`:

```go
package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type OrderStatusEvent struct {
	OrderID           string `json:"order_id"`
	Status            string `json:"status"`
	EstimatedDelivery string `json:"estimated_delivery,omitempty"`
	UpdatedAt         string `json:"updated_at"`
}

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

func (p *Publisher) PublishOrderStatus(orderID, status string, estimatedDelivery *time.Time, updatedAt time.Time) error {
	event := OrderStatusEvent{
		OrderID:   orderID,
		Status:    status,
		UpdatedAt: updatedAt.Format(time.RFC3339),
	}
	if estimatedDelivery != nil {
		event.EstimatedDelivery = estimatedDelivery.Format(time.RFC3339)
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	subject := fmt.Sprintf("orders.%s.status", orderID)
	if err := p.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}

	return nil
}

func ConnectNATS(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS at %s: %w", url, err)
	}
	return nc, nil
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/events/
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add internal/events/ go.mod go.sum
git commit -m "feat: add NATS event publisher for order status updates"
```

---

## Task 12: State Machine Ticker

**Files:**
- Create: `internal/statemachine/ticker.go`, `internal/statemachine/ticker_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/statemachine/ticker_test.go`:

```go
package statemachine_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jwilson/dronerx/internal/models"
	"github.com/jwilson/dronerx/internal/statemachine"
)

type mockAdvancer struct {
	mu       sync.Mutex
	advanced []string
	orders   []models.Order
}

func (m *mockAdvancer) ListNonTerminal(ctx context.Context) ([]models.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.orders, nil
}

func (m *mockAdvancer) AdvanceStatus(ctx context.Context, id string) (*models.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.advanced = append(m.advanced, id)

	for i, o := range m.orders {
		if o.ID == id {
			next, terminal := o.Status.Next()
			if !terminal {
				m.orders[i].Status = next
			}
			return &m.orders[i], nil
		}
	}
	return nil, nil
}

type mockPublisher struct {
	mu        sync.Mutex
	published []string
}

func (m *mockPublisher) PublishOrderStatus(orderID, status string, eta *time.Time, updatedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.published = append(m.published, orderID+":"+status)
	return nil
}

type mockNotifier struct {
	mu       sync.Mutex
	notified []string
}

func (m *mockNotifier) NotifyDelivered(orderID, patientName, address string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notified = append(m.notified, orderID)
}

func TestTicker_AdvancesOrders(t *testing.T) {
	now := time.Now()
	advancer := &mockAdvancer{
		orders: []models.Order{
			{ID: "o1", PatientName: "Alice", Status: models.StatusPlaced, UpdatedAt: now},
		},
	}
	pub := &mockPublisher{}
	notifier := &mockNotifier{}

	ticker := statemachine.NewTicker(advancer, pub, notifier, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go ticker.Start(ctx)
	time.Sleep(2500 * time.Millisecond)
	cancel()

	advancer.mu.Lock()
	defer advancer.mu.Unlock()

	if len(advancer.advanced) < 1 {
		t.Errorf("expected at least 1 advancement, got %d", len(advancer.advanced))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/statemachine/ -v -run TestTicker -timeout 10s
```

Expected: FAIL — package not found.

- [ ] **Step 3: Write implementation**

Create `internal/statemachine/ticker.go`:

```go
package statemachine

import (
	"context"
	"log"
	"time"

	"github.com/jwilson/dronerx/internal/models"
)

type OrderAdvancer interface {
	ListNonTerminal(ctx context.Context) ([]models.Order, error)
	AdvanceStatus(ctx context.Context, id string) (*models.Order, error)
}

type StatusPublisher interface {
	PublishOrderStatus(orderID, status string, eta *time.Time, updatedAt time.Time) error
}

type DeliveryNotifier interface {
	NotifyDelivered(orderID, patientName, address string)
}

type Ticker struct {
	advancer   OrderAdvancer
	publisher  StatusPublisher
	notifier   DeliveryNotifier
	intervalSec int
}

func NewTicker(advancer OrderAdvancer, publisher StatusPublisher, notifier DeliveryNotifier, intervalSec int) *Ticker {
	return &Ticker{
		advancer:   advancer,
		publisher:  publisher,
		notifier:   notifier,
		intervalSec: intervalSec,
	}
}

func (t *Ticker) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(t.intervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.tick(ctx)
		}
	}
}

func (t *Ticker) tick(ctx context.Context) {
	orders, err := t.advancer.ListNonTerminal(ctx)
	if err != nil {
		log.Printf("ticker: listing orders: %v", err)
		return
	}

	for _, order := range orders {
		updated, err := t.advancer.AdvanceStatus(ctx, order.ID)
		if err != nil {
			log.Printf("ticker: advancing order %s: %v", order.ID, err)
			continue
		}

		if err := t.publisher.PublishOrderStatus(
			updated.ID,
			string(updated.Status),
			updated.EstimatedDelivery,
			updated.UpdatedAt,
		); err != nil {
			log.Printf("ticker: publishing status for %s: %v", updated.ID, err)
		}

		if updated.Status == models.StatusDelivered {
			t.notifier.NotifyDelivered(updated.ID, updated.PatientName, updated.Address)
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/statemachine/ -v -run TestTicker -timeout 10s
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/statemachine/
git commit -m "feat: add state machine ticker that auto-advances order statuses"
```

---

## Task 13: WebSocket Tracking Handler

**Files:**
- Create: `internal/handlers/tracking.go`
- Dependencies: `github.com/coder/websocket`

- [ ] **Step 1: Add websocket dependency**

```bash
cd /Users/jwilson/git/dronerx
go get github.com/coder/websocket
```

- [ ] **Step 2: Write tracking handler**

Create `internal/handlers/tracking.go`:

```go
package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/nats-io/nats.go"
)

type TrackingHandler struct {
	nc *nats.Conn
}

func NewTrackingHandler(nc *nats.Conn) *TrackingHandler {
	return &TrackingHandler{nc: nc}
}

func (h *TrackingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	if orderID == "" {
		http.Error(w, "order ID required", http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
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

	// Keep connection open, read messages (ping/pong handled automatically)
	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			break
		}
	}

	conn.Close(websocket.StatusNormalClosure, "")
}

// TrackingEvent is the JSON structure sent over WebSocket (same as NATS event)
type TrackingEvent struct {
	OrderID           string `json:"order_id"`
	Status            string `json:"status"`
	EstimatedDelivery string `json:"estimated_delivery,omitempty"`
	UpdatedAt         string `json:"updated_at"`
}

func ParseTrackingEvent(data []byte) (*TrackingEvent, error) {
	var event TrackingEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/handlers/
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add internal/handlers/tracking.go go.mod go.sum
git commit -m "feat: add WebSocket tracking handler with NATS subscription relay"
```

---

## Task 14: Webhook Notifier

**Files:**
- Create: `internal/webhook/notifier.go`, `internal/webhook/notifier_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/webhook/notifier_test.go`:

```go
package webhook_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jwilson/dronerx/internal/webhook"
)

func TestNotifier_SendsWebhook(t *testing.T) {
	var received map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := webhook.NewNotifier(server.URL)
	n.NotifyDelivered("order-1", "Alice", "42 High St")

	if received["order_id"] != "order-1" {
		t.Errorf("expected order_id order-1, got %s", received["order_id"])
	}
	if received["patient_name"] != "Alice" {
		t.Errorf("expected patient_name Alice, got %s", received["patient_name"])
	}
	if received["event"] != "delivered" {
		t.Errorf("expected event delivered, got %s", received["event"])
	}
}

func TestNotifier_EmptyURL_NoOp(t *testing.T) {
	n := webhook.NewNotifier("")
	// Should not panic or error
	n.NotifyDelivered("order-1", "Alice", "42 High St")
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/webhook/ -v
```

Expected: FAIL — package not found.

- [ ] **Step 3: Write implementation**

Create `internal/webhook/notifier.go`:

```go
package webhook

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Notifier struct {
	url    string
	client *http.Client
}

func NewNotifier(url string) *Notifier {
	return &Notifier{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *Notifier) NotifyDelivered(orderID, patientName, address string) {
	if n.url == "" {
		return
	}

	payload := map[string]string{
		"event":        "delivered",
		"order_id":     orderID,
		"patient_name": patientName,
		"address":      address,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("webhook: marshal error: %v", err)
		return
	}

	resp, err := n.client.Post(n.url, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("webhook: POST to %s failed: %v", n.url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("webhook: POST to %s returned %d", n.url, resp.StatusCode)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./internal/webhook/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/webhook/
git commit -m "feat: add webhook notifier for delivery events"
```

---

## Task 15: Wire Up main.go

**Files:**
- Modify: `cmd/api/main.go`

- [ ] **Step 1: Rewrite main.go to wire all components**

Replace `cmd/api/main.go` with:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jwilson/dronerx/internal/config"
	"github.com/jwilson/dronerx/internal/database"
	"github.com/jwilson/dronerx/internal/events"
	"github.com/jwilson/dronerx/internal/handlers"
	"github.com/jwilson/dronerx/internal/models"
	"github.com/jwilson/dronerx/internal/statemachine"
	"github.com/jwilson/dronerx/internal/webhook"
)

type appHealthChecker struct {
	db   *database.Pool
	nats *events.Connection
}

func main() {
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Database
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if err := database.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("Running migrations: %v", err)
	}

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Connecting to database: %v", err)
	}
	defer db.Close()

	// NATS
	nc, err := events.ConnectNATS(cfg.NATSUrl)
	if err != nil {
		log.Fatalf("Connecting to NATS: %v", err)
	}
	defer nc.Close()

	// Stores
	medicineStore := models.NewMedicineStore(db)
	orderStore := models.NewOrderStore(db)

	// Events
	publisher := events.NewPublisher(nc)

	// Webhook
	notifier := webhook.NewNotifier(cfg.WebhookURL)

	// State machine
	ticker := statemachine.NewTicker(orderStore, publisher, notifier, cfg.TickerInterval)
	go ticker.Start(ctx)

	// Handlers
	medicineHandler := handlers.NewMedicineHandler(medicineStore)
	orderHandler := handlers.NewOrderHandler(orderStore, cfg.TickerInterval)
	trackingHandler := handlers.NewTrackingHandler(nc)
	healthHandler := handlers.NewHealthHandler(&healthChecker{db: db, nc: nc})

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler.ServeHTTP)
	mux.HandleFunc("GET /api/medicines", medicineHandler.List)
	mux.HandleFunc("GET /api/medicines/{id}", medicineHandler.GetByID)
	mux.HandleFunc("POST /api/orders", orderHandler.Create)
	mux.HandleFunc("GET /api/orders/{id}", orderHandler.GetByID)
	mux.HandleFunc("GET /api/orders", orderHandler.List)
	mux.Handle("GET /api/orders/{id}/track", trackingHandler)

	// CORS middleware for frontend
	handler := corsMiddleware(mux)

	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{Addr: addr, Handler: handler}

	go func() {
		log.Printf("API server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
}

type healthChecker struct {
	db *database.Pool
	nc *events.Connection
}

// Note: these types need to be aliases. Update the types after verifying compilation.
// For now, use the concrete types from the packages.

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

**Note:** This initial wiring will need type adjustments once we verify the package interfaces align. The `healthChecker` struct needs to satisfy the `HealthChecker` interface. We'll fix any compilation issues in the next step.

- [ ] **Step 2: Add health checker adapter using concrete types**

Update the `healthChecker` in `cmd/api/main.go` to use concrete types:

```go
type healthChecker struct {
	db *pgxpool.Pool
	nc *nats.Conn
}

func (h *healthChecker) PingDB() error {
	return h.db.Ping(context.Background())
}

func (h *healthChecker) PingNATS() error {
	if !h.nc.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}
	return nil
}
```

Add the necessary imports:

```go
import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)
```

And update the health handler construction:

```go
healthHandler := handlers.NewHealthHandler(&healthChecker{db: db, nc: nc})
```

Where `db` is `*pgxpool.Pool` (returned by `database.Connect`) and `nc` is `*nats.Conn` (returned by `events.ConnectNATS`).

- [ ] **Step 3: Verify it compiles**

```bash
go build ./cmd/api/
```

Fix any remaining type mismatches. The main issues will be ensuring `database.Connect` returns `*pgxpool.Pool` and `events.ConnectNATS` returns `*nats.Conn` — which they already do.

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add cmd/api/main.go
git commit -m "feat: wire up all components in main.go with graceful shutdown"
```

---

## Task 16: SvelteKit Frontend Scaffolding

**Files:**
- Create: `frontend/` (SvelteKit project)

- [ ] **Step 1: Scaffold SvelteKit project**

```bash
cd /Users/jwilson/git/dronerx
npx sv create frontend --template minimal --types ts
```

Select: Tailwind CSS when prompted (or add manually in step 2).

- [ ] **Step 2: Add Tailwind CSS**

```bash
cd /Users/jwilson/git/dronerx/frontend
npx sv add tailwindcss
```

- [ ] **Step 3: Install dependencies**

```bash
cd /Users/jwilson/git/dronerx/frontend
npm install
```

- [ ] **Step 4: Add adapter-node for server-side rendering in containers**

```bash
cd /Users/jwilson/git/dronerx/frontend
npm install -D @sveltejs/adapter-node
```

Update `frontend/svelte.config.js`:

```js
import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter()
	}
};

export default config;
```

- [ ] **Step 5: Configure API proxy for development**

Create `frontend/vite.config.ts`:

```ts
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			'/api': 'http://localhost:8080',
			'/healthz': 'http://localhost:8080'
		}
	}
});
```

- [ ] **Step 6: Verify dev server starts**

```bash
cd /Users/jwilson/git/dronerx/frontend
npm run dev
```

Expected: Dev server starts on http://localhost:5173.

- [ ] **Step 7: Commit**

```bash
cd /Users/jwilson/git/dronerx
git add frontend/
git commit -m "feat: scaffold SvelteKit frontend with TypeScript, Tailwind, and adapter-node"
```

---

## Task 17: Frontend Types and API Client

**Files:**
- Create: `frontend/src/lib/types.ts`, `frontend/src/lib/api.ts`

- [ ] **Step 1: Create shared types**

Create `frontend/src/lib/types.ts`:

```ts
export interface Medicine {
	id: string;
	name: string;
	description: string;
	price: number;
	in_stock: boolean;
	category: string;
}

export interface Order {
	id: string;
	patient_name: string;
	address: string;
	status: OrderStatus;
	estimated_delivery: string | null;
	remaining_eta_seconds?: number;
	created_at: string;
	updated_at: string;
	items?: OrderItem[];
}

export interface OrderItem {
	id: string;
	order_id: string;
	medicine_id: string;
	quantity: number;
	name?: string;
	price?: number;
}

export type OrderStatus = 'placed' | 'preparing' | 'in-flight' | 'delivered';

export interface CreateOrderRequest {
	patient_name: string;
	address: string;
	items: { medicine_id: string; quantity: number }[];
}

export interface TrackingEvent {
	order_id: string;
	status: OrderStatus;
	estimated_delivery?: string;
	updated_at: string;
}

export const STATUS_LABELS: Record<OrderStatus, string> = {
	placed: 'Order Placed',
	preparing: 'Preparing',
	'in-flight': 'Drone In Flight',
	delivered: 'Delivered'
};

export const STATUS_ORDER: OrderStatus[] = ['placed', 'preparing', 'in-flight', 'delivered'];
```

- [ ] **Step 2: Create API client**

Create `frontend/src/lib/api.ts`:

```ts
import type { Medicine, Order, CreateOrderRequest } from './types';

const BASE_URL = '/api';

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
	const response = await fetch(url, init);
	if (!response.ok) {
		const text = await response.text();
		throw new Error(`${response.status}: ${text}`);
	}
	return response.json();
}

export async function listMedicines(): Promise<Medicine[]> {
	return fetchJSON<Medicine[]>(`${BASE_URL}/medicines`);
}

export async function getMedicine(id: string): Promise<Medicine> {
	return fetchJSON<Medicine>(`${BASE_URL}/medicines/${id}`);
}

export async function createOrder(req: CreateOrderRequest): Promise<Order> {
	return fetchJSON<Order>(`${BASE_URL}/orders`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(req)
	});
}

export async function getOrder(id: string): Promise<Order> {
	return fetchJSON<Order>(`${BASE_URL}/orders/${id}`);
}

export async function listOrders(patientName: string): Promise<Order[]> {
	return fetchJSON<Order[]>(`${BASE_URL}/orders?patient_name=${encodeURIComponent(patientName)}`);
}

export function connectTracking(orderID: string): WebSocket {
	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	return new WebSocket(`${protocol}//${window.location.host}/api/orders/${orderID}/track`);
}
```

- [ ] **Step 3: Create cart store**

Create `frontend/src/lib/stores/cart.ts`:

```ts
import { writable, derived } from 'svelte/store';
import type { Medicine } from '$lib/types';

interface CartItem {
	medicine: Medicine;
	quantity: number;
}

function createCart() {
	const { subscribe, set, update } = writable<CartItem[]>([]);

	return {
		subscribe,
		add(medicine: Medicine) {
			update((items) => {
				const existing = items.find((i) => i.medicine.id === medicine.id);
				if (existing) {
					existing.quantity += 1;
					return [...items];
				}
				return [...items, { medicine, quantity: 1 }];
			});
		},
		remove(medicineId: string) {
			update((items) => items.filter((i) => i.medicine.id !== medicineId));
		},
		updateQuantity(medicineId: string, quantity: number) {
			update((items) => {
				if (quantity <= 0) {
					return items.filter((i) => i.medicine.id !== medicineId);
				}
				const item = items.find((i) => i.medicine.id === medicineId);
				if (item) item.quantity = quantity;
				return [...items];
			});
		},
		clear() {
			set([]);
		}
	};
}

export const cart = createCart();

export const cartTotal = derived(cart, ($cart) =>
	$cart.reduce((sum, item) => sum + item.medicine.price * item.quantity, 0)
);

export const cartCount = derived(cart, ($cart) =>
	$cart.reduce((sum, item) => sum + item.quantity, 0)
);
```

- [ ] **Step 4: Verify TypeScript compiles**

```bash
cd /Users/jwilson/git/dronerx/frontend
npx svelte-check
```

Expected: No errors.

- [ ] **Step 5: Commit**

```bash
cd /Users/jwilson/git/dronerx
git add frontend/src/lib/
git commit -m "feat: add frontend types, API client, and cart store"
```

---

## Task 18: Frontend — Medicine Browsing Page

**Files:**
- Create: `frontend/src/routes/+page.ts`, `frontend/src/routes/+page.svelte`, `frontend/src/lib/components/MedicineCard.svelte`

- [ ] **Step 1: Create page load function**

Create `frontend/src/routes/+page.ts`:

```ts
import { listMedicines } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	const medicines = await listMedicines();

	const categories = [...new Set(medicines.map((m) => m.category))];

	return { medicines, categories };
};
```

- [ ] **Step 2: Create MedicineCard component**

Create `frontend/src/lib/components/MedicineCard.svelte`:

```svelte
<script lang="ts">
	import type { Medicine } from '$lib/types';
	import { cart } from '$lib/stores/cart';

	export let medicine: Medicine;

	function addToCart() {
		cart.add(medicine);
	}
</script>

<div class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition-shadow hover:shadow-md">
	<div class="mb-2 text-xs font-medium uppercase tracking-wide text-teal-600">
		{medicine.category}
	</div>
	<h3 class="mb-1 text-lg font-semibold text-slate-900">{medicine.name}</h3>
	<p class="mb-4 text-sm text-slate-500">{medicine.description}</p>
	<div class="flex items-center justify-between">
		<span class="text-xl font-bold text-slate-900">£{medicine.price.toFixed(2)}</span>
		<button
			on:click={addToCart}
			class="rounded-lg bg-teal-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-teal-700 active:bg-teal-800"
		>
			Add to Cart
		</button>
	</div>
</div>
```

- [ ] **Step 3: Create landing page**

Create `frontend/src/routes/+page.svelte`:

```svelte
<script lang="ts">
	import type { PageData } from './$types';
	import MedicineCard from '$lib/components/MedicineCard.svelte';
	import { cart, cartCount, cartTotal } from '$lib/stores/cart';

	export let data: PageData;

	let selectedCategory = 'All';

	$: filteredMedicines =
		selectedCategory === 'All'
			? data.medicines
			: data.medicines.filter((m) => m.category === selectedCategory);
</script>

<div class="min-h-screen bg-slate-50">
	<!-- Header -->
	<header class="border-b border-slate-200 bg-white">
		<div class="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
			<div>
				<h1 class="text-2xl font-bold text-slate-900">DroneRx</h1>
				<p class="text-sm text-slate-500">Medicine delivered by drone</p>
			</div>
			<div class="flex items-center gap-4">
				{#if $cartCount > 0}
					<a
						href="/order"
						class="flex items-center gap-2 rounded-lg bg-teal-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-teal-700"
					>
						Cart ({$cartCount}) — £{$cartTotal.toFixed(2)}
					</a>
				{/if}
				<a href="/orders" class="text-sm font-medium text-slate-600 hover:text-slate-900">
					My Orders
				</a>
			</div>
		</div>
	</header>

	<!-- Category Filter -->
	<div class="mx-auto max-w-6xl px-6 pt-6">
		<div class="flex gap-2">
			<button
				on:click={() => (selectedCategory = 'All')}
				class="rounded-full px-4 py-1.5 text-sm font-medium transition-colors
					{selectedCategory === 'All'
					? 'bg-teal-600 text-white'
					: 'bg-white text-slate-600 hover:bg-slate-100'}"
			>
				All
			</button>
			{#each data.categories as category}
				<button
					on:click={() => (selectedCategory = category)}
					class="rounded-full px-4 py-1.5 text-sm font-medium transition-colors
						{selectedCategory === category
						? 'bg-teal-600 text-white'
						: 'bg-white text-slate-600 hover:bg-slate-100'}"
				>
					{category}
				</button>
			{/each}
		</div>
	</div>

	<!-- Medicine Grid -->
	<main class="mx-auto max-w-6xl px-6 py-6">
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each filteredMedicines as medicine (medicine.id)}
				<MedicineCard {medicine} />
			{/each}
		</div>
	</main>
</div>
```

- [ ] **Step 4: Verify dev server renders**

```bash
cd /Users/jwilson/git/dronerx/frontend
npm run dev
```

Open http://localhost:5173 — should render the page structure (API calls will fail without the backend, but the page should load).

- [ ] **Step 5: Commit**

```bash
cd /Users/jwilson/git/dronerx
git add frontend/src/routes/+page.ts frontend/src/routes/+page.svelte frontend/src/lib/components/MedicineCard.svelte
git commit -m "feat: add medicine browsing page with category filtering and cart"
```

---

## Task 19: Frontend — Order Form Page

**Files:**
- Create: `frontend/src/routes/order/+page.svelte`, `frontend/src/routes/order/+page.ts`

- [ ] **Step 1: Create page load (no server data needed)**

Create `frontend/src/routes/order/+page.ts`:

```ts
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	return {};
};
```

- [ ] **Step 2: Create order form page**

Create `frontend/src/routes/order/+page.svelte`:

```svelte
<script lang="ts">
	import { goto } from '$app/navigation';
	import { cart, cartTotal } from '$lib/stores/cart';
	import { createOrder } from '$lib/api';

	let patientName = '';
	let address = '';
	let submitting = false;
	let error = '';

	async function handleSubmit() {
		if (!patientName.trim() || !address.trim()) {
			error = 'Please fill in all fields';
			return;
		}

		if ($cart.length === 0) {
			error = 'Your cart is empty';
			return;
		}

		submitting = true;
		error = '';

		try {
			const order = await createOrder({
				patient_name: patientName.trim(),
				address: address.trim(),
				items: $cart.map((item) => ({
					medicine_id: item.medicine.id,
					quantity: item.quantity
				}))
			});

			cart.clear();
			goto(`/order/${order.id}`);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to place order';
			submitting = false;
		}
	}
</script>

<div class="min-h-screen bg-slate-50">
	<header class="border-b border-slate-200 bg-white">
		<div class="mx-auto flex max-w-2xl items-center gap-4 px-6 py-4">
			<a href="/" class="text-sm text-slate-500 hover:text-slate-900">← Back</a>
			<h1 class="text-xl font-bold text-slate-900">Place Order</h1>
		</div>
	</header>

	<main class="mx-auto max-w-2xl px-6 py-8">
		{#if $cart.length === 0}
			<div class="rounded-xl border border-slate-200 bg-white p-8 text-center">
				<p class="text-slate-500">Your cart is empty.</p>
				<a href="/" class="mt-4 inline-block text-sm font-medium text-teal-600 hover:text-teal-700">
					Browse medicines
				</a>
			</div>
		{:else}
			<!-- Cart Summary -->
			<div class="mb-6 rounded-xl border border-slate-200 bg-white p-5">
				<h2 class="mb-3 text-lg font-semibold text-slate-900">Your Items</h2>
				{#each $cart as item (item.medicine.id)}
					<div class="flex items-center justify-between border-b border-slate-100 py-3 last:border-0">
						<div>
							<p class="font-medium text-slate-900">{item.medicine.name}</p>
							<p class="text-sm text-slate-500">£{item.medicine.price.toFixed(2)} × {item.quantity}</p>
						</div>
						<div class="flex items-center gap-2">
							<button
								on:click={() => cart.updateQuantity(item.medicine.id, item.quantity - 1)}
								class="h-8 w-8 rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50"
							>
								−
							</button>
							<span class="w-8 text-center text-sm font-medium">{item.quantity}</span>
							<button
								on:click={() => cart.updateQuantity(item.medicine.id, item.quantity + 1)}
								class="h-8 w-8 rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50"
							>
								+
							</button>
							<button
								on:click={() => cart.remove(item.medicine.id)}
								class="ml-2 text-sm text-red-500 hover:text-red-700"
							>
								Remove
							</button>
						</div>
					</div>
				{/each}
				<div class="mt-3 flex justify-between pt-3">
					<span class="font-semibold text-slate-900">Total</span>
					<span class="text-xl font-bold text-slate-900">£{$cartTotal.toFixed(2)}</span>
				</div>
			</div>

			<!-- Delivery Details -->
			<form on:submit|preventDefault={handleSubmit} class="rounded-xl border border-slate-200 bg-white p-5">
				<h2 class="mb-4 text-lg font-semibold text-slate-900">Delivery Details</h2>

				{#if error}
					<div class="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</div>
				{/if}

				<div class="mb-4">
					<label for="name" class="mb-1 block text-sm font-medium text-slate-700">Your Name</label>
					<input
						id="name"
						bind:value={patientName}
						type="text"
						placeholder="Enter your full name"
						class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
					/>
				</div>

				<div class="mb-6">
					<label for="address" class="mb-1 block text-sm font-medium text-slate-700">
						Delivery Address
					</label>
					<textarea
						id="address"
						bind:value={address}
						rows="3"
						placeholder="Enter your full delivery address"
						class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
					/>
				</div>

				<button
					type="submit"
					disabled={submitting}
					class="w-full rounded-lg bg-teal-600 py-3 text-sm font-medium text-white transition-colors hover:bg-teal-700 disabled:bg-slate-300"
				>
					{submitting ? 'Placing order...' : 'Place Order — Drone Delivery'}
				</button>
			</form>
		{/if}
	</main>
</div>
```

- [ ] **Step 3: Commit**

```bash
cd /Users/jwilson/git/dronerx
git add frontend/src/routes/order/
git commit -m "feat: add order form page with cart summary and delivery details"
```

---

## Task 20: Frontend — Order Status Page with Tracking

**Files:**
- Create: `frontend/src/routes/order/[id]/+page.ts`, `frontend/src/routes/order/[id]/+page.svelte`, `frontend/src/lib/components/StatusTracker.svelte`

- [ ] **Step 1: Create StatusTracker component**

Create `frontend/src/lib/components/StatusTracker.svelte`:

```svelte
<script lang="ts">
	import { STATUS_ORDER, STATUS_LABELS, type OrderStatus } from '$lib/types';

	export let currentStatus: OrderStatus;

	$: currentIndex = STATUS_ORDER.indexOf(currentStatus);
</script>

<div class="flex items-center justify-between">
	{#each STATUS_ORDER as status, i}
		<div class="flex flex-1 flex-col items-center">
			<!-- Step circle -->
			<div
				class="flex h-10 w-10 items-center justify-center rounded-full text-sm font-bold transition-colors duration-500
					{i <= currentIndex
					? 'bg-teal-600 text-white'
					: 'bg-slate-200 text-slate-400'}"
			>
				{#if i < currentIndex}
					✓
				{:else}
					{i + 1}
				{/if}
			</div>
			<!-- Label -->
			<span
				class="mt-2 text-xs font-medium transition-colors duration-500
					{i <= currentIndex ? 'text-teal-700' : 'text-slate-400'}"
			>
				{STATUS_LABELS[status]}
			</span>
		</div>
		<!-- Connector line -->
		{#if i < STATUS_ORDER.length - 1}
			<div
				class="mx-2 mt-[-1.5rem] h-0.5 flex-1 transition-colors duration-500
					{i < currentIndex ? 'bg-teal-600' : 'bg-slate-200'}"
			/>
		{/if}
	{/each}
</div>
```

- [ ] **Step 2: Create page load function**

Create `frontend/src/routes/order/[id]/+page.ts`:

```ts
import { getOrder } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const order = await getOrder(params.id);
	return { order };
};
```

- [ ] **Step 3: Create order status page**

Create `frontend/src/routes/order/[id]/+page.svelte`:

```svelte
<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { PageData } from './$types';
	import type { OrderStatus, TrackingEvent } from '$lib/types';
	import { getOrder, connectTracking } from '$lib/api';
	import StatusTracker from '$lib/components/StatusTracker.svelte';

	export let data: PageData;

	let order = data.order;
	let ws: WebSocket | null = null;
	let trackingActive = false;
	let pollInterval: ReturnType<typeof setInterval> | null = null;

	function formatETA(seconds?: number): string {
		if (!seconds || seconds <= 0) return 'Arrived';
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		if (mins > 0) return `~${mins}m ${secs}s`;
		return `~${secs}s`;
	}

	function startTracking() {
		try {
			ws = connectTracking(order.id);

			ws.onopen = () => {
				trackingActive = true;
			};

			ws.onmessage = (event) => {
				const data: TrackingEvent = JSON.parse(event.data);
				order = { ...order, status: data.status, updated_at: data.updated_at };
			};

			ws.onerror = () => {
				trackingActive = false;
				startPolling();
			};

			ws.onclose = () => {
				trackingActive = false;
			};
		} catch {
			startPolling();
		}
	}

	function startPolling() {
		if (pollInterval) return;
		pollInterval = setInterval(async () => {
			try {
				order = await getOrder(order.id);
				if (order.status === 'delivered' && pollInterval) {
					clearInterval(pollInterval);
					pollInterval = null;
				}
			} catch {
				// Silently retry
			}
		}, 5000);
	}

	onMount(() => {
		if (order.status !== 'delivered') {
			startTracking();
		}
	});

	onDestroy(() => {
		ws?.close();
		if (pollInterval) clearInterval(pollInterval);
	});
</script>

<div class="min-h-screen bg-slate-50">
	<header class="border-b border-slate-200 bg-white">
		<div class="mx-auto flex max-w-2xl items-center gap-4 px-6 py-4">
			<a href="/" class="text-sm text-slate-500 hover:text-slate-900">← Home</a>
			<h1 class="text-xl font-bold text-slate-900">Order Status</h1>
		</div>
	</header>

	<main class="mx-auto max-w-2xl px-6 py-8">
		<!-- Status Tracker -->
		<div class="mb-6 rounded-xl border border-slate-200 bg-white p-6">
			<StatusTracker currentStatus={order.status} />

			<div class="mt-6 text-center">
				{#if order.status === 'delivered'}
					<p class="text-lg font-semibold text-teal-600">Your order has been delivered!</p>
				{:else}
					<p class="text-sm text-slate-500">
						Estimated delivery: {formatETA(order.remaining_eta_seconds)}
					</p>
				{/if}

				{#if trackingActive}
					<span class="mt-2 inline-flex items-center gap-1 text-xs text-teal-600">
						<span class="h-2 w-2 animate-pulse rounded-full bg-teal-500"></span>
						Live tracking
					</span>
				{/if}
			</div>
		</div>

		<!-- Order Details -->
		<div class="rounded-xl border border-slate-200 bg-white p-5">
			<h2 class="mb-3 text-lg font-semibold text-slate-900">Order Details</h2>
			<dl class="space-y-2 text-sm">
				<div class="flex justify-between">
					<dt class="text-slate-500">Order ID</dt>
					<dd class="font-mono text-slate-900">{order.id.slice(0, 8)}</dd>
				</div>
				<div class="flex justify-between">
					<dt class="text-slate-500">Name</dt>
					<dd class="text-slate-900">{order.patient_name}</dd>
				</div>
				<div class="flex justify-between">
					<dt class="text-slate-500">Address</dt>
					<dd class="text-slate-900">{order.address}</dd>
				</div>
			</dl>

			{#if order.items && order.items.length > 0}
				<h3 class="mb-2 mt-4 text-sm font-semibold text-slate-900">Items</h3>
				{#each order.items as item}
					<div class="flex justify-between border-t border-slate-100 py-2 text-sm">
						<span class="text-slate-700">{item.name || item.medicine_id} × {item.quantity}</span>
						{#if item.price}
							<span class="text-slate-900">£{(item.price * item.quantity).toFixed(2)}</span>
						{/if}
					</div>
				{/each}
			{/if}
		</div>
	</main>
</div>
```

- [ ] **Step 4: Commit**

```bash
cd /Users/jwilson/git/dronerx
git add frontend/src/routes/order/\[id\]/ frontend/src/lib/components/StatusTracker.svelte
git commit -m "feat: add order status page with live tracking and status stepper"
```

---

## Task 21: Frontend — Order History Page

**Files:**
- Create: `frontend/src/routes/orders/+page.svelte`, `frontend/src/routes/orders/+page.ts`

- [ ] **Step 1: Create page load**

Create `frontend/src/routes/orders/+page.ts`:

```ts
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	return {};
};
```

- [ ] **Step 2: Create order history page**

Create `frontend/src/routes/orders/+page.svelte`:

```svelte
<script lang="ts">
	import { listOrders } from '$lib/api';
	import { STATUS_LABELS, type Order } from '$lib/types';

	let patientName = '';
	let orders: Order[] = [];
	let searched = false;
	let loading = false;

	async function search() {
		if (!patientName.trim()) return;
		loading = true;
		try {
			orders = await listOrders(patientName.trim());
			searched = true;
		} catch {
			orders = [];
			searched = true;
		} finally {
			loading = false;
		}
	}

	function statusColor(status: string): string {
		switch (status) {
			case 'delivered':
				return 'bg-green-100 text-green-700';
			case 'in-flight':
				return 'bg-blue-100 text-blue-700';
			case 'preparing':
				return 'bg-amber-100 text-amber-700';
			default:
				return 'bg-slate-100 text-slate-700';
		}
	}
</script>

<div class="min-h-screen bg-slate-50">
	<header class="border-b border-slate-200 bg-white">
		<div class="mx-auto flex max-w-2xl items-center gap-4 px-6 py-4">
			<a href="/" class="text-sm text-slate-500 hover:text-slate-900">← Home</a>
			<h1 class="text-xl font-bold text-slate-900">My Orders</h1>
		</div>
	</header>

	<main class="mx-auto max-w-2xl px-6 py-8">
		<form on:submit|preventDefault={search} class="mb-6 flex gap-2">
			<input
				bind:value={patientName}
				type="text"
				placeholder="Enter your name to find orders"
				class="flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
			/>
			<button
				type="submit"
				disabled={loading}
				class="rounded-lg bg-teal-600 px-4 py-2 text-sm font-medium text-white hover:bg-teal-700 disabled:bg-slate-300"
			>
				{loading ? 'Searching...' : 'Search'}
			</button>
		</form>

		{#if searched && orders.length === 0}
			<p class="text-center text-sm text-slate-500">No orders found for "{patientName}".</p>
		{/if}

		{#each orders as order (order.id)}
			<a
				href="/order/{order.id}"
				class="mb-3 block rounded-xl border border-slate-200 bg-white p-4 transition-shadow hover:shadow-md"
			>
				<div class="flex items-center justify-between">
					<div>
						<p class="font-mono text-sm text-slate-500">{order.id.slice(0, 8)}</p>
						<p class="text-sm text-slate-700">{order.address}</p>
					</div>
					<span class="rounded-full px-3 py-1 text-xs font-medium {statusColor(order.status)}">
						{STATUS_LABELS[order.status]}
					</span>
				</div>
				<p class="mt-2 text-xs text-slate-400">
					{new Date(order.created_at).toLocaleString()}
				</p>
			</a>
		{/each}
	</main>
</div>
```

- [ ] **Step 3: Commit**

```bash
cd /Users/jwilson/git/dronerx
git add frontend/src/routes/orders/
git commit -m "feat: add order history page with name-based search"
```

---

## Task 22: Dockerfiles

**Files:**
- Create: `Dockerfile.api`, `Dockerfile.frontend`

- [ ] **Step 1: Create Go API Dockerfile**

Create `Dockerfile.api`:

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api/

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=builder /api /api

EXPOSE 8080

ENTRYPOINT ["/api"]
```

- [ ] **Step 2: Create SvelteKit frontend Dockerfile**

Create `Dockerfile.frontend`:

```dockerfile
FROM node:20-alpine AS builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ .
RUN npm run build

FROM node:20-alpine

WORKDIR /app

COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json .
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000

ENV PORT=3000
CMD ["node", "build"]
```

- [ ] **Step 3: Verify Docker builds**

```bash
cd /Users/jwilson/git/dronerx
docker build -f Dockerfile.api -t dronerx-api:dev .
docker build -f Dockerfile.frontend -t dronerx-frontend:dev .
```

Expected: Both images build successfully.

- [ ] **Step 4: Commit**

```bash
git add Dockerfile.api Dockerfile.frontend
git commit -m "feat: add multi-stage Dockerfiles for API and frontend"
```

---

## Task 23: Helm Chart Foundation

**Files:**
- Create: `chart/Chart.yaml`, `chart/values.yaml`, `chart/templates/_helpers.tpl`

- [ ] **Step 1: Create Chart.yaml with subcharts**

Create `chart/Chart.yaml`:

```yaml
apiVersion: v2
name: dronerx
description: Medicine delivery by drone
type: application
version: 0.1.0
appVersion: "0.1.0"

dependencies:
  - name: cloudnative-pg
    version: "0.22.0"
    repository: "https://cloudnative-pg.github.io/charts"
    condition: cloudnativepg.enabled
  - name: nats
    version: "1.2.6"
    repository: "https://nats-io.github.io/k8s/helm/charts"
    condition: nats.enabled
```

**Note:** Check the actual latest chart versions at install time and update if needed.

- [ ] **Step 2: Create values.yaml**

Create `chart/values.yaml`:

```yaml
# DroneRx Application Values

# API configuration
api:
  image:
    repository: ghcr.io/jwilson/dronerx-api
    tag: "latest"
    pullPolicy: IfNotPresent
  replicas: 1
  port: 8080
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  tickerInterval: 30
  webhookURL: ""

# Frontend configuration
frontend:
  image:
    repository: ghcr.io/jwilson/dronerx-frontend
    tag: "latest"
    pullPolicy: IfNotPresent
  replicas: 1
  port: 3000
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 200m
      memory: 128Mi

# Service configuration
service:
  type: ClusterIP
  apiPort: 8080
  frontendPort: 3000

# Ingress configuration
ingress:
  enabled: false
  hostname: ""
  tls:
    mode: "self-signed"  # auto | manual | self-signed
    secretName: ""       # for manual mode
    email: ""            # for auto mode (Let's Encrypt)

# CloudNativePG (embedded PostgreSQL)
cloudnativepg:
  enabled: true

postgresql:
  enabled: true
  instances: 1
  storage:
    size: 1Gi

# External database (when cloudnativepg.enabled=false and postgresql.enabled=false)
externalDatabase:
  host: ""
  port: 5432
  name: "dronerx"
  user: "dronerx"
  password: ""

# NATS
nats:
  enabled: true
```

- [ ] **Step 3: Create template helpers**

Create `chart/templates/_helpers.tpl`:

```yaml
{{/*
Expand the name of the chart.
*/}}
{{- define "dronerx.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "dronerx.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "dronerx.labels" -}}
helm.sh/chart: {{ include "dronerx.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Chart label
*/}}
{{- define "dronerx.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
API labels
*/}}
{{- define "dronerx.api.labels" -}}
{{ include "dronerx.labels" . }}
app.kubernetes.io/name: {{ include "dronerx.name" . }}-api
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: api
{{- end }}

{{/*
API selector labels
*/}}
{{- define "dronerx.api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "dronerx.name" . }}-api
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Frontend labels
*/}}
{{- define "dronerx.frontend.labels" -}}
{{ include "dronerx.labels" . }}
app.kubernetes.io/name: {{ include "dronerx.name" . }}-frontend
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: frontend
{{- end }}

{{/*
Frontend selector labels
*/}}
{{- define "dronerx.frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "dronerx.name" . }}-frontend
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Database URL
*/}}
{{- define "dronerx.databaseURL" -}}
{{- if .Values.postgresql.enabled }}
postgres://dronerx:$(DB_PASSWORD)@{{ include "dronerx.fullname" . }}-pg-rw:5432/dronerx?sslmode=disable
{{- else }}
postgres://{{ .Values.externalDatabase.user }}:$(DB_PASSWORD)@{{ .Values.externalDatabase.host }}:{{ .Values.externalDatabase.port }}/{{ .Values.externalDatabase.name }}?sslmode=disable
{{- end }}
{{- end }}

{{/*
NATS URL
*/}}
{{- define "dronerx.natsURL" -}}
nats://{{ .Release.Name }}-nats:4222
{{- end }}
```

- [ ] **Step 4: Build chart dependencies**

```bash
cd /Users/jwilson/git/dronerx/chart
helm dependency build
```

Expected: Dependencies downloaded to `chart/charts/`.

- [ ] **Step 5: Commit**

```bash
cd /Users/jwilson/git/dronerx
git add chart/Chart.yaml chart/values.yaml chart/templates/_helpers.tpl chart/Chart.lock
echo "chart/charts/" >> .gitignore
git add .gitignore
git commit -m "feat: add Helm chart foundation with CloudNativePG and NATS subcharts"
```

---

## Task 24: Helm Templates — Deployments and Services

**Files:**
- Create: `chart/templates/api-deployment.yaml`, `chart/templates/api-service.yaml`, `chart/templates/frontend-deployment.yaml`, `chart/templates/frontend-service.yaml`, `chart/templates/configmap-api.yaml`, `chart/templates/postgres-cluster.yaml`

- [ ] **Step 1: Create API ConfigMap**

Create `chart/templates/configmap-api.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "dronerx.fullname" . }}-api-config
  labels:
    {{- include "dronerx.api.labels" . | nindent 4 }}
data:
  PORT: "{{ .Values.api.port }}"
  NATS_URL: {{ include "dronerx.natsURL" . | quote }}
  TICKER_INTERVAL: "{{ .Values.api.tickerInterval }}"
  {{- if .Values.api.webhookURL }}
  WEBHOOK_URL: {{ .Values.api.webhookURL | quote }}
  {{- end }}
```

- [ ] **Step 2: Create CloudNativePG Cluster CR**

Create `chart/templates/postgres-cluster.yaml`:

```yaml
{{- if .Values.postgresql.enabled }}
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: {{ include "dronerx.fullname" . }}-pg
  labels:
    {{- include "dronerx.labels" . | nindent 4 }}
spec:
  instances: {{ .Values.postgresql.instances }}
  storage:
    size: {{ .Values.postgresql.storage.size }}
  bootstrap:
    initdb:
      database: dronerx
      owner: dronerx
{{- end }}
```

- [ ] **Step 3: Create API Deployment with init container**

Create `chart/templates/api-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "dronerx.fullname" . }}-api
  labels:
    {{- include "dronerx.api.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.api.replicas }}
  selector:
    matchLabels:
      {{- include "dronerx.api.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "dronerx.api.selectorLabels" . | nindent 8 }}
    spec:
      initContainers:
        - name: wait-for-db
          image: busybox:1.36
          command:
            - sh
            - -c
            - |
              {{- if .Values.postgresql.enabled }}
              until nc -z {{ include "dronerx.fullname" . }}-pg-rw 5432; do
                echo "Waiting for database..."
                sleep 2
              done
              {{- else }}
              until nc -z {{ .Values.externalDatabase.host }} {{ .Values.externalDatabase.port }}; do
                echo "Waiting for external database..."
                sleep 2
              done
              {{- end }}
              echo "Database is ready"
      containers:
        - name: api
          image: "{{ .Values.api.image.repository }}:{{ .Values.api.image.tag }}"
          imagePullPolicy: {{ .Values.api.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.api.port }}
              protocol: TCP
          envFrom:
            - configMapRef:
                name: {{ include "dronerx.fullname" . }}-api-config
          env:
            - name: DATABASE_URL
              value: {{ include "dronerx.databaseURL" . | quote }}
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  {{- if .Values.postgresql.enabled }}
                  name: {{ include "dronerx.fullname" . }}-pg-app
                  key: password
                  {{- else }}
                  name: {{ include "dronerx.fullname" . }}-external-db
                  key: password
                  {{- end }}
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 10
            periodSeconds: 15
          readinessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            {{- toYaml .Values.api.resources | nindent 12 }}
```

- [ ] **Step 4: Create API Service**

Create `chart/templates/api-service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ include "dronerx.fullname" . }}-api
  labels:
    {{- include "dronerx.api.labels" . | nindent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.apiPort }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "dronerx.api.selectorLabels" . | nindent 4 }}
```

- [ ] **Step 5: Create Frontend Deployment**

Create `chart/templates/frontend-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "dronerx.fullname" . }}-frontend
  labels:
    {{- include "dronerx.frontend.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.frontend.replicas }}
  selector:
    matchLabels:
      {{- include "dronerx.frontend.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "dronerx.frontend.selectorLabels" . | nindent 8 }}
    spec:
      containers:
        - name: frontend
          image: "{{ .Values.frontend.image.repository }}:{{ .Values.frontend.image.tag }}"
          imagePullPolicy: {{ .Values.frontend.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ .Values.frontend.port }}
              protocol: TCP
          env:
            - name: PORT
              value: "{{ .Values.frontend.port }}"
            - name: API_URL
              value: "http://{{ include "dronerx.fullname" . }}-api:{{ .Values.service.apiPort }}"
          livenessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 10
            periodSeconds: 15
          readinessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            {{- toYaml .Values.frontend.resources | nindent 12 }}
```

- [ ] **Step 6: Create Frontend Service**

Create `chart/templates/frontend-service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ include "dronerx.fullname" . }}-frontend
  labels:
    {{- include "dronerx.frontend.labels" . | nindent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.frontendPort }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "dronerx.frontend.selectorLabels" . | nindent 4 }}
```

- [ ] **Step 7: Run helm lint**

```bash
cd /Users/jwilson/git/dronerx
helm lint chart/
```

Expected: No errors (warnings are OK for missing subcharts).

- [ ] **Step 8: Commit**

```bash
git add chart/templates/
git commit -m "feat: add Helm templates for API and frontend deployments, services, and PostgreSQL cluster"
```

---

## Task 25: Ingress with TLS Modes

**Files:**
- Create: `chart/templates/ingress.yaml`, `chart/templates/self-signed-cert-job.yaml`

- [ ] **Step 1: Create Ingress template with 3 TLS modes**

Create `chart/templates/ingress.yaml`:

```yaml
{{- if .Values.ingress.enabled }}
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "dronerx.fullname" . }}
  labels:
    {{- include "dronerx.labels" . | nindent 4 }}
  annotations:
    {{- if eq .Values.ingress.tls.mode "auto" }}
    cert-manager.io/cluster-issuer: letsencrypt-prod
    {{- end }}
spec:
  {{- if ne .Values.ingress.tls.mode "" }}
  tls:
    - hosts:
        - {{ .Values.ingress.hostname | quote }}
      secretName: {{ include "dronerx.tlsSecretName" . }}
  {{- end }}
  rules:
    - host: {{ .Values.ingress.hostname | quote }}
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: {{ include "dronerx.fullname" . }}-api
                port:
                  number: {{ .Values.service.apiPort }}
          - path: /healthz
            pathType: Exact
            backend:
              service:
                name: {{ include "dronerx.fullname" . }}-api
                port:
                  number: {{ .Values.service.apiPort }}
          - path: /
            pathType: Prefix
            backend:
              service:
                name: {{ include "dronerx.fullname" . }}-frontend
                port:
                  number: {{ .Values.service.frontendPort }}
{{- end }}
```

- [ ] **Step 2: Add TLS secret name helper to _helpers.tpl**

Append to `chart/templates/_helpers.tpl`:

```yaml

{{/*
TLS secret name based on mode
*/}}
{{- define "dronerx.tlsSecretName" -}}
{{- if eq .Values.ingress.tls.mode "manual" }}
{{- .Values.ingress.tls.secretName }}
{{- else if eq .Values.ingress.tls.mode "auto" }}
{{- printf "%s-tls" (include "dronerx.fullname" .) }}
{{- else }}
{{- printf "%s-self-signed-tls" (include "dronerx.fullname" .) }}
{{- end }}
{{- end }}
```

- [ ] **Step 3: Create self-signed cert generation Job**

Create `chart/templates/self-signed-cert-job.yaml`:

```yaml
{{- if and .Values.ingress.enabled (eq .Values.ingress.tls.mode "self-signed") }}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "dronerx.fullname" . }}-gen-cert
  labels:
    {{- include "dronerx.labels" . | nindent 4 }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-delete-policy": before-hook-creation
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: gen-cert
          image: alpine/openssl:latest
          command:
            - sh
            - -c
            - |
              openssl req -x509 -nodes -days 365 \
                -newkey rsa:2048 \
                -keyout /certs/tls.key \
                -out /certs/tls.crt \
                -subj "/CN={{ .Values.ingress.hostname }}" \
                -addext "subjectAltName=DNS:{{ .Values.ingress.hostname }}"
          volumeMounts:
            - name: certs
              mountPath: /certs
      volumes:
        - name: certs
          secret:
            secretName: {{ include "dronerx.fullname" . }}-self-signed-tls
            optional: true
  backoffLimit: 1
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "dronerx.fullname" . }}-self-signed-tls
  labels:
    {{- include "dronerx.labels" . | nindent 4 }}
type: kubernetes.io/tls
data:
  tls.crt: ""
  tls.key: ""
{{- end }}
```

- [ ] **Step 4: Run helm lint**

```bash
helm lint chart/
```

Expected: No errors.

- [ ] **Step 5: Commit**

```bash
cd /Users/jwilson/git/dronerx
git add chart/templates/ingress.yaml chart/templates/self-signed-cert-job.yaml chart/templates/_helpers.tpl
git commit -m "feat: add ingress with 3 TLS modes (auto, manual, self-signed)"
```

---

## Task 26: Preflight and Support Bundle Placeholders

**Files:**
- Create: `chart/templates/_preflight.tpl`, `chart/templates/_supportbundle.tpl`

These are placeholder templates that will be fully implemented in Phase 4 (Tier 3). Including them now so `helm lint` passes and the chart structure is complete.

- [ ] **Step 1: Create preflight placeholder**

Create `chart/templates/_preflight.tpl`:

```yaml
{{/*
Preflight spec — fully implemented in Phase 4 (Tier 3: Support It)
*/}}
{{- define "dronerx.preflight" -}}
apiVersion: troubleshoot.sh/v1beta2
kind: Preflight
metadata:
  name: {{ include "dronerx.fullname" . }}-preflight
spec:
  analyzers: []
{{- end }}
```

- [ ] **Step 2: Create support bundle placeholder**

Create `chart/templates/_supportbundle.tpl`:

```yaml
{{/*
Support bundle spec — fully implemented in Phase 4 (Tier 3: Support It)
*/}}
{{- define "dronerx.supportbundle" -}}
apiVersion: troubleshoot.sh/v1beta2
kind: SupportBundle
metadata:
  name: {{ include "dronerx.fullname" . }}-support-bundle
spec:
  collectors: []
  analyzers: []
{{- end }}
```

- [ ] **Step 3: Commit**

```bash
git add chart/templates/_preflight.tpl chart/templates/_supportbundle.tpl
git commit -m "feat: add preflight and support bundle placeholder templates"
```

---

## Task 27: values.schema.json

**Files:**
- Create: `chart/values.schema.json`

- [ ] **Step 1: Create JSON schema for values validation**

Create `chart/values.schema.json`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "api": {
      "type": "object",
      "properties": {
        "image": {
          "type": "object",
          "properties": {
            "repository": { "type": "string" },
            "tag": { "type": "string" },
            "pullPolicy": { "type": "string", "enum": ["Always", "IfNotPresent", "Never"] }
          },
          "required": ["repository", "tag"]
        },
        "replicas": { "type": "integer", "minimum": 1 },
        "port": { "type": "integer" },
        "tickerInterval": { "type": "integer", "minimum": 1 },
        "webhookURL": { "type": "string" }
      }
    },
    "frontend": {
      "type": "object",
      "properties": {
        "image": {
          "type": "object",
          "properties": {
            "repository": { "type": "string" },
            "tag": { "type": "string" },
            "pullPolicy": { "type": "string", "enum": ["Always", "IfNotPresent", "Never"] }
          },
          "required": ["repository", "tag"]
        },
        "replicas": { "type": "integer", "minimum": 1 },
        "port": { "type": "integer" }
      }
    },
    "service": {
      "type": "object",
      "properties": {
        "type": { "type": "string", "enum": ["ClusterIP", "NodePort", "LoadBalancer"] }
      }
    },
    "ingress": {
      "type": "object",
      "properties": {
        "enabled": { "type": "boolean" },
        "hostname": { "type": "string" },
        "tls": {
          "type": "object",
          "properties": {
            "mode": { "type": "string", "enum": ["auto", "manual", "self-signed"] },
            "secretName": { "type": "string" },
            "email": { "type": "string" }
          }
        }
      }
    },
    "postgresql": {
      "type": "object",
      "properties": {
        "enabled": { "type": "boolean" },
        "instances": { "type": "integer", "minimum": 1 },
        "storage": {
          "type": "object",
          "properties": {
            "size": { "type": "string" }
          }
        }
      }
    },
    "externalDatabase": {
      "type": "object",
      "properties": {
        "host": { "type": "string" },
        "port": { "type": "integer" },
        "name": { "type": "string" },
        "user": { "type": "string" },
        "password": { "type": "string" }
      }
    },
    "nats": {
      "type": "object",
      "properties": {
        "enabled": { "type": "boolean" }
      }
    }
  }
}
```

- [ ] **Step 2: Run helm lint to verify schema**

```bash
helm lint chart/
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add chart/values.schema.json
git commit -m "feat: add values.schema.json for Helm values validation"
```

---

## Task 28: Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Create Makefile**

Create `Makefile`:

```makefile
.PHONY: build build-api build-frontend lint lint-go lint-frontend lint-helm test test-go test-frontend clean

# Docker image settings
REGISTRY ?= ghcr.io/jwilson
API_IMAGE ?= $(REGISTRY)/dronerx-api
FRONTEND_IMAGE ?= $(REGISTRY)/dronerx-frontend
TAG ?= $(shell git rev-parse --short HEAD)

## Build

build: build-api build-frontend

build-api:
	docker build -f Dockerfile.api -t $(API_IMAGE):$(TAG) -t $(API_IMAGE):latest .

build-frontend:
	docker build -f Dockerfile.frontend -t $(FRONTEND_IMAGE):$(TAG) -t $(FRONTEND_IMAGE):latest .

## Lint

lint: lint-go lint-frontend lint-helm

lint-go:
	cd . && go vet ./...

lint-frontend:
	cd frontend && npx svelte-check

lint-helm:
	helm lint chart/

## Test

test: test-go test-frontend

test-go:
	go test ./... -v

test-frontend:
	cd frontend && npm test

## Clean

clean:
	rm -rf frontend/build frontend/.svelte-kit
	docker rmi $(API_IMAGE):$(TAG) $(FRONTEND_IMAGE):$(TAG) 2>/dev/null || true
```

- [ ] **Step 2: Verify make targets**

```bash
cd /Users/jwilson/git/dronerx
make lint-go
make lint-helm
```

Expected: Both pass without errors.

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "feat: add Makefile with build, lint, test, and clean targets"
```

---

## Task 29: Final Verification

- [ ] **Step 1: Run all Go tests**

```bash
cd /Users/jwilson/git/dronerx
go test ./... -v
```

Expected: All unit tests pass.

- [ ] **Step 2: Run helm lint**

```bash
helm lint chart/
```

Expected: No errors.

- [ ] **Step 3: Verify Docker images build**

```bash
make build
```

Expected: Both images build successfully.

- [ ] **Step 4: Verify git history is clean**

```bash
git log --oneline
```

Expected: Series of focused, incremental commits.

---

## Rubric Coverage (Tier 0)

| Requirement | Task(s) | Status |
|------------|---------|--------|
| 0.1 Custom web app with stateful component | Tasks 1–15 (Go + SvelteKit + Postgres) | Covered |
| 0.2 Helm chart, helm lint, values.schema.json | Tasks 23–27 | Covered |
| 0.3 2 open-source subcharts, embedded + BYO | Task 23 (CloudNativePG + NATS), Task 24 (BYO toggle) | Covered |
| 0.4 K8s best practices (probes, resources, data persistence, /healthz) | Tasks 8, 24 | Covered |
| 0.5 HTTPS with 3 cert options | Task 25 | Covered |
| 0.6 App waits for database | Task 24 (init container) | Covered |
| 0.7 2+ user-facing demoable features | Tasks 18–21 (order placement, ETA, tracking, webhook) | Covered |
