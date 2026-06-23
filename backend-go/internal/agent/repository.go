package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

var (
	ErrAgentRunNotFound        = errors.New("agent run not found")
	ErrProposalNotFound        = errors.New("agent action proposal not found")
	ErrProposalConflict        = errors.New("agent action proposal status conflict")
	ErrContextSnapshotNotFound = errors.New("agent context snapshot not found")
)

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

type CreateActionProposalInput struct {
	RunID      int64
	StepID     *int64
	ToolName   string
	ActionType string
	Payload    json.RawMessage
	RiskLevel  ToolRiskLevel
	Status     ActionProposalStatus
}

type UpdateActionProposalInput struct {
	Status           ActionProposalStatus
	Result           json.RawMessage
	ErrorMessage     string
	Decide           bool
	Execute          bool
	TargetEntityType string
	TargetEntityID   *int64
}

type ActionProposalFilter struct {
	RunID  int64
	Status string
}

type CreateContextSnapshotInput struct {
	RunID               int64
	ContextJSON         json.RawMessage
	TokenEstimate       int
	OmittedSectionsJSON json.RawMessage
}

type AgentRunFilter struct {
	Status string
	Limit  int
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

func (r *Repository) ListAgentRuns(ctx context.Context, filter AgentRunFilter) ([]AgentRunListItem, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	where := ""
	args := make([]any, 0, 2)
	if filter.Status != "" {
		where = "WHERE ar.status = ?"
		args = append(args, filter.Status)
	}
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, `
		SELECT ar.id, ar.user_goal, DATE_FORMAT(ar.target_date, '%Y-%m-%d'), ar.status,
			COALESCE(ar.final_answer, ''), ar.created_at, ar.completed_at,
			(SELECT COUNT(*) FROM agent_action_proposals p WHERE p.run_id = ar.id) AS proposal_count,
			(SELECT COUNT(*) FROM agent_action_proposals p WHERE p.run_id = ar.id AND p.status = 'pending') AS pending_proposal_count,
			(SELECT COUNT(*) FROM agent_steps s WHERE s.run_id = ar.id) AS step_count
		FROM agent_runs ar
		`+where+`
		ORDER BY ar.created_at DESC, ar.id DESC
		LIMIT ?
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AgentRunListItem, 0)
	for rows.Next() {
		var item AgentRunListItem
		var completedAt sql.NullTime
		var finalAnswer string
		if err := rows.Scan(&item.ID, &item.UserGoal, &item.TargetDate, &item.Status,
			&finalAnswer, &item.CreatedAt, &completedAt,
			&item.ProposalCount, &item.PendingProposalCount, &item.StepCount); err != nil {
			return nil, err
		}
		item.FinalAnswerExcerpt, _ = excerpt(finalAnswer, 160)
		if completedAt.Valid {
			item.CompletedAt = &completedAt.Time
		}
		items = append(items, item)
	}
	return items, rows.Err()
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

func (r *Repository) CreateActionProposal(ctx context.Context, input CreateActionProposalInput) (*ActionProposal, error) {
	status := input.Status
	if status == "" {
		status = ActionProposalStatusPending
	}
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO agent_action_proposals (
			run_id, step_id, tool_name, action_type, payload_json, risk_level, status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, input.RunID, nullableInt64(input.StepID), input.ToolName, input.ActionType, nullableRawJSON(input.Payload), input.RiskLevel, status)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetActionProposal(ctx, id)
}

func (r *Repository) GetActionProposal(ctx context.Context, id int64) (*ActionProposal, error) {
	rows, err := r.queryActionProposals(ctx, "WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrProposalNotFound
	}
	proposal, err := scanActionProposal(rows)
	if err != nil {
		return nil, err
	}
	return &proposal, rows.Err()
}

func (r *Repository) ListActionProposals(ctx context.Context, filter ActionProposalFilter) ([]ActionProposal, error) {
	where := ""
	args := make([]any, 0, 2)
	if filter.RunID > 0 {
		where = appendWhere(where, "run_id = ?")
		args = append(args, filter.RunID)
	}
	if filter.Status != "" {
		where = appendWhere(where, "status = ?")
		args = append(args, filter.Status)
	}
	rows, err := r.queryActionProposals(ctx, where+" ORDER BY created_at DESC, id DESC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	proposals := make([]ActionProposal, 0)
	for rows.Next() {
		proposal, err := scanActionProposal(rows)
		if err != nil {
			return nil, err
		}
		proposals = append(proposals, proposal)
	}
	return proposals, rows.Err()
}

func (r *Repository) UpdateActionProposal(ctx context.Context, id int64, input UpdateActionProposalInput) (*ActionProposal, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE agent_action_proposals
		SET status = ?,
			result_json = ?,
			error_message = ?,
			decided_at = CASE WHEN ? THEN COALESCE(decided_at, CURRENT_TIMESTAMP) ELSE decided_at END,
			executed_at = CASE WHEN ? THEN COALESCE(executed_at, CURRENT_TIMESTAMP) ELSE executed_at END,
			target_entity_type = ?,
			target_entity_id = ?
		WHERE id = ?
	`, input.Status, nullableRawJSON(input.Result), input.ErrorMessage, input.Decide, input.Execute,
		nullableString(input.TargetEntityType), nullableInt64(input.TargetEntityID), id)
	if err != nil {
		return nil, err
	}
	return r.GetActionProposal(ctx, id)
}

func (r *Repository) CreateContextSnapshot(ctx context.Context, input CreateContextSnapshotInput) (*AgentContextSnapshot, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO agent_context_snapshots (
			run_id, context_json, token_estimate, omitted_sections_json
		)
		VALUES (?, ?, ?, ?)
	`, input.RunID, nullableRawJSON(input.ContextJSON), input.TokenEstimate, nullableRawJSON(input.OmittedSectionsJSON))
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.getContextSnapshotByID(ctx, id)
}

func (r *Repository) GetContextSnapshot(ctx context.Context, runID int64) (*AgentContextSnapshot, error) {
	rows, err := r.queryContextSnapshots(ctx, "WHERE run_id = ? ORDER BY id DESC LIMIT 1", runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrContextSnapshotNotFound
	}
	snapshot, err := scanContextSnapshot(rows)
	if err != nil {
		return nil, err
	}
	return &snapshot, rows.Err()
}

func (r *Repository) getContextSnapshotByID(ctx context.Context, id int64) (*AgentContextSnapshot, error) {
	rows, err := r.queryContextSnapshots(ctx, "WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrContextSnapshotNotFound
	}
	snapshot, err := scanContextSnapshot(rows)
	if err != nil {
		return nil, err
	}
	return &snapshot, rows.Err()
}

func (r *Repository) queryActionProposals(ctx context.Context, suffix string, args ...any) (*sql.Rows, error) {
	query := `
		SELECT id, run_id, COALESCE(step_id, 0), tool_name, action_type, payload_json, risk_level,
			status, result_json, COALESCE(error_message, ''), created_at, decided_at, executed_at,
			COALESCE(target_entity_type, ''), target_entity_id
		FROM agent_action_proposals
	`
	if suffix != "" {
		query += " " + suffix
	}
	return r.db.QueryContext(ctx, query, args...)
}

func scanActionProposal(rows *sql.Rows) (ActionProposal, error) {
	var proposal ActionProposal
	var stepID int64
	var payload, result sql.NullString
	var decidedAt, executedAt sql.NullTime
	var targetID sql.NullInt64
	if err := rows.Scan(&proposal.ID, &proposal.RunID, &stepID, &proposal.ToolName, &proposal.ActionType,
		&payload, &proposal.RiskLevel, &proposal.Status, &result, &proposal.ErrorMessage,
		&proposal.CreatedAt, &decidedAt, &executedAt, &proposal.TargetEntityType, &targetID); err != nil {
		return ActionProposal{}, err
	}
	proposal.StepID = stepID
	if payload.Valid && json.Valid([]byte(payload.String)) {
		proposal.Payload = json.RawMessage(payload.String)
	}
	if result.Valid && json.Valid([]byte(result.String)) {
		proposal.Result = json.RawMessage(result.String)
	}
	if decidedAt.Valid {
		proposal.DecidedAt = &decidedAt.Time
	}
	if executedAt.Valid {
		proposal.ExecutedAt = &executedAt.Time
	}
	if targetID.Valid {
		proposal.TargetEntityID = &targetID.Int64
	}
	return proposal, nil
}

func (r *Repository) queryContextSnapshots(ctx context.Context, suffix string, args ...any) (*sql.Rows, error) {
	query := `
		SELECT id, run_id, context_json, COALESCE(token_estimate, 0), omitted_sections_json, created_at
		FROM agent_context_snapshots
	`
	if suffix != "" {
		query += " " + suffix
	}
	return r.db.QueryContext(ctx, query, args...)
}

func scanContextSnapshot(rows *sql.Rows) (AgentContextSnapshot, error) {
	var snapshot AgentContextSnapshot
	var contextJSON, omittedJSON sql.NullString
	if err := rows.Scan(&snapshot.ID, &snapshot.RunID, &contextJSON, &snapshot.TokenEstimate, &omittedJSON, &snapshot.CreatedAt); err != nil {
		return AgentContextSnapshot{}, err
	}
	if contextJSON.Valid && json.Valid([]byte(contextJSON.String)) {
		snapshot.ContextJSON = json.RawMessage(contextJSON.String)
		_ = json.Unmarshal(snapshot.ContextJSON, &snapshot.ContextPack)
	}
	if omittedJSON.Valid && json.Valid([]byte(omittedJSON.String)) {
		snapshot.OmittedSectionsJSON = json.RawMessage(omittedJSON.String)
		_ = json.Unmarshal(snapshot.OmittedSectionsJSON, &snapshot.OmittedSections)
	}
	return snapshot, nil
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

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func appendWhere(where, clause string) string {
	if where == "" {
		return "WHERE " + clause
	}
	return where + " AND " + clause
}
