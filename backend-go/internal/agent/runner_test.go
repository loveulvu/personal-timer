package agent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type memoryAgentStore struct {
	nextRunID      int64
	nextStepID     int64
	nextProposalID int64
	nextSnapshotID int64
	runs           map[int64]*AgentRun
	steps          []AgentStep
	proposals      map[int64]*ActionProposal
	snapshots      map[int64]*AgentContextSnapshot
}

func newMemoryAgentStore() *memoryAgentStore {
	return &memoryAgentStore{
		nextRunID:      1,
		nextStepID:     1,
		nextProposalID: 1,
		nextSnapshotID: 1,
		runs:           map[int64]*AgentRun{},
		proposals:      map[int64]*ActionProposal{},
		snapshots:      map[int64]*AgentContextSnapshot{},
	}
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

func (s *memoryAgentStore) CreateContextSnapshot(ctx context.Context, input CreateContextSnapshotInput) (*AgentContextSnapshot, error) {
	snapshot := &AgentContextSnapshot{
		ID:                  s.nextSnapshotID,
		RunID:               input.RunID,
		ContextJSON:         input.ContextJSON,
		TokenEstimate:       input.TokenEstimate,
		OmittedSectionsJSON: input.OmittedSectionsJSON,
		CreatedAt:           time.Now(),
	}
	_ = json.Unmarshal(snapshot.ContextJSON, &snapshot.ContextPack)
	_ = json.Unmarshal(snapshot.OmittedSectionsJSON, &snapshot.OmittedSections)
	s.nextSnapshotID++
	s.snapshots[input.RunID] = snapshot
	return cloneSnapshot(snapshot), nil
}

func (s *memoryAgentStore) GetContextSnapshot(ctx context.Context, runID int64) (*AgentContextSnapshot, error) {
	snapshot := s.snapshots[runID]
	if snapshot == nil {
		return nil, ErrContextSnapshotNotFound
	}
	return cloneSnapshot(snapshot), nil
}

func (s *memoryAgentStore) ListAgentRuns(ctx context.Context, filter AgentRunFilter) ([]AgentRunListItem, error) {
	items := make([]AgentRunListItem, 0)
	for _, run := range s.runs {
		if filter.Status != "" && string(run.Status) != filter.Status {
			continue
		}
		item := AgentRunListItem{
			ID:                 run.ID,
			UserGoal:           run.UserGoal,
			TargetDate:         run.TargetDate,
			Status:             run.Status,
			FinalAnswerExcerpt: run.FinalAnswer,
			CreatedAt:          run.CreatedAt,
			CompletedAt:        run.CompletedAt,
		}
		for _, step := range s.steps {
			if step.RunID == run.ID {
				item.StepCount++
			}
		}
		for _, proposal := range s.proposals {
			if proposal.RunID == run.ID {
				item.ProposalCount++
				if proposal.Status == ActionProposalStatusPending {
					item.PendingProposalCount++
				}
			}
		}
		items = append(items, item)
		if filter.Limit > 0 && len(items) >= filter.Limit {
			break
		}
	}
	return items, nil
}

func (s *memoryAgentStore) CreateActionProposal(ctx context.Context, input CreateActionProposalInput) (*ActionProposal, error) {
	status := input.Status
	if status == "" {
		status = ActionProposalStatusPending
	}
	proposal := &ActionProposal{
		ID:         s.nextProposalID,
		RunID:      input.RunID,
		ToolName:   input.ToolName,
		ActionType: input.ActionType,
		Payload:    input.Payload,
		RiskLevel:  input.RiskLevel,
		Status:     status,
		CreatedAt:  time.Now(),
	}
	if input.StepID != nil {
		proposal.StepID = *input.StepID
	}
	s.nextProposalID++
	s.proposals[proposal.ID] = proposal
	return cloneProposal(proposal), nil
}

func (s *memoryAgentStore) GetActionProposal(ctx context.Context, id int64) (*ActionProposal, error) {
	proposal := s.proposals[id]
	if proposal == nil {
		return nil, ErrProposalNotFound
	}
	return cloneProposal(proposal), nil
}

func (s *memoryAgentStore) ListActionProposals(ctx context.Context, filter ActionProposalFilter) ([]ActionProposal, error) {
	items := make([]ActionProposal, 0)
	for _, proposal := range s.proposals {
		if filter.RunID > 0 && proposal.RunID != filter.RunID {
			continue
		}
		if filter.Status != "" && string(proposal.Status) != filter.Status {
			continue
		}
		items = append(items, *cloneProposal(proposal))
	}
	return items, nil
}

func (s *memoryAgentStore) UpdateActionProposal(ctx context.Context, id int64, input UpdateActionProposalInput) (*ActionProposal, error) {
	proposal := s.proposals[id]
	if proposal == nil {
		return nil, ErrProposalNotFound
	}
	proposal.Status = input.Status
	proposal.Result = input.Result
	proposal.ErrorMessage = input.ErrorMessage
	proposal.TargetEntityType = input.TargetEntityType
	proposal.TargetEntityID = input.TargetEntityID
	now := time.Now()
	if input.Decide && proposal.DecidedAt == nil {
		proposal.DecidedAt = &now
	}
	if input.Execute && proposal.ExecutedAt == nil {
		proposal.ExecutedAt = &now
	}
	return cloneProposal(proposal), nil
}

func cloneRun(run *AgentRun) *AgentRun {
	copied := *run
	return &copied
}

func cloneProposal(proposal *ActionProposal) *ActionProposal {
	copied := *proposal
	return &copied
}

func cloneSnapshot(snapshot *AgentContextSnapshot) *AgentContextSnapshot {
	copied := *snapshot
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

func TestRunnerSavesContextSnapshot(t *testing.T) {
	readCalls := 0
	store := newMemoryAgentStore()
	runner := NewRunner(store, testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{{FinalAnswer: "done", ThoughtSummary: "short"}},
	})

	result, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23", RecentDays: 99})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	snapshot, err := store.GetContextSnapshot(context.Background(), result.Run.ID)
	if err != nil {
		t.Fatalf("GetContextSnapshot err = %v", err)
	}
	if snapshot.TokenEstimate <= 0 {
		t.Fatalf("token estimate = %d, want > 0", snapshot.TokenEstimate)
	}
	if !contains(snapshot.OmittedSections, "recent_days_capped_to_14") {
		t.Fatalf("omitted = %v, want recent_days_capped_to_14", snapshot.OmittedSections)
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
	if len(result.Proposals) != 1 || result.Proposals[0].Status != ActionProposalStatusPending {
		t.Fatalf("proposals = %+v, want pending proposal", result.Proposals)
	}
	if readCalls != 0 {
		t.Fatalf("read/write side-effect calls = %d, want 0", readCalls)
	}
}

func TestDeterministicModelCreatesTaskProposalForDemoGoal(t *testing.T) {
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), NewDeterministicModelClient())

	result, err := runner.Start(context.Background(), AgentRunRequest{
		Goal:       "创建任务：今天安排一个 60 分钟的 Go GC 复习任务，项目使用 personal_study_timer",
		TargetDate: "2026-06-23",
		RecentDays: 5,
	})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	if result.Run.Status != AgentRunStatusRequiresConfirmation {
		t.Fatalf("status = %s, want requires_confirmation", result.Run.Status)
	}
	if len(result.Proposals) != 1 || result.Proposals[0].Status != ActionProposalStatusPending {
		t.Fatalf("proposals = %+v, want one pending proposal", result.Proposals)
	}
	if result.Proposals[0].ActionType != "create_daily_task" {
		t.Fatalf("action_type = %s, want create_daily_task", result.Proposals[0].ActionType)
	}
	if !strings.Contains(string(result.Proposals[0].Payload), `"title":"Go GC 复习"`) ||
		!strings.Contains(string(result.Proposals[0].Payload), `"estimated_minutes":60`) ||
		!strings.Contains(string(result.Proposals[0].Payload), `"project_id":1`) {
		t.Fatalf("payload = %s", result.Proposals[0].Payload)
	}
	if readCalls != 0 {
		t.Fatalf("side-effect calls = %d, want 0", readCalls)
	}
}

func TestDeterministicModelFailsWhenProjectNameMissingFromContext(t *testing.T) {
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), NewDeterministicModelClient())

	result, err := runner.Start(context.Background(), AgentRunRequest{
		Goal:       "创建任务：今天安排一个 60 分钟的 Go GC 复习任务，项目使用 missing_project",
		TargetDate: "2026-06-23",
		RecentDays: 5,
	})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	if result.Run.Status != AgentRunStatusFailed {
		t.Fatalf("status = %s, want failed", result.Run.Status)
	}
	if result.Run.ErrorMessage != "cannot infer project_id: project name not found in context" {
		t.Fatalf("error = %q", result.Run.ErrorMessage)
	}
	if len(result.Proposals) != 0 {
		t.Fatalf("proposals = %+v, want none", result.Proposals)
	}
	if readCalls != 0 {
		t.Fatalf("side-effect calls = %d, want 0", readCalls)
	}
}

func TestDeterministicModelPlanningGoalStaysReadOnlyCompleted(t *testing.T) {
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), NewDeterministicModelClient())

	result, err := runner.Start(context.Background(), AgentRunRequest{
		Goal:       "根据最近 5 天学习记录，帮我安排今天的 Go 后端复习计划",
		TargetDate: "2026-06-23",
		RecentDays: 5,
	})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	if result.Run.Status != AgentRunStatusCompleted {
		t.Fatalf("status = %s, want completed", result.Run.Status)
	}
	if len(result.Proposals) != 0 {
		t.Fatalf("proposals = %+v, want none", result.Proposals)
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

func TestRunnerTrajectoryIncludesSnapshotStepsAndProposals(t *testing.T) {
	readCalls := 0
	store := newMemoryAgentStore()
	runner := NewRunner(store, testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{{
			ToolCalls: []ToolCall{{ToolName: "create_daily_task", Input: json.RawMessage(`{"date":"2026-06-23","project_id":1,"title":"Go review","estimated_minutes":60}`)}},
		}},
	})
	run, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}

	trajectory, err := runner.GetTrajectory(context.Background(), run.Run.ID)
	if err != nil {
		t.Fatalf("GetTrajectory err = %v", err)
	}
	if trajectory.ContextSnapshot == nil || len(trajectory.Steps) == 0 || len(trajectory.Proposals) != 1 {
		t.Fatalf("trajectory = %+v", trajectory)
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

func TestRunnerListRuns(t *testing.T) {
	readCalls := 0
	store := newMemoryAgentStore()
	runner := NewRunner(store, testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{{FinalAnswer: "done", ThoughtSummary: "short"}},
	})
	if _, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"}); err != nil {
		t.Fatalf("Start err = %v", err)
	}
	items, err := runner.ListRuns(context.Background(), "", 20)
	if err != nil {
		t.Fatalf("ListRuns err = %v", err)
	}
	if len(items) != 1 || items[0].StepCount == 0 {
		t.Fatalf("items = %+v", items)
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

func TestTrajectoryHandlerReturns400ForInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	readCalls := 0
	runner := NewRunner(newMemoryAgentStore(), testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{})
	handler := NewHandler(testRegistryWithReadAndWrite(&readCalls), testContextBuilder(), runner)
	router := gin.New()
	router.GET("/api/agent/runs/:id/trajectory", handler.GetRunTrajectory)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/api/agent/runs/bad/trajectory", nil))
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.Code)
	}
}

func TestTrajectoryAndRunsHandlersReturnData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	readCalls := 0
	store := newMemoryAgentStore()
	runner := NewRunner(store, testContextBuilder(), testRegistryWithReadAndWrite(&readCalls), &scriptedModel{
		decisions: []AgentDecision{{FinalAnswer: "done", ThoughtSummary: "short"}},
	})
	run, err := runner.Start(context.Background(), AgentRunRequest{Goal: "plan", TargetDate: "2026-06-23"})
	if err != nil {
		t.Fatalf("Start err = %v", err)
	}
	handler := NewHandler(testRegistryWithReadAndWrite(&readCalls), testContextBuilder(), runner)
	router := gin.New()
	router.GET("/api/agent/runs", handler.ListRuns)
	router.GET("/api/agent/runs/:id/trajectory", handler.GetRunTrajectory)

	trajectoryResp := httptest.NewRecorder()
	router.ServeHTTP(trajectoryResp, httptest.NewRequest(http.MethodGet, "/api/agent/runs/"+strconv.FormatInt(run.Run.ID, 10)+"/trajectory", nil))
	if trajectoryResp.Code != http.StatusOK || !strings.Contains(trajectoryResp.Body.String(), "context_snapshot") {
		t.Fatalf("trajectory status = %d body = %s", trajectoryResp.Code, trajectoryResp.Body.String())
	}

	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, httptest.NewRequest(http.MethodGet, "/api/agent/runs", nil))
	if listResp.Code != http.StatusOK || !strings.Contains(listResp.Body.String(), "runs") {
		t.Fatalf("list status = %d body = %s", listResp.Code, listResp.Body.String())
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
