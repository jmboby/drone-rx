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
	SDKUrl         string
	Namespace      string
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		NATSUrl:        getEnv("NATS_URL", "nats://localhost:4222"),
		TickerInterval: getEnvInt("TICKER_INTERVAL", 5),
		WebhookURL:     getEnv("WEBHOOK_URL", ""),
		SDKUrl:         getEnv("REPLICATED_SDK_URL", "http://drone-rx-sdk:3000"),
		Namespace:      getEnv("POD_NAMESPACE", "default"),
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
