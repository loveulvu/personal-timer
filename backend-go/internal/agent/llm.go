package agent

import (
	"context"
	"encoding/json"
)

type ModelClient interface {
	Decide(ctx context.Context, input AgentDecisionInput) (AgentDecision, error)
}

type AgentDecisionInput struct {
	RunID          int64             `json:"run_id"`
	ContextPack    ContextPack       `json:"context_pack"`
	AvailableTools []AgentTool       `json:"available_tools"`
	Observations   []ToolObservation `json:"observations,omitempty"`
}

type ToolObservation struct {
	ToolName string          `json:"tool_name"`
	Output   json.RawMessage `json:"output,omitempty"`
}

type AgentDecision struct {
	Intent                string     `json:"intent"`
	ToolCalls             []ToolCall `json:"tool_calls,omitempty"`
	FinalAnswer           string     `json:"final_answer,omitempty"`
	NeedsUserConfirmation bool       `json:"needs_user_confirmation,omitempty"`
	ThoughtSummary        string     `json:"thought_summary,omitempty"`
	UnsupportedGoal       bool       `json:"unsupported_goal,omitempty"`
	ErrorMessage          string     `json:"error_message,omitempty"`
}

type DeterministicModelClient struct{}

func NewDeterministicModelClient() ModelClient {
	return DeterministicModelClient{}
}

func (DeterministicModelClient) Decide(ctx context.Context, input AgentDecisionInput) (AgentDecision, error) {
	return AgentDecision{
		Intent:         "final_answer",
		FinalAnswer:    "Agent run created a context pack. No live model client is configured yet.",
		ThoughtSummary: "deterministic fallback decision",
	}, nil
}
