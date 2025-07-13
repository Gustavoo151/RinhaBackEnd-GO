package health

import (
	"context"
	"log"
	"sync"
	"time"

	"RinhaBackend/models"
	"RinhaBackend/processor"
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
		defaultStatus:  models.HealthStatus{Failing: true, MinResponseTime: 0},
		fallbackStatus: models.HealthStatus{Failing: true, MinResponseTime: 0},
		stop:           make(chan struct{}),
	}
}

func (m *Monitor) Start() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Verificação inicial
	go m.checkHealth()

	for {
		select {
		case <-ticker.C:
			go m.checkHealth()
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
	start := time.Now()
	defaultStatus, err := m.defaultClient.CheckHealth(ctx)
	if err != nil {
		log.Printf("Erro ao verificar saúde do processador default: %v", err)
		m.setDefaultStatus(models.HealthStatus{Failing: true, MinResponseTime: 0})
	} else {
		responseTime := int(time.Since(start).Milliseconds())
		defaultStatus.MinResponseTime = responseTime
		m.setDefaultStatus(defaultStatus)
		log.Printf("Processador default saudável - Tempo de resposta: %dms", responseTime)
	}

	// Esperando um segundo para evitar rate limiting
	time.Sleep(1 * time.Second)

	// Verificando o status do processador fallback
	start = time.Now()
	fallbackStatus, err := m.fallbackClient.CheckHealth(ctx)
	if err != nil {
		log.Printf("Erro ao verificar saúde do processador fallback: %v", err)
		m.setFallbackStatus(models.HealthStatus{Failing: true, MinResponseTime: 0})
	} else {
		responseTime := int(time.Since(start).Milliseconds())
		fallbackStatus.MinResponseTime = responseTime
		m.setFallbackStatus(fallbackStatus)
		log.Printf("Processador fallback saudável - Tempo de resposta: %dms", responseTime)
	}
}

func (m *Monitor) GetDefaultStatus() models.HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultStatus
}

func (m *Monitor) GetFallbackStatus() models.HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fallbackStatus
}

func (m *Monitor) setDefaultStatus(status models.HealthStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultStatus = status
}

func (m *Monitor) setFallbackStatus(status models.HealthStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallbackStatus = status
}
