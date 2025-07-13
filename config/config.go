package config

import (
	"os"
	"strconv"
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

func Load() *Config {
	port := getEnv("PORT", "9999")
	defaultURL := getEnv("PAYMENT_PROCESSOR_URL_DEFAULT", "http://payment-processor-default:8080")
	fallbackURL := getEnv("PAYMENT_PROCESSOR_URL_FALLBACK", "http://payment-processor-fallback:8080")

	healthCheckInterval, _ := strconv.Atoi(getEnv("HEALTH_CHECK_INTERVAL", "5"))
	httpTimeout, _ := strconv.Atoi(getEnv("HTTP_TIMEOUT", "2"))
	maxConcurrent, _ := strconv.Atoi(getEnv("MAX_CONCURRENT", "1000"))

	return &Config{
		Port:                 port,
		DefaultProcessorURL:  defaultURL,
		FallbackProcessorURL: fallbackURL,
		HealthCheckInterval:  time.Duration(healthCheckInterval) * time.Second,
		HTTPTimeout:          time.Duration(httpTimeout) * time.Second,
		MaxConcurrent:        maxConcurrent,
		DatabaseDSN:          getEnv("DATABASE_DSN", "postgres://postgres:postgres@db:5432/rinha?sslmode=disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
