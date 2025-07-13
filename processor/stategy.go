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
