package storage

import (
	"RinhaBackend/config"
	"RinhaBackend/models"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type Repository struct {
	db         *sql.DB
	cache      map[string]models.Payment
	cacheMutex sync.RWMutex
}

func NewRepository(cfg *config.Config) *Repository {
	var db *sql.DB
	var err error

	// Tentativas de conexão com retry
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", cfg.DatabaseDSN)
		if err != nil {
			log.Printf("Tentativa %d/%d: Erro ao abrir conexão com banco: %v", i+1, maxRetries, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// Testando a conexão
		err = db.Ping()
		if err != nil {
			log.Printf("Tentativa %d/%d: Erro ao conectar com banco: %v", i+1, maxRetries, err)
			db.Close()
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		log.Println("Conexão com banco estabelecida com sucesso")
		break
	}

	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco após %d tentativas: %v", maxRetries, err)
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

	// Criando índices
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payments_requested_at ON payments(requested_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_payments_processor ON payments(processor)")

	log.Println("Repositório inicializado com sucesso")
	return repo
}

func (r *Repository) SavePayment(payment models.Payment) error {
	// Definindo processador padrão se não estiver definido
	if payment.ProcessedBy == "" {
		payment.ProcessedBy = "simulated"
	}

	// Salvando no cache
	r.cacheMutex.Lock()
	r.cache[payment.CorrelationID] = payment
	r.cacheMutex.Unlock()

	// Salvando no banco de dados de forma síncrona para garantir consistência
	_, err := r.db.Exec(
		"INSERT INTO payments (correlation_id, amount, requested_at, processor) VALUES ($1, $2, $3, $4) ON CONFLICT (correlation_id) DO NOTHING",
		payment.CorrelationID, payment.Amount, payment.RequestedAt, payment.ProcessedBy,
	)
	if err != nil {
		log.Printf("Erro ao salvar pagamento no banco: %v", err)
		return fmt.Errorf("failed to save payment: %v", err)
	}

	log.Printf("Pagamento %s salvo com processador %s", payment.CorrelationID, payment.ProcessedBy)
	return nil
}

func (r *Repository) GetSummary(from, to *time.Time) (models.SummaryResponse, error) {
	// Definindo o período padrão se não fornecido
	var fromTime, toTime time.Time
	if from == nil {
		fromTime = time.Now().Add(-24 * time.Hour)
	} else {
		fromTime = *from
	}

	if to == nil {
		toTime = time.Now()
	} else {
		toTime = *to
	}

	log.Printf("Buscando resumo de %v até %v", fromTime, toTime)

	// Consultando o banco de dados com tipos explícitos
	rows, err := r.db.Query(`
        SELECT processor, COUNT(*) as total_requests, COALESCE(SUM(amount), 0) as total_amount
        FROM payments 
        WHERE requested_at BETWEEN $1 AND $2
        GROUP BY processor
    `, fromTime, toTime)
	if err != nil {
		log.Printf("Erro ao consultar banco: %v", err)
		return models.SummaryResponse{}, fmt.Errorf("erro ao consultar banco: %v", err)
	}
	defer rows.Close()

	// Inicializando os valores
	summary := models.SummaryResponse{
		Default:  models.ProcessorSummary{TotalRequests: 0, TotalAmount: 0},
		Fallback: models.ProcessorSummary{TotalRequests: 0, TotalAmount: 0},
	}

	// Processando os resultados
	for rows.Next() {
		var processor string
		var totalRequests int
		var totalAmount float64

		if err := rows.Scan(&processor, &totalRequests, &totalAmount); err != nil {
			log.Printf("Erro ao processar linha: %v", err)
			continue
		}

		log.Printf("Processador %s: %d requests, total %.2f", processor, totalRequests, totalAmount)

		switch processor {
		case "default":
			summary.Default.TotalRequests = totalRequests
			summary.Default.TotalAmount = totalAmount
		case "fallback":
			summary.Fallback.TotalRequests = totalRequests
			summary.Fallback.TotalAmount = totalAmount
		case "simulated":
			// Para testes, vamos adicionar aos valores do default
			summary.Default.TotalRequests += totalRequests
			summary.Default.TotalAmount += totalAmount
		}
	}

	log.Printf("Resumo final: Default(%d, %.2f), Fallback(%d, %.2f)",
		summary.Default.TotalRequests, summary.Default.TotalAmount,
		summary.Fallback.TotalRequests, summary.Fallback.TotalAmount)

	return summary, nil
}
