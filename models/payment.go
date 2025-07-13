package models

import (
	_ "time"

	_ "github.com/google/uuid"
)

type Payment struct {
	CorrelationID string  `json:"correlationID"`
	Amount        float64 `json:"amount"`
	RequestedAt   string  `json:"requestedAt"`
	ProcessedBy   string  `json:"processedBy"`
}

type PaymentRequest struct {
	CorrelationID string  `json:"correlationID" binding:"required,uuid"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
}
