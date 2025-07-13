package health

import (
	"RinhaBackend/models"
	"RinhaBackend/processor"
	"sync"
	"time"
)

type Monitor struct {
	defaultClient  *processor.Client
	fallbackClient *processor.Client
	interval       time.Duration
	defaultStatus  models.HealthStatus
	fallbackStatus models.HealthStatus
	mu             sync.RWMutex
	stop           chan struct{}
}

func NewMonitor(defaultClient, fallbackClient *processor.Client, interval time.Duration) *Monitor {
	return &Monitor{
		defaultClient:  defaultClient,
		fallbackClient: fallbackClient,
		interval:       interval,
		defaultStatus:  models.HealthStatus{Failing: false, MinResponseTime: 100},
		fallbackStatus: models.HealthStatus{Failing: false, MinResponseTime: 100},
		stop:           make(chan struct{}),
	}
}

func (m *Monitor) Start() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Verificação inicial
	m.checkHealth()

	for {
		select {
		case <-ticker.C:
			m.checkHealth()
		case <-m.stop:
			return
		}
	}
}

func (m *Monitor) Stop() {
	close(m.stop)
}
