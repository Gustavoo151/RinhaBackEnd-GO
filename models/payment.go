package models

import (
	"time"
)

type Payment struct {
	CorrelationID string    `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
	ProcessedBy   string    `json:"-"` // default ou fallback
}

type PaymentRequest struct {
	CorrelationID string  `json:"correlationId" binding:"required,uuid4"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
}

type PaymentResponse struct {
	Message string `json:"message"`
}

type SummaryResponse struct {
	Default  ProcessorSummary `json:"default"`
	Fallback ProcessorSummary `json:"fallback"`
}

type ProcessorSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type HealthStatus struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}
