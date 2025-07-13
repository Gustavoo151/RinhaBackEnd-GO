package processor

import (
	"context"
	"log"
	"sync"
	"time"

	"RinhaBackend/models"
)

type Strategy struct {
	defaultClient  *Client
	fallbackClient *Client
	healthMonitor  HealthChecker
	workers        chan struct{}
	mu             sync.Mutex
}

type HealthChecker interface {
	GetDefaultStatus() models.HealthStatus
	GetFallbackStatus() models.HealthStatus
}

func NewStrategy(defaultClient, fallbackClient *Client, healthMonitor HealthChecker) *Strategy {
	return &Strategy{
		defaultClient:  defaultClient,
		fallbackClient: fallbackClient,
		healthMonitor:  healthMonitor,
		workers:        make(chan struct{}, 1000), // Limitar concorrência
	}
}

func (s *Strategy) ProcessPayment(payment models.Payment) error {
	// Adicionando timestamp
	if payment.RequestedAt.IsZero() {
		payment.RequestedAt = time.Now()
	}

	// Obtendo status de saúde dos processadores
	defaultStatus := s.healthMonitor.GetDefaultStatus()
	fallbackStatus := s.healthMonitor.GetFallbackStatus()

	// Determinando o melhor processador
	var client *Client
	if !defaultStatus.Failing {
		client = s.defaultClient
		payment.ProcessedBy = "default"
	} else if !fallbackStatus.Failing {
		client = s.fallbackClient
		payment.ProcessedBy = "fallback"
	} else {
		// Ambos processadores estão falhando, simular processamento
		log.Printf("Ambos processadores indisponíveis, simulando processamento do pagamento %s", payment.CorrelationID)
		payment.ProcessedBy = "simulated"
		return nil
	}

	// Criando contexto com timeout adequado
	timeout := 5 * time.Second
	if !defaultStatus.Failing && defaultStatus.MinResponseTime > 0 {
		timeout = time.Duration(defaultStatus.MinResponseTime*3) * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Processando o pagamento com limite de concorrência
	s.workers <- struct{}{}
	defer func() { <-s.workers }()

	err := client.ProcessPayment(ctx, payment)
	if err != nil {
		log.Printf("Erro ao processar pagamento %s com %s: %v", payment.CorrelationID, client.GetName(), err)
		return err
	}

	log.Printf("Pagamento %s processado com sucesso usando %s", payment.CorrelationID, client.GetName())
	return nil
}

func (s *Strategy) ProcessPaymentAsync(payment models.Payment) {
	go func() {
		if err := s.ProcessPayment(payment); err != nil {
			log.Printf("Erro no processamento assíncrono: %v", err)
		}
	}()
}
