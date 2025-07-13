package models

import (
	_ "github.com/google/uuid"

	_ "time"
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
