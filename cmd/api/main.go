package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging for health check probes.
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Milliseconds()
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", duration,
		)
	})
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
	// Configure structured JSON logging.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// 1. Load config
	cfg := config.Load()

	slog.Info("starting dronerx",
		"port", cfg.Port,
		"ticker_interval", cfg.TickerInterval,
		"sdk_url", cfg.SDKUrl,
		"nats_url", cfg.NATSUrl,
	)

	// 2. Run migrations
	if cfg.DatabaseURL != "" {
		if err := database.Migrate(cfg.DatabaseURL); err != nil {
			slog.Error("migrations failed", "error", err)
			os.Exit(1)
		}
	}

	// 3. Connect to database
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 4. Connect to NATS
	nc, err := events.ConnectNATS(cfg.NATSUrl)
	if err != nil {
		slog.Error("nats connection failed", "error", err)
		os.Exit(1)
	}
	defer nc.Drain()

	// 4a. Create Replicated SDK client
	sdkClient := sdk.NewClient(cfg.SDKUrl)
	if cfg.LiveTrackingEnabled != "" {
		sdkClient.SetFeatureOverride("live_tracking_enabled", cfg.LiveTrackingEnabled)
	}

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
	go sdk.StartMetricsSender(ctx, sdkClient, orderStore, 30*time.Second)

	// 8. Create all handlers
	healthHandler := handlers.NewHealthHandler(&healthChecker{db: db, nc: nc})
	medicineHandler := handlers.NewMedicineHandler(medicineStore)
	orderHandler := handlers.NewOrderHandler(orderStore, cfg.TickerInterval)
	trackingHandler := handlers.NewTrackingHandler(nc, sdkClient)
	licenseHandler := handlers.NewLicenseHandler(sdkClient)
	updatesHandler := handlers.NewUpdatesHandler(sdkClient)
	adminHandler := handlers.NewAdminHandler(cfg.Namespace, cfg.SDKUrl, "sh", nil)

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
	mux.HandleFunc("POST /api/admin/support-bundle", adminHandler.GenerateSupportBundle)

	// 10. Wrap with logging then CORS middleware
	handler := loggingMiddleware(corsMiddleware(mux))

	// 11. Start HTTP server with graceful shutdown
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		slog.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
