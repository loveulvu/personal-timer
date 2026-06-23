package agent

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
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

var minutesPattern = regexp.MustCompile(`(\d+)\s*分钟`)

func NewDeterministicModelClient() ModelClient {
	return DeterministicModelClient{}
}

func (DeterministicModelClient) Decide(ctx context.Context, input AgentDecisionInput) (AgentDecision, error) {
	if len(input.Observations) == 0 {
		if decision, ok := deterministicCreateTaskDecision(input); ok {
			return decision, nil
		}
	}
	return AgentDecision{
		Intent:         "final_answer",
		FinalAnswer:    "Agent run created a context pack. No live model client is configured yet.",
		ThoughtSummary: "deterministic fallback decision",
	}, nil
}

func deterministicCreateTaskDecision(input AgentDecisionInput) (AgentDecision, bool) {
	goal := input.ContextPack.UserGoal
	if !hasAny(goal, "创建", "新建", "安排") || !strings.Contains(goal, "任务") {
		return AgentDecision{}, false
	}
	if !toolAvailable(input.AvailableTools, "create_daily_task") {
		return AgentDecision{
			UnsupportedGoal: true,
			ErrorMessage:    "create_daily_task tool is unavailable",
			ThoughtSummary:  "deterministic fallback could not find write tool",
		}, true
	}
	projectID, ok := inferDemoProjectID(goal, input.ContextPack.TodayTasks)
	if !ok {
		return AgentDecision{
			UnsupportedGoal: true,
			ErrorMessage:    "cannot infer project_id for create_daily_task demo",
			ThoughtSummary:  "deterministic fallback refused to hardcode project_id",
		}, true
	}
	minutes := inferMinutes(goal)
	payload := map[string]any{
		"date":              input.ContextPack.TargetDate,
		"project_id":        projectID,
		"title":             inferDemoTitle(goal),
		"estimated_minutes": minutes,
	}
	data, _ := json.Marshal(payload)
	return AgentDecision{
		Intent:         "tool_call",
		ToolCalls:      []ToolCall{{ToolName: "create_daily_task", Input: data}},
		ThoughtSummary: "deterministic fallback demo create_daily_task proposal",
	}, true
}

func hasAny(value string, items ...string) bool {
	for _, item := range items {
		if strings.Contains(value, item) {
			return true
		}
	}
	return false
}

func toolAvailable(tools []AgentTool, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func inferDemoProjectID(goal string, tasks []ContextTask) (int64, bool) {
	wantPersonalTimer := strings.Contains(strings.ToLower(goal), "personal_study_timer")
	unique := map[int64]bool{}
	for _, task := range tasks {
		if task.ProjectID == nil || *task.ProjectID <= 0 {
			continue
		}
		if wantPersonalTimer && normalizeProjectName(task.ProjectName) == "personalstudytimer" {
			return *task.ProjectID, true
		}
		unique[*task.ProjectID] = true
	}
	if len(unique) == 1 {
		// ponytail: deterministic demo fallback; uses the only project already present in today's context, never a hardcoded id.
		for id := range unique {
			return id, true
		}
	}
	return 0, false
}

func normalizeProjectName(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func inferMinutes(goal string) int {
	match := minutesPattern.FindStringSubmatch(goal)
	if len(match) == 2 {
		if minutes, err := strconv.Atoi(match[1]); err == nil && minutes > 0 {
			return minutes
		}
	}
	return 60
}

func inferDemoTitle(goal string) string {
	if strings.Contains(goal, "Go GC") || strings.Contains(goal, "go gc") {
		return "Go GC 复习"
	}
	return "学习任务"
}
