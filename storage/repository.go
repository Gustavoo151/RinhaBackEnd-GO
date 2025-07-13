package storage

import (
	"RinhaBackend/config"
	"RinhaBackend/models"
	"database/sql"
	_ "gi
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

func (r *Repository) SavePayment(payment models.Payment) error {
	// Salvando no cache
	r.cacheMutex.Lock()
	r.cache[payment.CorrelationID] = payment
	r.cacheMutex.Unlock()

	// Salvando no banco de dados de forma assincrona
	go func() {
		_, err := r.db.Exec(
			"INSERT INTO payments (correlation_id, amount, requested_at, processor) VALUES ($1, $2, $3, $4)",
			payment.CorrelationID,
			payment.Amount,
			payment.RequestedAt,
			payment.ProcessedBy,
		)
		if err != nil {
			log.Printf("Erro ao salvar pagamento no banco: %v", err)
		}
	}()

	return nil
}

func (r *Repository) GetSummary(from, to *time.Time) (models.SummaryResponse, error) {
	var summary models.SummaryResponse

	// Definindo o período
	var fromTime, toTime time.Time
	if from != nil {
		fromTime = *from
	} else {
		fromTime = time.Time{}
	}

	if to != nil {
		toTime = *to
	} else {
		toTime = time.Now().UTC()
	}

	// Consultando o banco de dados
	rows, err := r.db.Query(`
		SELECT processor, COUNT(*) as total_requests, COALESCE(SUM(amount), 0) as total_amount
		FROM payments
		WHERE ($1 IS NULL OR requested_at >= $1)
		AND ($2 IS NULL OR requested_at <= $2)
		GROUP BY processor
	`, fromTime, toTime)

	if err != nil {
		return summary, err
	}
	defer rows.Close()

	// Inicializando os valores
	summary.Default = models.ProcessorSummary{TotalRequests: 0, TotalAmount: 0}
	summary.Fallback = models.ProcessorSummary{TotalRequests: 0, TotalAmount: 0}

	// Processando os resultados
	for rows.Next() {
		var processor string
		var totalRequests int
		var totalAmount float64

		if err := rows.Scan(&processor, &totalRequests, &totalAmount); err != nil {
			return summary, err
		}

		if processor == "default" {
			summary.Default.TotalRequests = totalRequests
			summary.Default.TotalAmount = totalAmount
		} else if processor == "fallback" {
			summary.Fallback.TotalRequests = totalRequests
			summary.Fallback.TotalAmount = totalAmount
		}
	}

	return summary, nil
}
