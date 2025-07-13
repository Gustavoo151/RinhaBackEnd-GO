package api

import (
	"RinhaBackend/processor"
	"RinhaBackend/storage"
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
