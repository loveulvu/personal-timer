package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrNotConfigured = errors.New("LLM is not configured")
	ErrEmptyResponse = errors.New("LLM returned empty content")
)

type Client interface {
	GenerateSummary(ctx context.Context, prompt string) (string, error)
}

type HTTPClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func NewClientFromEnv() *HTTPClient {
	return &HTTPClient{
		apiKey:  strings.TrimSpace(os.Getenv("LLM_API_KEY")),
		baseURL: strings.TrimSpace(os.Getenv("LLM_BASE_URL")),
		model:   strings.TrimSpace(os.Getenv("LLM_MODEL")),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HTTPClient) GenerateSummary(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" || c.baseURL == "" || c.model == "" {
		return "", ErrNotConfigured
	}

	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "Write a concise, rational study review. Do not praise or use motivational language.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})
	if err != nil {
		return "", err
	}

	endpoint := strings.TrimRight(c.baseURL, "/")
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint += "/chat/completions"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("LLM request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var result chatResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", ErrEmptyResponse
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	if content == "" {
		return "", ErrEmptyResponse
	}
	return content, nil
}
