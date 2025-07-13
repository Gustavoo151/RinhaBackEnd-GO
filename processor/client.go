package processor

import (
	"RinhaBackend/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	name       string
	httpClient *http.Client
}

func NewClient(baseURL, name string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		name:       name,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) ProcessPayment(ctx context.Context, payment models.Payment) error {
	paymentData, err := json.Marshal(payment)

	if err != nil {
		return fmt.Errorf("Error marshalling payment data: %v", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/payments", c.baseURL),
		bytes.NewBuffer(paymentData),
	)

	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("Error executing request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("resposta inválida do processador: %d", resp.StatusCode)
}
