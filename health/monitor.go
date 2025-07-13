package health

import (
	"RinhaBackend/models"
	"RinhaBackend/processor"
	"context"
	"log"
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

func (m *Monitor) checkHealth() {
	// Usando context com timeout para a verificação de saúde
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Verificando o status do processador default
	defaultStatus, err := m.defaultClient.CheckHealth(ctx)
	if err != nil {
		log.Printf("Erro ao verificar saúde do processador default: %v", err)
		m.setDefaultStatus(models.HealthStatus{Failing: true, MinResponseTime: 9999})
	} else {
		m.setDefaultStatus(defaultStatus)
	}

	// Esperando um segundo para evitar rate limiting
	time.Sleep(1 * time.Second)

	// Verificando o status do processador fallback
	fallbackStatus, err := m.fallbackClient.CheckHealth(ctx)
	if err != nil {
		log.Printf("Erro ao verificar saúde do processador fallback: %v", err)
		m.setFallbackStatus(models.HealthStatus{Failing: true, MinResponseTime: 9999})
	} else {
		m.setFallbackStatus(fallbackStatus)
	}
}

func (m *Monitor) GetDefaultStatus() models.HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultStatus
}
