package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"RinhaBackend/api"
	"RinhaBackend/config"
	"RinhaBackend/health"
	"RinhaBackend/processor"
	"RinhaBackend/storage"
)

func main() {
	// Carregando configurações
	cfg := config.Load()

	// Inicializando o repositório
	repo := storage.NewRepository(cfg)

	// Inicializando o cliente HTTP para os processadores
	defaultClient := processor.NewClient(cfg.DefaultProcessorURL, "default", cfg.HTTPTimeout)
	fallbackClient := processor.NewClient(cfg.FallbackProcessorURL, "fallback", cfg.HTTPTimeout)

	// Inicializando o monitor de saúde
	healthMonitor := health.NewMonitor(
		defaultClient,
		fallbackClient,
		cfg.HealthCheckInterval,
	)
	go healthMonitor.Start()

	// Inicializando a estratégia de processamento
	strategy := processor.NewStrategy(
		defaultClient,
		fallbackClient,
		healthMonitor,
	)

	// Inicializando o roteador HTTP
	router := api.NewRouter(strategy, repo)

	// Inicializando o servidor HTTP
	server := api.NewServer(cfg.Port, router)

	// Configurando graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Iniciando o servidor HTTP
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	// Aguardando sinal de encerramento
	<-ctx.Done()

	// Parando o servidor HTTP
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Erro ao encerrar o servidor: %v", err)
	}

	// Parando o monitor de saúde
	healthMonitor.Stop()

	log.Println("Servidor encerrado com sucesso")
}
