package eval

import (
	"context"
	"errors"
	"fmt"
	"personal/internal/agent"
)

var ErrInvalidReplayRunID = errors.New("invalid replay run id")

type TrajectoryReader interface {
	GetTrajectory(ctx context.Context, id int64) (*agent.AgentTrajectory, error)
}

type ReplaySummary struct {
	RunID                  int64  `json:"run_id"`
	Status                 string `json:"status"`
	ContextSnapshotPresent bool   `json:"context_snapshot_present"`
	StepCount              int    `json:"step_count"`
	ProposalCount          int    `json:"proposal_count"`
	PendingProposalCount   int    `json:"pending_proposal_count"`
	Summary                string `json:"summary"`
}

func ReplayWithoutLLM(ctx context.Context, reader TrajectoryReader, runID int64) (*ReplaySummary, error) {
	if runID <= 0 {
		return nil, ErrInvalidReplayRunID
	}
	trajectory, err := reader.GetTrajectory(ctx, runID)
	if err != nil {
		return nil, err
	}
	pending := 0
	for _, proposal := range trajectory.Proposals {
		if proposal.Status == agent.ActionProposalStatusPending {
			pending++
		}
	}
	return &ReplaySummary{
		RunID:                  trajectory.Run.ID,
		Status:                 string(trajectory.Run.Status),
		ContextSnapshotPresent: trajectory.ContextSnapshot != nil,
		StepCount:              len(trajectory.Steps),
		ProposalCount:          len(trajectory.Proposals),
		PendingProposalCount:   pending,
		Summary:                replayText(trajectory, pending),
	}, nil
}

func replayText(t *agent.AgentTrajectory, pending int) string {
	if len(t.Proposals) == 0 {
		return fmt.Sprintf("Run %d built context and completed without action proposals.", t.Run.ID)
	}
	action := t.Proposals[0].ActionType
	if pending > 0 {
		return fmt.Sprintf("Run %d built context and created %s proposal, waiting for confirmation.", t.Run.ID, action)
	}
	return fmt.Sprintf("Run %d built context and recorded %d proposal(s).", t.Run.ID, len(t.Proposals))
}
