package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration loaded from environment variables
type Config struct {
	Port              string
	DatabaseURL       string
	TemporalHostPort  string
	TemporalNamespace string
	TemporalTaskQueue string
	MinEligibilityAge int
	RunMode           string // "all", "http", "worker"
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8081"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://temporal:temporal@localhost:5432/postgres?sslmode=disable"),
		TemporalHostPort:  getEnv("TEMPORAL_HOST_PORT", "localhost:7233"),
		TemporalNamespace: getEnv("TEMPORAL_NAMESPACE", "default"),
		TemporalTaskQueue: getEnv("TEMPORAL_TASK_QUEUE", "CUSTOMER_ONBOARDING_TASK_QUEUE"),
		MinEligibilityAge: getEnvAsInt("MIN_ELIGIBILITY_AGE", 18),
		RunMode:           getEnv("RUN_MODE", "all"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}
