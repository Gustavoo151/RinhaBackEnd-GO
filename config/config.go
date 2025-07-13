package config

import "time"

type Config struct {
	Port                 string
	DefaultProcessorURL  string
	FallbackProcessorURL string
	HealthCheckInterval  time.Duration
	HTTPTimeout          time.Duration
	MaxConcurrent        int
	DatabaseDSN          string
}
