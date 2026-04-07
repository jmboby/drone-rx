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
		TickerInterval: getEnvInt("TICKER_INTERVAL", 10),
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
