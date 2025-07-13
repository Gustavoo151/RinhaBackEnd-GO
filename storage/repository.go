package storage

import (
	"RinhaBackend/models"
	"database/sql"
	"sync"
)

type Repository struct {
	db         *sql.DB
	cache      map[string]models.Payment
	cacheMutex sync.RWMutex
}
