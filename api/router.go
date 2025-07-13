package api

import (
	"RinhaBackend/models"
	"RinhaBackend/processor"
	"RinhaBackend/storage"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convertendo para o modelo interno
	payment := models.Payment{
		CorrelationID: req.CorrelationID,
		Amount:        req.Amount,
		RequestedAt:   time.Now().UTC(),
	}

	// Processando o pagamento de forma assíncrona
	go func() {
		if err := r.strategy.ProcessPayment(payment); err != nil {
			// Logando o erro, mas não bloqueando a resposta
			// pois o processamento é assíncrono
		} else {
			// Salvando o pagamento processado
			r.repo.SavePayment(payment)
		}
	}()

	// Retornando sucesso imediatamente
	c.JSON(http.StatusAccepted, models.PaymentResponse{
		Message: "Payment request accepted",
	})
}

func (r *Router) handleSummary(c *gin.Context) {
	// Processando parâmetros de data
	var fromTime, toTime *time.Time

	if fromStr := c.Query("from"); fromStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			fromTime = &parsedTime
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			toTime = &parsedTime
		}
	}

	// Obtendo o resumo
	summary, err := r.repo.GetSummary(fromTime, toTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
