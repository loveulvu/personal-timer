package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTestLLMSummaryUsesLongPromptAndNoSummaryStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	llmCalls := 0
	llmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		llmCalls++
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s, want /chat/completions", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"summary test ok"}}]}`))
	}))
	defer llmServer.Close()

	t.Setenv("LLM_API_KEY", "test-key")
	t.Setenv("LLM_BASE_URL", llmServer.URL)
	t.Setenv("LLM_MODEL", "test-model")

	r := gin.New()
	r.POST("/api/llm/test-summary", TestLLMSummary)

	req := httptest.NewRequest(http.MethodPost, "/api/llm/test-summary", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	if llmCalls != 1 {
		t.Fatalf("llm calls = %d, want 1", llmCalls)
	}

	var body struct {
		Status          string `json:"status"`
		PromptChars     int    `json:"prompt_chars"`
		ResponsePreview string `json:"response_preview"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("status = %q, want ok", body.Status)
	}
	if body.PromptChars < 2500 || body.PromptChars > 4000 {
		t.Fatalf("prompt_chars = %d, want 2500..4000", body.PromptChars)
	}
	if body.ResponsePreview != "summary test ok" {
		t.Fatalf("response_preview = %q, want summary test ok", body.ResponsePreview)
	}
}
