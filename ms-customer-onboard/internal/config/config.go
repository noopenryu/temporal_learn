package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration loaded from environment variables
type Config struct {
	Port                 string
	DatabaseURL          string
	TemporalHostPort     string
	TemporalNamespace    string
	TemporalTaskQueue    string // Queue where ms-customer-onboard listens for workflows & local activities
	EligibilityTaskQueue string // Queue where ms-eligiblity listens for EvaluateEligibilityActivity
	KycTaskQueue         string // Queue where ms-kyc listens for KYC checks
	DraftTTLDays         int
	RunMode              string // "all", "http", "worker"
}

// Load loads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		Port:                 getEnv("PORT", "8082"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://temporal:temporal@localhost:5432/postgres?sslmode=disable"),
		TemporalHostPort:     getEnv("TEMPORAL_HOST_PORT", "localhost:7233"),
		TemporalNamespace:    getEnv("TEMPORAL_NAMESPACE", "default"),
		TemporalTaskQueue:    getEnv("TEMPORAL_TASK_QUEUE", "ONBOARDING_WORKFLOW_TASK_QUEUE"),
		EligibilityTaskQueue: getEnv("ELIGIBILITY_TASK_QUEUE", "CUSTOMER_ONBOARDING_TASK_QUEUE"),
		KycTaskQueue:         getEnv("KYC_TASK_QUEUE", "KYC_ONBOARDING_TASK_QUEUE"),
		DraftTTLDays:         getEnvAsInt("DRAFT_TTL_DAYS", 14),
		RunMode:              getEnv("RUN_MODE", "all"),
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
