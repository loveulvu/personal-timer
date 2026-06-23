package agent

import (
	"encoding/json"
	"time"
)

type AgentRunStatus string

const (
	AgentRunStatusPending              AgentRunStatus = "pending"
	AgentRunStatusRunning              AgentRunStatus = "running"
	AgentRunStatusCompleted            AgentRunStatus = "completed"
	AgentRunStatusFailed               AgentRunStatus = "failed"
	AgentRunStatusCancelled            AgentRunStatus = "cancelled"
	AgentRunStatusRequiresConfirmation AgentRunStatus = "requires_confirmation"
)

type AgentStepStatus string

const (
	AgentStepStatusPending   AgentStepStatus = "pending"
	AgentStepStatusRunning   AgentStepStatus = "running"
	AgentStepStatusCompleted AgentStepStatus = "completed"
	AgentStepStatusFailed    AgentStepStatus = "failed"
	AgentStepStatusSkipped   AgentStepStatus = "skipped"
)

type AgentStepType string

const (
	AgentStepTypeModelCall      AgentStepType = "model_call"
	AgentStepTypeBuildContext   AgentStepType = "build_context"
	AgentStepTypeModelDecision  AgentStepType = "model_decision"
	AgentStepTypeToolCall       AgentStepType = "tool_call"
	AgentStepTypeActionProposal AgentStepType = "action_proposal"
	AgentStepTypeFinalAnswer    AgentStepType = "final_answer"
	AgentStepTypeError          AgentStepType = "error"
)

type ActionProposalStatus string

const (
	ActionProposalStatusPending  ActionProposalStatus = "pending"
	ActionProposalStatusAccepted ActionProposalStatus = "accepted"
	ActionProposalStatusRejected ActionProposalStatus = "rejected"
	ActionProposalStatusExecuted ActionProposalStatus = "executed"
	ActionProposalStatusFailed   ActionProposalStatus = "failed"
	ActionProposalStatusExpired  ActionProposalStatus = "expired"
)

type AgentRun struct {
	ID             int64           `json:"id"`
	UserGoal       string          `json:"user_goal"`
	TargetDate     string          `json:"target_date"`
	Status         AgentRunStatus  `json:"status"`
	FinalAnswer    string          `json:"final_answer,omitempty"`
	PendingActions json.RawMessage `json:"pending_actions,omitempty"`
	ErrorMessage   string          `json:"error_message,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
}

type AgentRunListItem struct {
	ID                   int64          `json:"id"`
	UserGoal             string         `json:"user_goal"`
	TargetDate           string         `json:"target_date"`
	Status               AgentRunStatus `json:"status"`
	FinalAnswerExcerpt   string         `json:"final_answer_excerpt,omitempty"`
	ProposalCount        int            `json:"proposal_count"`
	PendingProposalCount int            `json:"pending_proposal_count"`
	StepCount            int            `json:"step_count"`
	CreatedAt            time.Time      `json:"created_at"`
	CompletedAt          *time.Time     `json:"completed_at,omitempty"`
}

type AgentStep struct {
	ID             int64           `json:"id"`
	RunID          int64           `json:"run_id"`
	StepIndex      int             `json:"step_index"`
	StepType       AgentStepType   `json:"step_type"`
	ToolName       string          `json:"tool_name,omitempty"`
	ToolInput      json.RawMessage `json:"tool_input,omitempty"`
	ToolOutput     json.RawMessage `json:"tool_output,omitempty"`
	ThoughtSummary string          `json:"thought_summary,omitempty"`
	Status         AgentStepStatus `json:"status"`
	ErrorMessage   string          `json:"error_message,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type AgentContextSnapshot struct {
	ID                  int64           `json:"id"`
	RunID               int64           `json:"run_id"`
	ContextPack         ContextPack     `json:"context_pack"`
	ContextJSON         json.RawMessage `json:"-"`
	TokenEstimate       int             `json:"token_estimate"`
	OmittedSections     []string        `json:"omitted_sections"`
	OmittedSectionsJSON json.RawMessage `json:"-"`
	CreatedAt           time.Time       `json:"created_at"`
}

type AgentTrajectory struct {
	Run             AgentRun              `json:"run"`
	ContextSnapshot *AgentContextSnapshot `json:"context_snapshot,omitempty"`
	Steps           []AgentStep           `json:"steps"`
	Proposals       []ActionProposal      `json:"proposals"`
}

type ActionProposal struct {
	ID               int64                `json:"id"`
	RunID            int64                `json:"run_id"`
	StepID           int64                `json:"step_id"`
	ToolName         string               `json:"tool_name"`
	ActionType       string               `json:"action_type"`
	Payload          json.RawMessage      `json:"payload,omitempty"`
	RiskLevel        ToolRiskLevel        `json:"risk_level"`
	Status           ActionProposalStatus `json:"status"`
	CreatedAt        time.Time            `json:"created_at"`
	DecidedAt        *time.Time           `json:"decided_at,omitempty"`
	ExecutedAt       *time.Time           `json:"executed_at,omitempty"`
	Result           json.RawMessage      `json:"result,omitempty"`
	ErrorMessage     string               `json:"error_message,omitempty"`
	TargetEntityType string               `json:"target_entity_type,omitempty"`
	TargetEntityID   *int64               `json:"target_entity_id,omitempty"`
}
