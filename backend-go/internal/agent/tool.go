package agent

import "encoding/json"

type ToolRiskLevel string

const (
	ToolRiskLevelRead        ToolRiskLevel = "read"
	ToolRiskLevelWrite       ToolRiskLevel = "write"
	ToolRiskLevelDestructive ToolRiskLevel = "destructive"
)

type AgentTool struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	RiskLevel    ToolRiskLevel   `json:"risk_level"`
	InputSchema  json.RawMessage `json:"input_schema,omitempty"`
	OutputSchema json.RawMessage `json:"output_schema,omitempty"`
}

type ToolCall struct {
	ToolName string          `json:"tool_name"`
	Input    json.RawMessage `json:"input,omitempty"`
}

type ToolResult struct {
	Success      bool            `json:"success"`
	Output       json.RawMessage `json:"output,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	Proposal     *ActionProposal `json:"proposal,omitempty"`
}
