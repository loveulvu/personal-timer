package agent

import (
	"context"
	"encoding/json"
	"errors"
	"personal/internal/dailytasks"
	"personal/internal/memories"
	"personal/internal/plans"
	"testing"
)

type fakeTaskLister struct {
	createCalled bool
}

func (f *fakeTaskLister) ListDailyTasksByDate(date string) ([]dailytasks.DailyTask, error) {
	return []dailytasks.DailyTask{{ID: 1, TaskDate: date, Title: "Review Go", EstimatedMinutes: 45, Status: "planned"}}, nil
}

type fakePlanRiskGetter struct{}

func (f fakePlanRiskGetter) GetPlanRisk(ctx context.Context, date string) (*plans.PlanRiskResponse, error) {
	return &plans.PlanRiskResponse{Date: date, PlannedTotalMinutes: 45, RiskLevel: plans.PlanRiskLow}, nil
}

type fakeMemoryRecaller struct{}

func (f fakeMemoryRecaller) RecallRelevantMemories(ctx context.Context, input memories.RecallInput) ([]memories.StudyMemory, error) {
	return []memories.StudyMemory{{ID: 2, Title: "Evening works", Status: "active", Confidence: 0.8}}, nil
}

func testRegistry() *ToolRegistry {
	return NewDefaultToolRegistry(&fakeTaskLister{}, fakePlanRiskGetter{}, fakeMemoryRecaller{})
}

func TestToolRegistryListsToolsWithoutExecutors(t *testing.T) {
	tools := testRegistry().ListTools()
	if len(tools) != 5 {
		t.Fatalf("tool count = %d, want 5", len(tools))
	}
	for _, tool := range tools {
		if tool.Execute != nil {
			t.Fatalf("tool %s exposed executor", tool.Name)
		}
	}
}

func TestToolRegistryUnknownTool(t *testing.T) {
	_, err := testRegistry().Call(context.Background(), ToolCall{ToolName: "missing", Input: json.RawMessage(`{}`)})
	if !errors.Is(err, ErrToolNotFound) {
		t.Fatalf("err = %v, want ErrToolNotFound", err)
	}
}

func TestReadToolsExecute(t *testing.T) {
	r := testRegistry()
	tests := []ToolCall{
		{ToolName: "list_today_tasks", Input: json.RawMessage(`{"date":"2026-06-23"}`)},
		{ToolName: "evaluate_plan_risk", Input: json.RawMessage(`{"date":"2026-06-23"}`)},
		{ToolName: "recall_memories", Input: json.RawMessage(`{"query":"planning","limit":1}`)},
	}
	for _, tt := range tests {
		result, err := r.Call(context.Background(), tt)
		if err != nil {
			t.Fatalf("%s err = %v", tt.ToolName, err)
		}
		if !result.Success || len(result.Output) == 0 || result.RequiresConfirmation {
			t.Fatalf("%s result = %+v", tt.ToolName, result)
		}
	}
}

func TestWriteToolReturnsConfirmationWithoutCallingTaskWrite(t *testing.T) {
	result, err := testRegistry().Call(context.Background(), ToolCall{
		ToolName: "create_daily_task",
		Input:    json.RawMessage(`{"date":"2026-06-23","project_id":1,"title":"Review Go","estimated_minutes":45}`),
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !result.Success || !result.RequiresConfirmation {
		t.Fatalf("result = %+v, want confirmation", result)
	}
	if result.ProposedAction == nil || result.ProposedAction.Status != ActionProposalStatusPending {
		t.Fatalf("proposal = %+v, want pending proposal", result.ProposedAction)
	}
	if result.ProposedAction.RiskLevel != ToolRiskLevelWrite {
		t.Fatalf("risk = %q, want write", result.ProposedAction.RiskLevel)
	}
}

func TestInvalidJSONInputReturnsError(t *testing.T) {
	_, err := testRegistry().Call(context.Background(), ToolCall{
		ToolName: "list_today_tasks",
		Input:    json.RawMessage(`{"date":`),
	})
	if !errors.Is(err, ErrInvalidToolInput) {
		t.Fatalf("err = %v, want ErrInvalidToolInput", err)
	}
}

func TestRiskLevelSerializes(t *testing.T) {
	data, err := json.Marshal(AgentTool{Name: "x", RiskLevel: ToolRiskLevelRead})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != `{"name":"x","description":"","risk_level":"read"}` {
		t.Fatalf("json = %s", data)
	}
}
