package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const maxAgentDecisions = 3

var ErrMaxStepsExceeded = errors.New("max_steps_exceeded")

type AgentRunRequest struct {
	Goal       string `json:"goal"`
	TargetDate string `json:"target_date"`
	RecentDays int    `json:"recent_days"`
}

type AgentRunResponse struct {
	Run   AgentRun    `json:"run"`
	Steps []AgentStep `json:"steps"`
}

type agentRunStore interface {
	CreateAgentRun(ctx context.Context, input CreateAgentRunInput) (*AgentRun, error)
	UpdateAgentRun(ctx context.Context, id int64, input UpdateAgentRunInput) error
	CreateAgentStep(ctx context.Context, input CreateAgentStepInput) (*AgentStep, error)
	GetAgentRun(ctx context.Context, id int64) (*AgentRun, error)
	ListAgentSteps(ctx context.Context, runID int64) ([]AgentStep, error)
}

type Runner struct {
	store    agentRunStore
	builder  *ContextPackBuilder
	registry *ToolRegistry
	model    ModelClient
}

func NewRunner(store agentRunStore, builder *ContextPackBuilder, registry *ToolRegistry, model ModelClient) *Runner {
	if model == nil {
		model = NewDeterministicModelClient()
	}
	return &Runner{store: store, builder: builder, registry: registry, model: model}
}

func (r *Runner) Start(ctx context.Context, req AgentRunRequest) (*AgentRunResponse, error) {
	contextReq, err := normalizeRunRequest(req)
	if err != nil {
		return nil, err
	}
	run, err := r.store.CreateAgentRun(ctx, CreateAgentRunInput{
		UserGoal:   contextReq.Goal,
		TargetDate: contextReq.TargetDate,
		Status:     AgentRunStatusRunning,
	})
	if err != nil {
		return nil, err
	}

	stepIndex := 1
	pack, err := r.builder.Build(ctx, contextReq)
	if err != nil {
		return r.fail(ctx, run.ID, stepIndex, AgentStepTypeBuildContext, "", nil, nil, "", "context_build_failed")
	}
	packJSON, _ := json.Marshal(pack)
	if _, err := r.store.CreateAgentStep(ctx, CreateAgentStepInput{
		RunID:      run.ID,
		StepIndex:  stepIndex,
		StepType:   AgentStepTypeBuildContext,
		ToolOutput: packJSON,
		Status:     AgentStepStatusCompleted,
	}); err != nil {
		return nil, err
	}
	stepIndex++

	observations := []ToolObservation{}
	for decisionCount := 0; decisionCount < maxAgentDecisions; decisionCount++ {
		decisionInput := AgentDecisionInput{
			RunID:          run.ID,
			ContextPack:    pack,
			AvailableTools: r.registry.ListTools(),
			Observations:   observations,
		}
		decision, err := r.model.Decide(ctx, decisionInput)
		if err != nil {
			return r.fail(ctx, run.ID, stepIndex, AgentStepTypeModelDecision, "", nil, nil, "", "model_decision_failed")
		}
		decision.ThoughtSummary = thoughtSummary(decision.ThoughtSummary)
		decisionJSON, _ := json.Marshal(decision)
		if _, err := r.store.CreateAgentStep(ctx, CreateAgentStepInput{
			RunID:          run.ID,
			StepIndex:      stepIndex,
			StepType:       AgentStepTypeModelDecision,
			ToolOutput:     decisionJSON,
			ThoughtSummary: decision.ThoughtSummary,
			Status:         AgentStepStatusCompleted,
		}); err != nil {
			return nil, err
		}
		stepIndex++

		if decision.UnsupportedGoal || strings.TrimSpace(decision.ErrorMessage) != "" {
			msg := strings.TrimSpace(decision.ErrorMessage)
			if msg == "" {
				msg = "unsupported_goal"
			}
			return r.finish(ctx, run.ID, AgentRunStatusFailed, "", nil, msg)
		}
		if strings.TrimSpace(decision.FinalAnswer) != "" && len(decision.ToolCalls) == 0 {
			answer := strings.TrimSpace(decision.FinalAnswer)
			if _, err := r.store.CreateAgentStep(ctx, CreateAgentStepInput{
				RunID:      run.ID,
				StepIndex:  stepIndex,
				StepType:   AgentStepTypeFinalAnswer,
				ToolOutput: mustJSON(map[string]string{"final_answer": answer}),
				Status:     AgentStepStatusCompleted,
			}); err != nil {
				return nil, err
			}
			return r.finish(ctx, run.ID, AgentRunStatusCompleted, answer, nil, "")
		}
		if len(decision.ToolCalls) == 0 {
			return r.finish(ctx, run.ID, AgentRunStatusFailed, "", nil, "empty_model_decision")
		}

		for _, call := range decision.ToolCalls {
			result, err := r.registry.Call(ctx, call)
			output := mustJSON(result)
			status := AgentStepStatusCompleted
			errMsg := ""
			if err != nil {
				status = AgentStepStatusFailed
				errMsg = err.Error()
			}
			if _, stepErr := r.store.CreateAgentStep(ctx, CreateAgentStepInput{
				RunID:        run.ID,
				StepIndex:    stepIndex,
				StepType:     AgentStepTypeToolCall,
				ToolName:     call.ToolName,
				ToolInput:    call.Input,
				ToolOutput:   output,
				Status:       status,
				ErrorMessage: errMsg,
			}); stepErr != nil {
				return nil, stepErr
			}
			stepIndex++
			if err != nil {
				return r.finish(ctx, run.ID, AgentRunStatusFailed, "", nil, errMsg)
			}
			if result.RequiresConfirmation || result.ProposedAction != nil {
				if result.ProposedAction == nil {
					return r.finish(ctx, run.ID, AgentRunStatusFailed, "", nil, "missing_action_proposal")
				}
				pending := []ActionProposal{*result.ProposedAction}
				return r.finish(ctx, run.ID, AgentRunStatusRequiresConfirmation, "", mustJSON(pending), "")
			}
			observations = append(observations, ToolObservation{ToolName: call.ToolName, Output: result.Output})
		}
	}

	return r.finish(ctx, run.ID, AgentRunStatusFailed, "", nil, ErrMaxStepsExceeded.Error())
}

func (r *Runner) Get(ctx context.Context, id int64) (*AgentRunResponse, error) {
	run, err := r.store.GetAgentRun(ctx, id)
	if err != nil {
		return nil, err
	}
	steps, err := r.store.ListAgentSteps(ctx, id)
	if err != nil {
		return nil, err
	}
	return &AgentRunResponse{Run: *run, Steps: steps}, nil
}

func (r *Runner) fail(ctx context.Context, runID int64, stepIndex int, stepType AgentStepType, toolName string, input, output json.RawMessage, thought, message string) (*AgentRunResponse, error) {
	_, _ = r.store.CreateAgentStep(ctx, CreateAgentStepInput{
		RunID:          runID,
		StepIndex:      stepIndex,
		StepType:       stepType,
		ToolName:       toolName,
		ToolInput:      input,
		ToolOutput:     output,
		ThoughtSummary: thoughtSummary(thought),
		Status:         AgentStepStatusFailed,
		ErrorMessage:   message,
	})
	return r.finish(ctx, runID, AgentRunStatusFailed, "", nil, message)
}

func (r *Runner) finish(ctx context.Context, runID int64, status AgentRunStatus, finalAnswer string, pendingActions json.RawMessage, errorMessage string) (*AgentRunResponse, error) {
	if err := r.store.UpdateAgentRun(ctx, runID, UpdateAgentRunInput{
		Status:         status,
		FinalAnswer:    finalAnswer,
		PendingActions: pendingActions,
		ErrorMessage:   errorMessage,
		Complete:       true,
	}); err != nil {
		return nil, err
	}
	return r.Get(ctx, runID)
}

func normalizeRunRequest(req AgentRunRequest) (ContextPreviewRequest, error) {
	goal := strings.TrimSpace(req.Goal)
	if goal == "" {
		return ContextPreviewRequest{}, ErrInvalidContextPreviewInput
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(req.TargetDate)); err != nil {
		return ContextPreviewRequest{}, ErrInvalidContextPreviewInput
	}
	if req.RecentDays < 0 {
		return ContextPreviewRequest{}, ErrInvalidContextPreviewInput
	}
	return ContextPreviewRequest{Goal: goal, TargetDate: strings.TrimSpace(req.TargetDate), RecentDays: req.RecentDays}, nil
}

func thoughtSummary(value string) string {
	value, _ = excerpt(value, 500)
	return value
}

func mustJSON(value any) json.RawMessage {
	data, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`null`)
	}
	return data
}
