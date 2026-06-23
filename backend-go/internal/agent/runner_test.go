package agent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type memoryAgentStore struct {
	nextRunID  int64
	nextStepID int64
	runs       map[int64]*AgentRun
	steps      []AgentStep
}

func newMemoryAgentStore() *memoryAgentStore {
	return &memoryAgentStore{nextRunID: 1, nextStepID: 1, runs: map[int64]*AgentRun{}}
}

func (s *memoryAgentStore) CreateAgentRun(ctx context.Context, input CreateAgentRunInput) (*AgentRun, error) {
	run := &AgentRun{
		ID:         s.nextRunID,
		UserGoal:   input.UserGoal,
		TargetDate: input.TargetDate,
		Status:     input.Status,
		CreatedAt:  time.Now(),
	}
	s.nextRunID++
	s.runs[run.ID] = run
	return cloneRun(run), nil
}

func (s *memoryAgentStore) UpdateAgentRun(ctx context.Context, id int64, input UpdateAgentRunInput) error {
	run := s.runs[id]
	if run == nil {
		return ErrAgentRunNotFound
	}
	run.Status = input.Status
	run.FinalAnswer = input.FinalAnswer
	run.PendingActions = input.PendingActions
	run.ErrorMessage = input.ErrorMessage
	if input.Complete {
		now := time.Now()
		run.CompletedAt = &now
	}
	return nil
}

func (s *memoryAgentStore) CreateAgentStep(ctx context.Context, input CreateAgentStepInput) (*AgentStep, error) {
	step := AgentStep{
		ID:             s.nextStepID,
		RunID:          input.RunID,
		StepIndex:      input.StepIndex,
		StepType:       input.StepType,
		ToolName:       input.ToolName,
		ToolInput:      input.ToolInput,
		ToolOutput:     input.ToolOutput,
		ThoughtSummary: input.ThoughtSummary,
		Status:         input.Status,
		ErrorMessage:   input.ErrorMessage,
		CreatedAt:      time.Now(),
	}
	s.nextStepID++
	s.steps = append(s.steps, step)
	return &step, nil
}

func (s *memoryAgentStore) GetAgentRun(ctx context.Context, id int64) (*AgentRun, error) {
	run := s.runs[id]
	if run == nil {
		return nil, ErrAgentRunNotFound
	}
	return cloneRun(run), nil
}

func (s *memoryAgentStore) ListAgentSteps(ctx context.Context, runID int64) ([]AgentStep, error) {
	steps := make([]AgentStep, 0)
	for _, step := range s.steps {
		if step.RunID == runID {
			steps = append(steps, step)
		}
	}
	return steps, nil
}

func cloneRun(run *AgentRun) *AgentRun {
	copied := *run
	return &copied
}

type scriptedModel struct {
	decisions []AgentDecision
	err       error
	calls     int
}

func (m *scriptedModel) Decide(ctx context.Context, input AgentDecisionInput) (AgentDecision, error) {
	m.calls++
	if m.err != nil {
		return AgentDecision{}, m.err
	}
	if len(m.decisions) == 0 {
		return AgentDecision{}, nil
	}
	decision := m.decisions[0]
	if len(m.decisions) > 1 {
		m.decisions = m.decisions[1:]
	}
	return decision, nil
}

func testRegistryWithReadAndWrite(readCalls *int) *ToolRegistry {
	registry := NewToolRegistry()
	registry.Register(AgentTool{
		Name:      "read_tool",
		RiskLevel: ToolRiskLevelRead,
		Execute: func(ctx context.Context, input json.RawMessage) (ToolResult, error) {
			(*readCalls)++
			return ToolResult{Success: true, Output: json.RawMessage(`{"ok":true}`)}, nil
		},
	})
	registry.Register(writeProposalTool("create_daily_task", "Create a daily task proposal."))
	return registry
}

func TestRunnerCreatesRunAndBuildContextStep(t *testing.T) {
	readCalls := 0
	store := newMemoryAgentStore()
	runner := NewRunner(store, testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{{FinalAnswer: "done", ThoughtSummary: "short"}},
	})

	result, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	if result.Run.Status != AgentRunStatusCompleted || result.Run.FinalAnswer != "done" {
		t.Fatalf("run = %+v", result.Run)
	}
	if result.Steps[0].StepType != AgentStepTypeBuildContext {
		t.Fatalf("first step = %s, want build_context", result.Steps[0].StepType)
	}
}

func TestRunnerExecutesReadToolAndStoresOutput(t *testing.T) {
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{
			{ToolCalls: []ToolCall{{ToolName: "read_tool", Input: json.RawMessage(`{}`)}}, ThoughtSummary: "use read"},
			{FinalAnswer: "done after read", ThoughtSummary: "answer"},
		},
	})

	result, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	if readCalls != 1 {
		t.Fatalf("read calls = %d, want 1", readCalls)
	}
	toolStep := findStep(result.Steps, AgentStepTypeToolCall)
	if toolStep == nil || !strings.Contains(string(toolStep.ToolOutput), `"ok":true`) {
		t.Fatalf("tool step = %+v", toolStep)
	}
}

func TestRunnerWriteToolRequiresConfirmationWithoutWrite(t *testing.T) {
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{{
			ToolCalls: []ToolCall{{ToolName: "create_daily_task", Input: json.RawMessage(`{"date":"2026-06-23","project_id":1,"title":"Go review","estimated_minutes":60}`)}},
		}},
	})

	result, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	if result.Run.Status != AgentRunStatusRequiresConfirmation {
		t.Fatalf("status = %s, want requires_confirmation", result.Run.Status)
	}
	if len(result.Run.PendingActions) == 0 || !strings.Contains(string(result.Run.PendingActions), "create_daily_task") {
		t.Fatalf("pending actions = %s", result.Run.PendingActions)
	}
	if readCalls != 0 {
		t.Fatalf("read/write side-effect calls = %d, want 0", readCalls)
	}
}

func TestRunnerUnknownToolFailsRun(t *testing.T) {
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{{ToolCalls: []ToolCall{{ToolName: "missing", Input: json.RawMessage(`{}`)}}}},
	})

	result, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	if result.Run.Status != AgentRunStatusFailed || result.Run.ErrorMessage == "" {
		t.Fatalf("run = %+v, want failed", result.Run)
	}
}

func TestRunnerModelErrorFailsRun(t *testing.T) {
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{err: errors.New("model down")})

	result, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	if result.Run.Status != AgentRunStatusFailed || result.Run.ErrorMessage != "model_decision_failed" {
		t.Fatalf("run = %+v", result.Run)
	}
}

func TestRunnerMaxStepsExceeded(t *testing.T) {
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{{ToolCalls: []ToolCall{{ToolName: "read_tool", Input: json.RawMessage(`{}`)}}}},
	})

	result, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	if result.Run.Status != AgentRunStatusFailed || result.Run.ErrorMessage != "max_steps_exceeded" {
		t.Fatalf("run = %+v", result.Run)
	}
}

func TestRunnerStoresOnlyThoughtSummary(t *testing.T) {
	readCalls := 0
	longThought := strings.Repeat("chain ", 200)
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{{FinalAnswer: "done", ThoughtSummary: longThought}},
	})

	result, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	modelStep := findStep(result.Steps, AgentStepTypeModelDecision)
	if modelStep == nil {
		t.Fatal("missing model decision step")
	}
	if len([]rune(modelStep.ThoughtSummary)) > 500 {
		t.Fatalf("thought summary length = %d, want <= 500", len([]rune(modelStep.ThoughtSummary)))
	}
}

func TestAgentRunHandlerInvalidInputReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{})
	handler := NewHandler(testRegistryWithReadAndWrite(&readCalls), testContextBuilder(), runner)
	router := gin.New()
	router.POST("/api/agent/runs", handler.CreateRun)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, "/api/agent/runs", strings.NewReader(`{"goal":"","target_date":"bad"}`)))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.Code)
	}
}

func findStep(steps []AgentStep, stepType AgentStepType) *AgentStep {
	for i := range steps {
		if steps[i].StepType == stepType {
			return &steps[i]
		}
	}
	return nil
}
