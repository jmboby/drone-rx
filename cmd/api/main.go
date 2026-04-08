package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/jwilson/dronerx/internal/config"
	"github.com/jwilson/dronerx/internal/database"
	"github.com/jwilson/dronerx/internal/events"
	"github.com/jwilson/dronerx/internal/handlers"
	"github.com/jwilson/dronerx/internal/models"
	"github.com/jwilson/dronerx/internal/sdk"
	"github.com/jwilson/dronerx/internal/statemachine"
	"github.com/jwilson/dronerx/internal/webhook"
)

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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Run migrations
	if cfg.DatabaseURL != "" {
		if err := database.Migrate(cfg.DatabaseURL); err != nil {
			log.Fatalf("migrations: %v", err)
		}
	}

	// 3. Connect to database
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// 4. Connect to NATS
	nc, err := events.ConnectNATS(cfg.NATSUrl)
	if err != nil {
		log.Fatalf("nats: %v", err)
	}
	defer nc.Drain()

	// 4a. Create Replicated SDK client
	sdkClient := sdk.NewClient(cfg.SDKUrl)

	// 5. Create stores
	medicineStore := models.NewMedicineStore(db)
	orderStore := models.NewOrderStore(db)

	// 6. Create publisher and webhook notifier
	publisher := events.NewPublisher(nc)
	notifier := webhook.NewNotifier(cfg.WebhookURL)

	// 7. Start state machine ticker in a goroutine
	ticker := statemachine.NewTicker(orderStore, publisher, notifier, cfg.TickerInterval)
	go ticker.Start(ctx)

	// 7a. Start background metrics sender
	go sdk.StartMetricsSender(ctx, sdkClient, orderStore, 5*time.Minute)

	// 8. Create all handlers
	healthHandler := handlers.NewHealthHandler(&healthChecker{db: db, nc: nc})
	medicineHandler := handlers.NewMedicineHandler(medicineStore)
	orderHandler := handlers.NewOrderHandler(orderStore, cfg.TickerInterval)
	trackingHandler := handlers.NewTrackingHandler(nc, sdkClient)
	licenseHandler := handlers.NewLicenseHandler(sdkClient)
	updatesHandler := handlers.NewUpdatesHandler(sdkClient)

	// 9. Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler.ServeHTTP)
	mux.HandleFunc("GET /api/medicines", medicineHandler.List)
	mux.HandleFunc("GET /api/medicines/{id}", medicineHandler.GetByID)
	mux.HandleFunc("POST /api/orders", orderHandler.Create)
	mux.HandleFunc("GET /api/orders/{id}", orderHandler.GetByID)
	mux.HandleFunc("GET /api/orders", orderHandler.List)
	mux.Handle("GET /api/orders/{id}/track", trackingHandler)
	mux.HandleFunc("GET /api/license/status", licenseHandler.Status)
	mux.HandleFunc("GET /api/updates", updatesHandler.Check)

	// 10. Wrap with CORS middleware
	handler := corsMiddleware(mux)

	// 11. Start HTTP server with graceful shutdown
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	log.Println("Server stopped.")
}
