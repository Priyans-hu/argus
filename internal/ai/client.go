package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with a local Ollama instance
type Client struct {
	endpoint   string
	model      string
	httpClient *http.Client
}

// generateRequest is the Ollama /api/generate request body
type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// generateResponse is the Ollama /api/generate response body
type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewClient creates a new Ollama API client
func NewClient(endpoint, model string, timeout time.Duration) *Client {
	return &Client{
		endpoint: endpoint,
		model:    model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// IsAvailable checks if the Ollama server is reachable and the model exists
func (c *Client) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

// Generate sends a prompt to Ollama and returns the response text
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := generateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/api/generate", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var genResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return genResp.Response, nil
}
