package config

import (
	"os"
)

// Config holds the client gateway service configuration
type Config struct {
	Port                string
	DatabaseURL         string
	CustomerOnboardURL  string
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8083"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://temporal:temporal@localhost:5432/postgres?sslmode=disable"),
		CustomerOnboardURL: getEnv("CUSTOMER_ONBOARD_URL", "http://localhost:8082"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
