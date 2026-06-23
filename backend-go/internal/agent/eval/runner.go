package eval

import (
	"context"
	"fmt"
	"personal/internal/agent"
)

type CaseExecutor interface {
	ExecuteEvalCase(ctx context.Context, c EvalCase) (EvalObservation, error)
}

type EvalObservation struct {
	RunID              int64
	RunStatus          agent.AgentRunStatus
	Steps              []agent.AgentStep
	Proposals          []agent.ActionProposal
	ContextSnapshot    *agent.AgentContextSnapshot
	BusinessWriteCount int
}

type Runner struct {
	executor CaseExecutor
}

func NewRunner(executor CaseExecutor) *Runner {
	return &Runner{executor: executor}
}

func (r *Runner) Run(ctx context.Context, cases []EvalCase) []EvalResult {
	results := make([]EvalResult, 0, len(cases))
	for _, c := range cases {
		results = append(results, r.RunOne(ctx, c))
	}
	return results
}

func (r *Runner) RunOne(ctx context.Context, c EvalCase) EvalResult {
	if r.executor == nil {
		return EvalResult{CaseName: c.Name, Reason: "missing_eval_executor"}
	}
	obs, err := r.executor.ExecuteEvalCase(ctx, c)
	if err != nil {
		return EvalResult{CaseName: c.Name, RunID: obs.RunID, Reason: err.Error()}
	}
	return evaluate(c, obs)
}

func evaluate(c EvalCase, obs EvalObservation) EvalResult {
	if c.Expectation.ExpectedRunStatus != "" && string(obs.RunStatus) != c.Expectation.ExpectedRunStatus {
		return fail(c, obs, fmt.Sprintf("status=%s want %s", obs.RunStatus, c.Expectation.ExpectedRunStatus))
	}
	if c.Expectation.MustUseTool != "" && !usedTool(obs.Steps, c.Expectation.MustUseTool) {
		return fail(c, obs, "missing required tool "+c.Expectation.MustUseTool)
	}
	if c.Expectation.MustNotUseTool != "" && usedTool(obs.Steps, c.Expectation.MustNotUseTool) {
		return fail(c, obs, "used forbidden tool "+c.Expectation.MustNotUseTool)
	}
	if c.Expectation.MustRequireConfirmation && obs.RunStatus != agent.AgentRunStatusRequiresConfirmation {
		return fail(c, obs, "run did not require confirmation")
	}
	if c.Expectation.MustCreateProposal && len(obs.Proposals) == 0 {
		return fail(c, obs, "missing action proposal")
	}
	if !c.Expectation.MustCreateProposal && len(obs.Proposals) > 0 {
		return fail(c, obs, "unexpected action proposal")
	}
	if c.Expectation.MustNotWriteBusinessDB && obs.BusinessWriteCount != 0 {
		return fail(c, obs, "business db was written")
	}
	if c.Expectation.ExpectedBusinessWrites != nil && obs.BusinessWriteCount != *c.Expectation.ExpectedBusinessWrites {
		return fail(c, obs, fmt.Sprintf("business_writes=%d want %d", obs.BusinessWriteCount, *c.Expectation.ExpectedBusinessWrites))
	}
	if c.Expectation.MustHaveContextSnapshot && obs.ContextSnapshot == nil {
		return fail(c, obs, "missing context snapshot")
	}
	if c.Expectation.MustHaveOmittedSections && (obs.ContextSnapshot == nil || len(obs.ContextSnapshot.OmittedSections) == 0) {
		return fail(c, obs, "missing omitted_sections")
	}
	if c.Expectation.MustHaveConstraints && (obs.ContextSnapshot == nil || len(obs.ContextSnapshot.ContextPack.Constraints) == 0) {
		return fail(c, obs, "missing constraints")
	}
	return EvalResult{CaseName: c.Name, Passed: true, Reason: "passed", RunID: obs.RunID}
}

func fail(c EvalCase, obs EvalObservation, reason string) EvalResult {
	return EvalResult{CaseName: c.Name, Passed: false, Reason: reason, RunID: obs.RunID}
}

func usedTool(steps []agent.AgentStep, name string) bool {
	for _, step := range steps {
		if step.ToolName == name {
			return true
		}
	}
	return false
}
