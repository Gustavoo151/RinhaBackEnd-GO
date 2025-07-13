package api

import (
	"RinhaBackend/processor"
	"RinhaBackend/storage"
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
