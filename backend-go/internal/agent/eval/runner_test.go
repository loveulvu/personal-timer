package eval

import (
	"context"
	"encoding/json"
	"errors"
	"personal/internal/agent"
	"strings"
	"testing"
)

type fakeExecutor map[string]EvalObservation

func (f fakeExecutor) ExecuteEvalCase(ctx context.Context, c EvalCase) (EvalObservation, error) {
	obs, ok := f[c.Name]
	if !ok {
		return EvalObservation{}, errors.New("missing fake case")
	}
	return obs, nil
}

func TestReadCaseDoesNotGenerateProposal(t *testing.T) {
	result := NewRunner(fakeExecutor{
		"read_today_tasks_no_confirmation": {
			RunID:     1,
			RunStatus: agent.AgentRunStatusCompleted,
			Steps:     []agent.AgentStep{{ToolName: "list_today_tasks"}},
		},
	}).RunOne(context.Background(), FixedCases()[0])

	if !result.Passed {
		t.Fatalf("result = %+v", result)
	}
}

func TestWriteCaseRequiresProposalWithoutBusinessWrite(t *testing.T) {
	result := NewRunner(fakeExecutor{
		"task_creation_requires_confirmation": {
			RunID:     2,
			RunStatus: agent.AgentRunStatusRequiresConfirmation,
			Steps:     []agent.AgentStep{{ToolName: "create_daily_task"}},
			Proposals: []agent.ActionProposal{{ActionType: "create_daily_task", Status: agent.ActionProposalStatusPending}},
		},
	}).RunOne(context.Background(), FixedCases()[1])

	if !result.Passed {
		t.Fatalf("result = %+v", result)
	}
}

func TestRejectProposalNoWrite(t *testing.T) {
	result := NewRunner(fakeExecutor{
		"reject_proposal_no_write": {
			RunID:     3,
			RunStatus: agent.AgentRunStatusRequiresConfirmation,
			Proposals: []agent.ActionProposal{{ActionType: "create_daily_task", Status: agent.ActionProposalStatusRejected}},
		},
	}).RunOne(context.Background(), FixedCases()[2])

	if !result.Passed {
		t.Fatalf("result = %+v", result)
	}
}

func TestRepeatedAcceptIdempotentObservedAsSingleBusinessWrite(t *testing.T) {
	result := NewRunner(fakeExecutor{
		"repeated_accept_idempotent": {
			RunID:              4,
			RunStatus:          agent.AgentRunStatusRequiresConfirmation,
			BusinessWriteCount: 1,
			Proposals: []agent.ActionProposal{{
				ActionType: "create_daily_task",
				Status:     agent.ActionProposalStatusExecuted,
				Result:     json.RawMessage(`{"task_id":10}`),
			}},
		},
	}).RunOne(context.Background(), FixedCases()[3])

	if !result.Passed {
		t.Fatalf("result = %+v", result)
	}
}

func TestContextSnapshotExpectations(t *testing.T) {
	result := NewRunner(fakeExecutor{
		"memory_context_constraints_present": {
			RunID:     5,
			RunStatus: agent.AgentRunStatusCompleted,
			ContextSnapshot: &agent.AgentContextSnapshot{
				OmittedSections: []string{"summary_content_excerpted"},
				ContextPack: agent.ContextPack{
					Constraints: []string{"write tools require user confirmation"},
					Memories:    []agent.ContextMemory{{ID: 1, Status: "active"}},
				},
			},
		},
	}).RunOne(context.Background(), FixedCases()[4])

	if !result.Passed {
		t.Fatalf("result = %+v", result)
	}
}

type fakeTrajectoryReader struct {
	calls      int
	trajectory *agent.AgentTrajectory
	err        error
}

func (f *fakeTrajectoryReader) GetTrajectory(ctx context.Context, id int64) (*agent.AgentTrajectory, error) {
	f.calls++
	return f.trajectory, f.err
}

func TestReplayWithoutLLMReadsTrajectoryOnly(t *testing.T) {
	reader := &fakeTrajectoryReader{trajectory: &agent.AgentTrajectory{
		Run:             agent.AgentRun{ID: 9, Status: agent.AgentRunStatusRequiresConfirmation},
		ContextSnapshot: &agent.AgentContextSnapshot{},
		Steps:           []agent.AgentStep{{StepType: agent.AgentStepTypeBuildContext}},
		Proposals:       []agent.ActionProposal{{ActionType: "create_daily_task", Status: agent.ActionProposalStatusPending}},
	}}

	summary, err := ReplayWithoutLLM(context.Background(), reader, 9)
	if err != nil {
		t.Fatalf("ReplayWithoutLLM err = %v", err)
	}
	if reader.calls != 1 || summary.PendingProposalCount != 1 || !summary.ContextSnapshotPresent {
		t.Fatalf("summary = %+v calls = %d", summary, reader.calls)
	}
	if !strings.Contains(summary.Summary, "waiting for confirmation") {
		t.Fatalf("summary text = %q", summary.Summary)
	}
}

func TestReplayUnknownRunReturnsError(t *testing.T) {
	_, err := ReplayWithoutLLM(context.Background(), &fakeTrajectoryReader{err: agent.ErrAgentRunNotFound}, 404)
	if !errors.Is(err, agent.ErrAgentRunNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestEvalResultHasCaseNameReasonAndNoChainOfThought(t *testing.T) {
	result := EvalResult{CaseName: "case", Passed: false, Reason: "failed", RunID: 1}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal err = %v", err)
	}
	if !strings.Contains(string(data), "case_name") || strings.Contains(string(data), "chain_of_thought") {
		t.Fatalf("json = %s", data)
	}
}
