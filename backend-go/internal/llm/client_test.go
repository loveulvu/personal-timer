package llm

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestLLMTimeoutFromEnvDefaultsToNinetySeconds(t *testing.T) {
	t.Setenv("SUMMARY_LLM_TIMEOUT_SECONDS", "")
	t.Setenv("LLM_TIMEOUT_SECONDS", "")

	if got := llmTimeoutFromEnv(); got != 90*time.Second {
		t.Fatalf("timeout = %s, want 90s", got)
	}
}

func TestLLMTimeoutFromEnvUsesSummaryOverrideFirst(t *testing.T) {
	t.Setenv("SUMMARY_LLM_TIMEOUT_SECONDS", "120")
	t.Setenv("LLM_TIMEOUT_SECONDS", "45")

	if got := llmTimeoutFromEnv(); got != 120*time.Second {
		t.Fatalf("timeout = %s, want 120s", got)
	}
}

func TestLLMTimeoutFromEnvUsesLLMOverride(t *testing.T) {
	t.Setenv("SUMMARY_LLM_TIMEOUT_SECONDS", "")
	t.Setenv("LLM_TIMEOUT_SECONDS", "75")

	if got := llmTimeoutFromEnv(); got != 75*time.Second {
		t.Fatalf("timeout = %s, want 75s", got)
	}
}

func TestLLMMaxRetriesFromEnvDefaultsToTwo(t *testing.T) {
	t.Setenv("SUMMARY_LLM_MAX_RETRIES", "")
	t.Setenv("LLM_MAX_RETRIES", "")

	if got := llmMaxRetriesFromEnv(); got != 2 {
		t.Fatalf("retries = %d, want 2", got)
	}
}

func TestLLMMaxRetriesFromEnvUsesSummaryOverrideFirst(t *testing.T) {
	t.Setenv("SUMMARY_LLM_MAX_RETRIES", "4")
	t.Setenv("LLM_MAX_RETRIES", "1")

	if got := llmMaxRetriesFromEnv(); got != 4 {
		t.Fatalf("retries = %d, want 4", got)
	}
}

func TestGenerateSummaryRetriesNetworkErrors(t *testing.T) {
	attempts := 0
	client := testClientWithTransport(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return nil, &net.DNSError{Err: "temporary timeout", IsTimeout: true}
		}
		return successResponse("ok"), nil
	}), 2)

	got, err := client.GenerateSummary(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("GenerateSummary error = %v", err)
	}
	if got != "ok" {
		t.Fatalf("content = %q, want ok", got)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestGenerateSummaryRetriesRetryableHTTPStatuses(t *testing.T) {
	for _, status := range []int{429, 500, 502, 503, 504} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			attempts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				if attempts == 1 {
					http.Error(w, "try again", status)
					return
				}
				writeSuccess(w, "ok")
			}))
			defer server.Close()

			client := testClient(server.URL, 1)
			got, err := client.GenerateSummary(context.Background(), "prompt")
			if err != nil {
				t.Fatalf("GenerateSummary error = %v", err)
			}
			if got != "ok" {
				t.Fatalf("content = %q, want ok", got)
			}
			if attempts != 2 {
				t.Fatalf("attempts = %d, want 2", attempts)
			}
		})
	}
}

func TestGenerateSummaryDoesNotRetryNonRetryableHTTPStatuses(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			attempts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				http.Error(w, "bad request", status)
			}))
			defer server.Close()

			client := testClient(server.URL, 2)
			_, err := client.GenerateSummary(context.Background(), "prompt")
			if err == nil {
				t.Fatal("GenerateSummary error = nil, want error")
			}
			if attempts != 1 {
				t.Fatalf("attempts = %d, want 1", attempts)
			}
		})
	}
}

func TestGenerateSummaryDoesNotExceedConfiguredRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		http.Error(w, "still down", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := testClient(server.URL, 1)
	_, err := client.GenerateSummary(context.Background(), "prompt")
	if err == nil {
		t.Fatal("GenerateSummary error = nil, want error")
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestGenerateSummaryReturnsFinalSuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			http.Error(w, "busy", http.StatusBadGateway)
			return
		}
		writeSuccess(w, "eventual ok")
	}))
	defer server.Close()

	client := testClient(server.URL, 2)
	got, err := client.GenerateSummary(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("GenerateSummary error = %v", err)
	}
	if got != "eventual ok" {
		t.Fatalf("content = %q, want eventual ok", got)
	}
}

func TestGenerateSummaryReturnsLastErrorAfterRetries(t *testing.T) {
	statuses := []int{http.StatusInternalServerError, http.StatusServiceUnavailable}
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := statuses[attempts]
		attempts++
		http.Error(w, "last body", status)
	}))
	defer server.Close()

	client := testClient(server.URL, 1)
	_, err := client.GenerateSummary(context.Background(), "prompt")
	if err == nil {
		t.Fatal("GenerateSummary error = nil, want error")
	}
	if !strings.Contains(err.Error(), "status 503") || !strings.Contains(err.Error(), "last body") {
		t.Fatalf("error = %q, want final 503 body", err.Error())
	}
}

func TestGenerateSummaryDoesNotRetryJSONUnmarshalErrors(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{"))
	}))
	defer server.Close()

	client := testClient(server.URL, 2)
	_, err := client.GenerateSummary(context.Background(), "prompt")
	if err == nil {
		t.Fatal("GenerateSummary error = nil, want error")
	}
	if !errors.Is(err, ErrRequestFailed) {
		t.Fatalf("error = %v, want ErrRequestFailed", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func testClient(baseURL string, maxRetries int) *HTTPClient {
	return &HTTPClient{
		apiKey:     "test-key",
		baseURL:    baseURL,
		model:      "test-model",
		httpClient: http.DefaultClient,
		maxRetries: maxRetries,
		sleep:      func(time.Duration) {},
	}
}

func testClientWithTransport(transport http.RoundTripper, maxRetries int) *HTTPClient {
	return &HTTPClient{
		apiKey:     "test-key",
		baseURL:    "https://example.test",
		model:      "test-model",
		httpClient: &http.Client{Transport: transport},
		maxRetries: maxRetries,
		sleep:      func(time.Duration) {},
	}
}

func successResponse(content string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"role":"assistant","content":"` + content + `"}}]}`)),
		Header:     make(http.Header),
	}
}

func writeSuccess(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"` + content + `"}}]}`))
}
