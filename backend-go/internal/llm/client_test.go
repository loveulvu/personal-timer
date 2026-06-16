package llm

import (
	"testing"
	"time"
)

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
