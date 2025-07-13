package processor

import (
	"RinhaBackend/health"
	"sync"
)

type Strategy struct {
	defaultClient  *Client
	fallbackClient *Client
	healthMonitor  *health.Monitor
	workers        chan struct{}
	mu             sync.Mutex
}

func NewStrategy(defaultClient, fallbackClient *Client, healthMonitor *health.Monitor) *Strategy {
	return &Strategy{
		defaultClient:  defaultClient,
		fallbackClient: fallbackClient,
		healthMonitor:  healthMonitor,
		workers:        make(chan struct{}, 1000), // Limitar concorrência
	}
}
