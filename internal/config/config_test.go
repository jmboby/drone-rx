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
	if cfg.TickerInterval != 5 {
		t.Errorf("expected default ticker interval 5, got %d", cfg.TickerInterval)
	}
	if cfg.WebhookURL != "" {
		t.Errorf("expected empty WebhookURL, got %s", cfg.WebhookURL)
	}
	if cfg.SDKUrl != "http://drone-rx-sdk:3000" {
		t.Errorf("expected default SDKUrl http://drone-rx-sdk:3000, got %s", cfg.SDKUrl)
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
		t.Errorf("expected NATS URL from env, env, got %s", cfg.NATSUrl)
	}
	if cfg.TickerInterval != 10 {
		t.Errorf("expected ticker interval 10, got %d", cfg.TickerInterval)
	}
	if cfg.WebhookURL != "https://example.com/hook" {
		t.Errorf("expected WebhookURL from env, got %s", cfg.WebhookURL)
	}
}
