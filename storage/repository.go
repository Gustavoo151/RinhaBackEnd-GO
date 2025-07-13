package storage

import (
	"RinhaBackend/config"
	"RinhaBackend/models"
	"database/sql"
	"log"
	"sync"
	"time"
)

type Repository struct {
	db         *sql.DB
	cache      map[string]models.Payment
	cacheMutex sync.RWMutex
}

func NewRepository(cfg *config.Config) *Repository {
	db, err := sql.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	// Configurando pool de conexões
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	repo := &Repository{
		db:    db,
		cache: make(map[string]models.Payment),
	}

	// Criando tabela se não existir
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			correlation_id UUID PRIMARY KEY,
			amount DECIMAL(15,2) NOT NULL,
			requested_at TIMESTAMP NOT NULL,
			processor VARCHAR(10) NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}
	return repo
}
