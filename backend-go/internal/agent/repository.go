package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

var ErrAgentRunNotFound = errors.New("agent run not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type CreateAgentRunInput struct {
	UserGoal   string
	TargetDate string
	Status     AgentRunStatus
}

type UpdateAgentRunInput struct {
	Status         AgentRunStatus
	FinalAnswer    string
	PendingActions json.RawMessage
	ErrorMessage   string
	Complete       bool
}

type CreateAgentStepInput struct {
	RunID          int64
	StepIndex      int
	StepType       AgentStepType
	ToolName       string
	ToolInput      json.RawMessage
	ToolOutput     json.RawMessage
	ThoughtSummary string
	Status         AgentStepStatus
	ErrorMessage   string
}

func (r *Repository) CreateAgentRun(ctx context.Context, input CreateAgentRunInput) (*AgentRun, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO agent_runs (user_goal, target_date, status)
		VALUES (?, ?, ?)
	`, input.UserGoal, input.TargetDate, input.Status)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetAgentRun(ctx, id)
}

func (r *Repository) UpdateAgentRun(ctx context.Context, id int64, input UpdateAgentRunInput) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE agent_runs
		SET status = ?,
			final_answer = ?,
			pending_actions_json = ?,
			error_message = ?,
			completed_at = CASE WHEN ? THEN CURRENT_TIMESTAMP ELSE completed_at END
		WHERE id = ?
	`, input.Status, input.FinalAnswer, nullableRawJSON(input.PendingActions), input.ErrorMessage, input.Complete, id)
	return err
}

func (r *Repository) CreateAgentStep(ctx context.Context, input CreateAgentStepInput) (*AgentStep, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO agent_steps (
			run_id, step_index, step_type, tool_name, tool_input_json, tool_output_json,
			thought_summary, status, error_message
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.RunID, input.StepIndex, input.StepType, nullableString(input.ToolName),
		nullableRawJSON(input.ToolInput), nullableRawJSON(input.ToolOutput),
		input.ThoughtSummary, input.Status, input.ErrorMessage)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	steps, err := r.ListAgentSteps(ctx, input.RunID)
	if err != nil {
		return nil, err
	}
	for i := range steps {
		if steps[i].ID == id {
			return &steps[i], nil
		}
	}
	return nil, ErrAgentRunNotFound
}

func (r *Repository) GetAgentRun(ctx context.Context, id int64) (*AgentRun, error) {
	var run AgentRun
	var pending sql.NullString
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_goal, DATE_FORMAT(target_date, '%Y-%m-%d'), status,
			COALESCE(final_answer, ''), pending_actions_json, COALESCE(error_message, ''),
			created_at, completed_at
		FROM agent_runs
		WHERE id = ?
	`, id).Scan(&run.ID, &run.UserGoal, &run.TargetDate, &run.Status, &run.FinalAnswer,
		&pending, &run.ErrorMessage, &run.CreatedAt, &completedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAgentRunNotFound
	}
	if err != nil {
		return nil, err
	}
	if pending.Valid && json.Valid([]byte(pending.String)) {
		run.PendingActions = json.RawMessage(pending.String)
	}
	if completedAt.Valid {
		run.CompletedAt = &completedAt.Time
	}
	return &run, nil
}

func (r *Repository) ListAgentSteps(ctx context.Context, runID int64) ([]AgentStep, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, run_id, step_index, step_type, COALESCE(tool_name, ''),
			tool_input_json, tool_output_json, COALESCE(thought_summary, ''),
			status, COALESCE(error_message, ''), created_at
		FROM agent_steps
		WHERE run_id = ?
		ORDER BY step_index ASC, id ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	steps := make([]AgentStep, 0)
	for rows.Next() {
		var step AgentStep
		var input, output sql.NullString
		if err := rows.Scan(&step.ID, &step.RunID, &step.StepIndex, &step.StepType, &step.ToolName,
			&input, &output, &step.ThoughtSummary, &step.Status, &step.ErrorMessage, &step.CreatedAt); err != nil {
			return nil, err
		}
		if input.Valid && json.Valid([]byte(input.String)) {
			step.ToolInput = json.RawMessage(input.String)
		}
		if output.Valid && json.Valid([]byte(output.String)) {
			step.ToolOutput = json.RawMessage(output.String)
		}
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

func nullableRawJSON(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
