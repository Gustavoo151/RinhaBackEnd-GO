package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	port   string
	router *Router
	server *http.Server
}

func NewServer(port string, router *Router) *Server {
	return &Server{
		port:   port,
		router: router,
	}
}

func (s *Server) Start() error {
	gin.SetMode(gin.ReleaseMode)

	// Configurando o servidor HTTP
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.port),
		Handler: s.router.SetupRoutes(),
	}

	// Iniciando o servidor
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
