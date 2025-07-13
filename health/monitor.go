package health

import (
	"RinhaBackend/models"
	"RinhaBackend/processor"
	"sync"
	"time"
)

type Monitor struct {
	defaultClient  *processor.Client
	fallbackClient *processor.Client
	interval       time.Duration
	defaultStatus  models.HealthStatus
	fallbackStatus models.HealthStatus
	mu             sync.RWMutex
	stop           chan struct{}
}
