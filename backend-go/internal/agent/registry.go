package agent

import (
	"context"
	"encoding/json"
	"errors"
	"personal/internal/dailytasks"
	"personal/internal/memories"
	"personal/internal/plans"
	"sort"
	"strings"
	"time"
)

var (
	ErrToolNotFound     = errors.New("agent tool not found")
	ErrInvalidToolInput = errors.New("invalid agent tool input")
)

type dailyTaskLister interface {
	ListDailyTasksByDate(date string) ([]dailytasks.DailyTask, error)
}

type planRiskGetter interface {
	GetPlanRisk(ctx context.Context, date string) (*plans.PlanRiskResponse, error)
}

type memoryRecaller interface {
	RecallRelevantMemories(ctx context.Context, input memories.RecallInput) ([]memories.StudyMemory, error)
}

type ToolRegistry struct {
	tools map[string]AgentTool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: map[string]AgentTool{}}
}

func NewDefaultToolRegistry(tasks dailyTaskLister, planRisk planRiskGetter, recall memoryRecaller) *ToolRegistry {
	r := NewToolRegistry()
	r.Register(listTodayTasksTool(tasks))
	r.Register(evaluatePlanRiskTool(planRisk))
	r.Register(recallMemoriesTool(recall))
	r.Register(writeProposalTool("create_daily_task", "Create a daily task proposal."))
	r.Register(writeProposalTool("finish_task", "Finish a task proposal."))
	return r
}

func (r *ToolRegistry) Register(tool AgentTool) {
	if r.tools == nil {
		r.tools = map[string]AgentTool{}
	}
	r.tools[tool.Name] = tool
}

func (r *ToolRegistry) ListTools() []AgentTool {
	tools := make([]AgentTool, 0, len(r.tools))
	for _, tool := range r.tools {
		tool.Execute = nil
		tools = append(tools, tool)
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})
	return tools
}

func (r *ToolRegistry) Call(ctx context.Context, call ToolCall) (ToolResult, error) {
	tool, ok := r.tools[strings.TrimSpace(call.ToolName)]
	if !ok {
		return ToolResult{}, ErrToolNotFound
	}
	input := call.Input
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}
	if !json.Valid(input) {
		return ToolResult{}, ErrInvalidToolInput
	}
	if tool.Execute == nil {
		return ToolResult{}, errors.New("agent tool has no executor")
	}
	return tool.Execute(ctx, input)
}

func listTodayTasksTool(service dailyTaskLister) AgentTool {
	return AgentTool{
		Name:         "list_today_tasks",
		Description:  "List daily tasks for a date.",
		RiskLevel:    ToolRiskLevelRead,
		InputSchema:  rawJSON(`{"type":"object","required":["date"],"properties":{"date":{"type":"string","format":"date"}}}`),
		OutputSchema: rawJSON(`{"type":"array"}`),
		Execute: func(ctx context.Context, input json.RawMessage) (ToolResult, error) {
			var req struct {
				Date string `json:"date"`
			}
			if err := decodeToolInput(input, &req); err != nil {
				return ToolResult{}, err
			}
			if err := validateDate(req.Date); err != nil {
				return ToolResult{}, err
			}
			tasks, err := service.ListDailyTasksByDate(req.Date)
			if err != nil {
				return ToolResult{}, err
			}
			return rawOutput(tasks)
		},
	}
}

func evaluatePlanRiskTool(service planRiskGetter) AgentTool {
	return AgentTool{
		Name:         "evaluate_plan_risk",
		Description:  "Evaluate plan load risk for a date.",
		RiskLevel:    ToolRiskLevelRead,
		InputSchema:  rawJSON(`{"type":"object","required":["date"],"properties":{"date":{"type":"string","format":"date"}}}`),
		OutputSchema: rawJSON(`{"type":"object"}`),
		Execute: func(ctx context.Context, input json.RawMessage) (ToolResult, error) {
			var req struct {
				Date string `json:"date"`
			}
			if err := decodeToolInput(input, &req); err != nil {
				return ToolResult{}, err
			}
			if err := validateDate(req.Date); err != nil {
				return ToolResult{}, err
			}
			risk, err := service.GetPlanRisk(ctx, req.Date)
			if err != nil {
				return ToolResult{}, err
			}
			return rawOutput(risk)
		},
	}
}

func recallMemoriesTool(service memoryRecaller) AgentTool {
	return AgentTool{
		Name:         "recall_memories",
		Description:  "Recall active study memories for context.",
		RiskLevel:    ToolRiskLevelRead,
		InputSchema:  rawJSON(`{"type":"object","required":["query"],"properties":{"query":{"type":"string"},"limit":{"type":"integer","minimum":1,"maximum":20}}}`),
		OutputSchema: rawJSON(`{"type":"array"}`),
		Execute: func(ctx context.Context, input json.RawMessage) (ToolResult, error) {
			var req struct {
				Query string `json:"query"`
				Limit int    `json:"limit"`
			}
			if err := decodeToolInput(input, &req); err != nil {
				return ToolResult{}, err
			}
			if strings.TrimSpace(req.Query) == "" {
				return ToolResult{}, ErrInvalidToolInput
			}
			if req.Limit < 0 || req.Limit > 20 {
				return ToolResult{}, ErrInvalidToolInput
			}
			memories, err := service.RecallRelevantMemories(ctx, memories.RecallInput{Limit: req.Limit})
			if err != nil {
				return ToolResult{}, err
			}
			return rawOutput(memories)
		},
	}
}

func writeProposalTool(name, description string) AgentTool {
	return AgentTool{
		Name:         name,
		Description:  description,
		RiskLevel:    ToolRiskLevelWrite,
		InputSchema:  writeToolInputSchema(name),
		OutputSchema: rawJSON(`{"type":"object"}`),
		Execute: func(ctx context.Context, input json.RawMessage) (ToolResult, error) {
			if err := validateWriteToolInput(name, input); err != nil {
				return ToolResult{}, err
			}
			return ToolResult{
				Success:              true,
				RequiresConfirmation: true,
				ProposedAction: &ActionProposal{
					ToolName:   name,
					ActionType: name,
					Payload:    input,
					RiskLevel:  ToolRiskLevelWrite,
					Status:     ActionProposalStatusPending,
					CreatedAt:  time.Now(),
				},
			}, nil
		},
	}
}

func validateWriteToolInput(name string, input json.RawMessage) error {
	switch name {
	case "create_daily_task":
		var req struct {
			Date             string `json:"date"`
			ProjectID        *int64 `json:"project_id"`
			Title            string `json:"title"`
			EstimatedMinutes int    `json:"estimated_minutes"`
		}
		if err := decodeToolInput(input, &req); err != nil {
			return err
		}
		if validateDate(req.Date) != nil || req.ProjectID == nil || *req.ProjectID <= 0 ||
			strings.TrimSpace(req.Title) == "" || req.EstimatedMinutes <= 0 {
			return ErrInvalidToolInput
		}
	case "finish_task":
		var req struct {
			TaskID     int64  `json:"task_id"`
			FinishNote string `json:"finish_note"`
		}
		if err := decodeToolInput(input, &req); err != nil {
			return err
		}
		if req.TaskID <= 0 {
			return ErrInvalidToolInput
		}
	default:
		return ErrToolNotFound
	}
	return nil
}

func decodeToolInput(input json.RawMessage, out any) error {
	if err := json.Unmarshal(input, out); err != nil {
		return ErrInvalidToolInput
	}
	return nil
}

func validateDate(value string) error {
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(value)); err != nil {
		return ErrInvalidToolInput
	}
	return nil
}

func rawOutput(value any) (ToolResult, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return ToolResult{}, err
	}
	return ToolResult{Success: true, Output: data}, nil
}

func rawJSON(value string) json.RawMessage {
	return json.RawMessage(value)
}

func writeToolInputSchema(name string) json.RawMessage {
	switch name {
	case "create_daily_task":
		return rawJSON(`{"type":"object","required":["date","project_id","title","estimated_minutes"],"properties":{"date":{"type":"string","format":"date"},"project_id":{"type":"integer"},"title":{"type":"string"},"estimated_minutes":{"type":"integer","minimum":1}}}`)
	case "finish_task":
		return rawJSON(`{"type":"object","required":["task_id"],"properties":{"task_id":{"type":"integer"},"finish_note":{"type":"string"}}}`)
	default:
		return rawJSON(`{"type":"object"}`)
	}
}
