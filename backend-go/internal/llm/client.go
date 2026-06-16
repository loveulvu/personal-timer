package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNotConfigured  = errors.New("LLM is not configured")
	ErrMissingAPIKey  = errors.New("LLM_API_KEY is required")
	ErrMissingBaseURL = errors.New("LLM_BASE_URL is required")
	ErrMissingModel   = errors.New("LLM_MODEL is required")
	ErrRequestFailed  = errors.New("LLM request failed")
	ErrRequestTimeout = errors.New("LLM request timed out")
	ErrEmptyResponse  = errors.New("LLM returned empty content")
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
			Timeout: llmTimeoutFromEnv(),
		},
	}
}

func (c *HTTPClient) GenerateSummary(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("%w: %w", ErrNotConfigured, ErrMissingAPIKey)
	}
	if c.baseURL == "" {
		return "", fmt.Errorf("%w: %w", ErrNotConfigured, ErrMissingBaseURL)
	}
	if c.model == "" {
		return "", fmt.Errorf("%w: %w", ErrNotConfigured, ErrMissingModel)
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
		if isTimeout(err) {
			return "", fmt.Errorf("%w: %v", ErrRequestTimeout, err)
		}
		return "", fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf(
			"%w with status %d: %s",
			ErrRequestFailed,
			resp.StatusCode,
			truncateForLog(string(responseBody), 1000),
		)
	}

	var result chatResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", fmt.Errorf("%w: %v", ErrRequestFailed, err)
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

func isTimeout(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func llmTimeoutFromEnv() time.Duration {
	for _, key := range []string{"SUMMARY_LLM_TIMEOUT_SECONDS", "LLM_TIMEOUT_SECONDS"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		seconds, err := strconv.Atoi(value)
		if err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return 90 * time.Second
}

func truncateForLog(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "...(truncated)"
}
