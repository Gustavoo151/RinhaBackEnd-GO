package processor

import "net/http"

type Client struct {
	baseURL    string
	name       string
	httpClient *http.Client
}
