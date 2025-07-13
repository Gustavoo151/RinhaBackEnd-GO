package models

type Payment struct {
	CorrelationID string  `json:"correlationID"`
	Amount        float64 `json:"amount"`
	RequestedAt   string  `json:"requestedAt"`
	ProcessedBy   string  `json:"processedBy"`
}
