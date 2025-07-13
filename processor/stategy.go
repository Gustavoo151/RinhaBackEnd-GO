package processor

import (
	"context"
	_ "errors"
	"log"
	"sync"
	"time"

	"RinhaBackend/health"
	"RinhaBackend/models"
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

func (s *Strategy) ProcessPayment(payment models.Payment) error {
	// Adicionando timestamp
	if payment.RequestedAt.IsZero() {
		payment.RequestedAt = time.Now().UTC()
	}

	// Obtendo status de saúde dos processadores
	defaultStatus := s.healthMonitor.GetDefaultStatus()
	fallbackStatus := s.healthMonitor.GetFallbackStatus()

	// Determinando o melhor processador
	var client *Client
	if !defaultStatus.Failing && (fallbackStatus.Failing || defaultStatus.MinResponseTime <= fallbackStatus.MinResponseTime) {
		client = s.defaultClient
		payment.ProcessedBy = "default"
	} else if !fallbackStatus.Failing {
		client = s.fallbackClient
		payment.ProcessedBy = "fallback"
	} else {
		// Ambos estão falhando, tenta o default mesmo assim
		client = s.defaultClient
		payment.ProcessedBy = "default"
	}

	// Criando contexto com timeout adequado
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Processando o pagamento com limite de concorrência
	select {
	case s.workers <- struct{}{}:
		defer func() { <-s.workers }()
		err := client.ProcessPayment(ctx, payment)
		if err != nil {
			// Se falhar e for o default, tenta o fallback
			if client == s.defaultClient && !fallbackStatus.Failing {
				log.Printf("Erro ao processar pagamento no default, tentando fallback: %v", err)
				client = s.fallbackClient
				payment.ProcessedBy = "fallback"
				return client.ProcessPayment(ctx, payment)
			}
			return err
		}
		return nil
	case <-time.After(100 * time.Millisecond):
		// Timeout de concorrência, usar abordagem síncrona
		return client.ProcessPayment(ctx, payment)
	}
}

func (s *Strategy) ProcessPaymentAsync(payment models.Payment) {
	go func() {
		err := s.ProcessPayment(payment)
		if err != nil {
			log.Printf("Erro ao processar pagamento assíncrono: %v", err)
		}
	}()
}
