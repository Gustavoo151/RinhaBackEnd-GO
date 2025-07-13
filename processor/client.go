package processor

import (
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
