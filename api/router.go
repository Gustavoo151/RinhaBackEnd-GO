package api

import (
	"RinhaBackend/processor"
	"RinhaBackend/storage"
)

type Router struct {
	strategy *processor.Strategy
	repo     *storage.Repository
}
