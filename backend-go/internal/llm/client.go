package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	maxRetries int
	sleep      func(time.Duration)
}

type llmTransportError struct {
	err     error
	timeout bool
}

func (e llmTransportError) Error() string {
	if e.timeout {
		return fmt.Sprintf("%s: %v", ErrRequestTimeout, e.err)
	}
	return fmt.Sprintf("%s: %v", ErrRequestFailed, e.err)
}

func (e llmTransportError) Unwrap() []error {
	if e.timeout {
		return []error{ErrRequestTimeout, e.err}
	}
	return []error{ErrRequestFailed, e.err}
}

type llmStatusError struct {
	statusCode int
	body       string
}

func (e llmStatusError) Error() string {
	return fmt.Sprintf("%s with status %d: %s", ErrRequestFailed, e.statusCode, e.body)
}

func (e llmStatusError) Unwrap() error {
	return ErrRequestFailed
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
		maxRetries: llmMaxRetriesFromEnv(),
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

	maxAttempts := c.maxRetries + 1
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		start := time.Now()
		content, err := c.generateSummaryOnce(ctx, endpoint, body)
		if err == nil {
			return content, nil
		}

		lastErr = err
		retrying := attempt < maxAttempts && isRetryableLLMError(err)
		sleepBeforeNextRetry := time.Duration(0)
		if retrying {
			sleepBeforeNextRetry = retryBackoff(attempt)
		}
		log.Printf(
			"LLM request attempt failed: attempt=%d max_attempts=%d elapsed=%s error=%v retrying=%t sleep_before_next_retry=%s",
			attempt,
			maxAttempts,
			time.Since(start),
			err,
			retrying,
			sleepBeforeNextRetry,
		)
		if !retrying {
			return "", err
		}
		if err := c.sleepBeforeRetry(ctx, sleepBeforeNextRetry); err != nil {
			return "", err
		}
	}
	return "", lastErr
}

func (c *HTTPClient) generateSummaryOnce(ctx context.Context, endpoint string, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", llmTransportError{err: err, timeout: isTimeout(err)}
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", llmTransportError{err: err, timeout: isTimeout(err)}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", llmStatusError{
			statusCode: resp.StatusCode,
			body:       truncateForLog(string(responseBody), 1000),
		}
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

func (c *HTTPClient) sleepBeforeRetry(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	if c.sleep != nil {
		c.sleep(d)
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return llmTransportError{err: ctx.Err(), timeout: isTimeout(ctx.Err())}
	case <-timer.C:
		return nil
	}
}

func retryBackoff(attempt int) time.Duration {
	if attempt == 1 {
		return 500 * time.Millisecond
	}
	return 1500 * time.Millisecond
}

func isRetryableLLMError(err error) bool {
	var statusErr llmStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.statusCode {
		case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		default:
			return false
		}
	}

	var transportErr llmTransportError
	if errors.As(err, &transportErr) {
		if isTimeout(transportErr.err) {
			return true
		}
		return containsRetryableNetworkText(transportErr.err.Error())
	}

	return false
}

func containsRetryableNetworkText(value string) bool {
	value = strings.ToLower(value)
	for _, token := range []string{
		"wsarecv",
		"connection reset",
		"connection refused",
		"eof",
		"tls handshake timeout",
	} {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
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

func llmMaxRetriesFromEnv() int {
	for _, key := range []string{"SUMMARY_LLM_MAX_RETRIES", "LLM_MAX_RETRIES"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		retries, err := strconv.Atoi(value)
		if err == nil && retries >= 0 {
			return retries
		}
	}
	return 2
}

func truncateForLog(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "...(truncated)"
}
