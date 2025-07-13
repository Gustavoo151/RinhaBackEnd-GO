package config

import (
	"os"
	"time"
)

type Config struct {
	Port                 string
	DefaultProcessorURL  string
	FallbackProcessorURL string
	HealthCheckInterval  time.Duration
	HTTPTimeout          time.Duration
	MaxConcurrent        int
	DatabaseDSN          string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
