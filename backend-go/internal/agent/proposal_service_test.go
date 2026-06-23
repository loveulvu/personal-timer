package agent

import (
	"context"
	"encoding/json"
	"errors"
	"personal/internal/dailytasks"
	"personal/internal/timer"
	"strings"
	"testing"
)

type fakeDailyTasks struct {
	created int
	tasks   map[int64]*dailytasks.DailyTask
}

func newFakeDailyTasks() *fakeDailyTasks {
	return &fakeDailyTasks{tasks: map[int64]*dailytasks.DailyTask{}}
}

func (f *fakeDailyTasks) CreateDailyTask(req dailytasks.CreateDailyTaskRequest) (int64, error) {
	f.created++
	id := int64(f.created)
	f.tasks[id] = &dailytasks.DailyTask{
		ID:               id,
		ProjectID:        req.ProjectID,
		TaskDate:         req.TaskDate,
		Title:            req.Title,
		EstimatedMinutes: req.EstimatedMinutes,
		Status:           "planned",
	}
	return id, nil
}

func (f *fakeDailyTasks) GetDailyTaskByID(id int64) (*dailytasks.DailyTask, error) {
	task := f.tasks[id]
	if task == nil {
		return nil, errors.New("task not found")
	}
	return task, nil
}

type fakeTimer struct {
	finished int
	lastID   int64
}

func (f *fakeTimer) FinishTask(taskID int64, input timer.FinishTaskInput) error {
	if strings.TrimSpace(input.FinishNote) == "" {
		return timer.ErrFinishNoteRequired
	}
	if strings.TrimSpace(input.FinishDescription) == "" {
		return timer.ErrFinishDescriptionRequired
	}
	f.finished++
	f.lastID = taskID
	return nil
}

func proposalServiceWithStore() (*ProposalService, *memoryAgentStore, *fakeDailyTasks, *fakeTimer) {
	store := newMemoryAgentStore()
	tasks := newFakeDailyTasks()
	finisher := &fakeTimer{}
	return NewProposalService(store, tasks, finisher), store, tasks, finisher
}

func createPendingProposal(t *testing.T, store *memoryAgentStore, actionType string, payload json.RawMessage) *ActionProposal {
	t.Helper()
	proposal, err := store.CreateActionProposal(context.Background(), CreateActionProposalInput{
		RunID:      1,
		ToolName:   actionType,
		ActionType: actionType,
		Payload:    payload,
		RiskLevel:  ToolRiskLevelWrite,
		Status:     ActionProposalStatusPending,
	})
	if err != nil {
		t.Fatalf("CreateActionProposal: %v", err)
	}
	return proposal
}

func TestProposalServiceListsPending(t *testing.T) {
	service, store, _, _ := proposalServiceWithStore()
	createPendingProposal(t, store, "create_daily_task", json.RawMessage(`{"date":"2026-06-23","project_id":1,"title":"Go","estimated_minutes":30}`))

	items, err := service.List(context.Background(), "pending")
	if err != nil {
		t.Fatalf("List err = %v", err)
	}
	if len(items) != 1 || items[0].Status != ActionProposalStatusPending {
		t.Fatalf("items = %+v", items)
	}
}

func TestProposalRejectDoesNotCreateTask(t *testing.T) {
	service, store, tasks, _ := proposalServiceWithStore()
	proposal := createPendingProposal(t, store, "create_daily_task", json.RawMessage(`{"date":"2026-06-23","project_id":1,"title":"Go","estimated_minutes":30}`))

	rejected, err := service.Reject(context.Background(), proposal.ID)
	if err != nil {
		t.Fatalf("Reject err = %v", err)
	}
	if rejected.Status != ActionProposalStatusRejected || tasks.created != 0 {
		t.Fatalf("proposal = %+v created = %d", rejected, tasks.created)
	}
}

func TestProposalAcceptCreateDailyTaskIsIdempotent(t *testing.T) {
	service, store, tasks, _ := proposalServiceWithStore()
	proposal := createPendingProposal(t, store, "create_daily_task", json.RawMessage(`{"date":"2026-06-23","project_id":1,"title":"Go","estimated_minutes":30}`))

	first, err := service.Accept(context.Background(), proposal.ID)
	if err != nil {
		t.Fatalf("first Accept err = %v", err)
	}
	second, err := service.Accept(context.Background(), proposal.ID)
	if err != nil {
		t.Fatalf("second Accept err = %v", err)
	}
	if tasks.created != 1 {
		t.Fatalf("created = %d, want 1", tasks.created)
	}
	if first.Status != ActionProposalStatusExecuted || second.Status != ActionProposalStatusExecuted ||
		first.TargetEntityID == nil || second.TargetEntityID == nil || *first.TargetEntityID != *second.TargetEntityID {
		t.Fatalf("first = %+v second = %+v", first, second)
	}
	if !strings.Contains(string(first.Result), "task_id") {
		t.Fatalf("result = %s, want task_id", first.Result)
	}
}

func TestProposalAcceptRejectedConflicts(t *testing.T) {
	service, store, tasks, _ := proposalServiceWithStore()
	proposal := createPendingProposal(t, store, "create_daily_task", json.RawMessage(`{"date":"2026-06-23","project_id":1,"title":"Go","estimated_minutes":30}`))
	if _, err := service.Reject(context.Background(), proposal.ID); err != nil {
		t.Fatalf("Reject err = %v", err)
	}

	_, err := service.Accept(context.Background(), proposal.ID)
	if !errors.Is(err, ErrProposalConflict) {
		t.Fatalf("err = %v, want ErrProposalConflict", err)
	}
	if tasks.created != 0 {
		t.Fatalf("created = %d, want 0", tasks.created)
	}
}

func TestProposalAcceptFinishTaskUsesTimerValidation(t *testing.T) {
	service, store, _, finisher := proposalServiceWithStore()
	proposal := createPendingProposal(t, store, "finish_task", json.RawMessage(`{"task_id":7,"finish_note":"done","finish_description":"reviewed GC"}`))

	executed, err := service.Accept(context.Background(), proposal.ID)
	if err != nil {
		t.Fatalf("Accept err = %v", err)
	}
	if executed.Status != ActionProposalStatusExecuted || finisher.finished != 1 || finisher.lastID != 7 {
		t.Fatalf("executed = %+v finished = %d lastID = %d", executed, finisher.finished, finisher.lastID)
	}
}

func TestProposalInvalidPayloadFailsAndRecordsError(t *testing.T) {
	service, store, tasks, _ := proposalServiceWithStore()
	proposal := createPendingProposal(t, store, "create_daily_task", json.RawMessage(`{"date":"bad"}`))

	updated, err := service.Accept(context.Background(), proposal.ID)
	if !errors.Is(err, ErrInvalidToolInput) {
		t.Fatalf("err = %v, want ErrInvalidToolInput", err)
	}
	if updated.Status != ActionProposalStatusFailed || updated.ErrorMessage == "" || tasks.created != 0 {
		t.Fatalf("updated = %+v created = %d", updated, tasks.created)
	}
}
