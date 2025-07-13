package api

import (
	"RinhaBackend/models"
	"RinhaBackend/processor"
	"RinhaBackend/storage"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Router struct {
	strategy *processor.Strategy
	repo     *storage.Repository
}

func NewRouter(strategy *processor.Strategy, repo *storage.Repository) *Router {
	return &Router{
		strategy: strategy,
		repo:     repo,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middlewares
	router.Use(gin.Recovery())

	// Configurando as rotas
	router.POST("/payments", r.handlePayment)
	router.GET("/payments-summary", r.handleSummary)

	return router
}

func (r *Router) handlePayment(c *gin.Context) {
	var req models.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Convertendo para o modelo interno
	payment := models.Payment{
		CorrelationID: req.CorrelationID,
		Amount:        req.Amount,
		RequestedAt:   time.Now(),
	}

	// Salvando no repositório
	if err := r.repo.SavePayment(payment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save payment"})
		return
	}

	// Processando o pagamento de forma assíncrona
	r.strategy.ProcessPaymentAsync(payment)

	// Retornando sucesso imediatamente
	c.JSON(http.StatusCreated, models.PaymentResponse{
		Message: "Payment processing initiated",
	})
}

func (r *Router) handleSummary(c *gin.Context) {
	// Processando parâmetros de data
	var from, to *time.Time

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = &parsed
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = &parsed
		}
	}

	// Obtendo o resumo
	summary, err := r.repo.GetSummary(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}
