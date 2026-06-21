package memories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	defaultListLimit = 50
	maxListLimit     = 200
	maxUIListLimit   = 100
)

var (
	ErrMemoryNotFound       = errors.New("memory not found")
	ErrInvalidMemoryInput   = errors.New("invalid memory input")
	ErrInvalidEvidenceInput = errors.New("invalid memory evidence input")
	ErrSummaryNotFound      = errors.New("summary not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateMemory(ctx context.Context, input CreateMemoryInput) (StudyMemory, error) {
	if err := validateCreateMemory(input); err != nil {
		return StudyMemory{}, err
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if input.FirstSeenAt.IsZero() {
		input.FirstSeenAt = time.Now()
	}
	if input.LastSeenAt.IsZero() {
		input.LastSeenAt = input.FirstSeenAt
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO study_memories (
			memory_type, scope_type, project_id, title, content, structured_data,
			confidence, support_count, contradiction_count, first_seen_at, last_seen_at, status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.MemoryType, input.ScopeType, nullableInt64(input.ProjectID), input.Title, input.Content, nullableJSON(input.StructuredData),
		input.Confidence, input.SupportCount, input.ContradictionCount, input.FirstSeenAt, input.LastSeenAt, input.Status)
	if err != nil {
		return StudyMemory{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return StudyMemory{}, err
	}
	return r.GetMemoryByID(ctx, id)
}

func (r *Repository) GetMemoryByID(ctx context.Context, id int64) (StudyMemory, error) {
	return r.scanMemory(ctx, `
		SELECT id, memory_type, scope_type, project_id, title, content, structured_data,
			confidence, support_count, contradiction_count, first_seen_at, last_seen_at,
			status, created_at, updated_at
		FROM study_memories
		WHERE id = ?
	`, id)
}

func (r *Repository) FindActiveMemoryByIdentity(ctx context.Context, memoryType, scopeType string, projectID *int64, title string) (StudyMemory, error) {
	query := `
		SELECT id, memory_type, scope_type, project_id, title, content, structured_data,
			confidence, support_count, contradiction_count, first_seen_at, last_seen_at,
			status, created_at, updated_at
		FROM study_memories
		WHERE memory_type = ? AND scope_type = ? AND title = ? AND status = 'active'
	`
	args := []any{memoryType, scopeType, title}
	if projectID == nil {
		query += " AND project_id IS NULL"
	} else {
		query += " AND project_id = ?"
		args = append(args, *projectID)
	}
	query += " LIMIT 1"
	return r.scanMemory(ctx, query, args...)
}

func (r *Repository) ListMemories(ctx context.Context, filter ListMemoriesFilter) ([]StudyMemory, error) {
	args := make([]any, 0)
	conditions := []string{"status = ?"}
	status := strings.TrimSpace(filter.Status)
	if status == "" {
		status = "active"
	}
	if !validMemoryStatus(status) {
		return nil, ErrInvalidMemoryInput
	}
	args = append(args, status)

	if filter.MemoryType != "" {
		if !validMemoryType(filter.MemoryType) {
			return nil, ErrInvalidMemoryInput
		}
		conditions = append(conditions, "memory_type = ?")
		args = append(args, filter.MemoryType)
	}
	if filter.ScopeType != "" {
		if !validScopeType(filter.ScopeType) {
			return nil, ErrInvalidMemoryInput
		}
		conditions = append(conditions, "scope_type = ?")
		args = append(args, filter.ScopeType)
	}
	if filter.ProjectID != nil {
		conditions = append(conditions, "project_id = ?")
		args = append(args, *filter.ProjectID)
	}

	limit := normalizeLimit(filter.Limit)
	args = append(args, limit)
	query := `
		SELECT id, memory_type, scope_type, project_id, title, content, structured_data,
			confidence, support_count, contradiction_count, first_seen_at, last_seen_at,
			status, created_at, updated_at
		FROM study_memories
		WHERE ` + strings.Join(conditions, " AND ") + `
		ORDER BY last_seen_at DESC, id DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memories := make([]StudyMemory, 0)
	for rows.Next() {
		memory, err := scanMemoryRows(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return memories, nil
}

func (r *Repository) ListMemoriesForUI(ctx context.Context, filter ListMemoryItemsFilter) ([]MemoryListItem, error) {
	filter, allStatuses, err := normalizeMemoryItemsFilter(filter)
	if err != nil {
		return nil, err
	}

	conditions := make([]string, 0)
	args := make([]any, 0)
	if !allStatuses {
		conditions = append(conditions, "m.status = ?")
		args = append(args, filter.Status)
	}
	if filter.MemoryType != "" {
		conditions = append(conditions, "m.memory_type = ?")
		args = append(args, filter.MemoryType)
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	args = append(args, filter.Limit)

	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.memory_type, m.scope_type, m.project_id, p.name AS project_name,
			m.title, m.content, m.confidence, m.support_count, m.contradiction_count,
			m.status, m.first_seen_at, m.last_seen_at, m.created_at, m.updated_at
		FROM study_memories m
		LEFT JOIN projects p ON p.id = m.project_id
		`+where+`
		ORDER BY m.last_seen_at DESC, m.id DESC
		LIMIT ?
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]MemoryListItem, 0)
	for rows.Next() {
		var item MemoryListItem
		var projectID sql.NullInt64
		var projectName sql.NullString
		if err := rows.Scan(
			&item.ID, &item.MemoryType, &item.ScopeType, &projectID, &projectName,
			&item.Title, &item.Content, &item.Confidence, &item.SupportCount, &item.ContradictionCount,
			&item.Status, &item.FirstSeenAt, &item.LastSeenAt, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if projectID.Valid {
			item.ProjectID = &projectID.Int64
		}
		if projectName.Valid {
			item.ProjectName = &projectName.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) ListActiveMemoriesForRecall(ctx context.Context, projectIDs []int64, limit int) ([]StudyMemory, error) {
	args := []any{"repeated_blocker", "estimate_bias", "time_pattern", 0.5}
	query := `
		SELECT id, memory_type, scope_type, project_id, title, content, structured_data,
			confidence, support_count, contradiction_count, first_seen_at, last_seen_at,
			status, created_at, updated_at
		FROM study_memories
		WHERE status = 'active'
			AND memory_type IN (?, ?, ?)
			AND confidence >= ?
			AND (
				scope_type IN ('global', 'topic')
	`
	if len(projectIDs) > 0 {
		query += ` OR (scope_type = 'project' AND project_id IN (` + placeholders(len(projectIDs)) + `))`
		for _, id := range projectIDs {
			args = append(args, id)
		}
	}
	query += `
			)
		ORDER BY confidence DESC, last_seen_at DESC, id DESC
		LIMIT ?
	`
	args = append(args, normalizeLimit(limit))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memories := make([]StudyMemory, 0)
	for rows.Next() {
		memory, err := scanMemoryRows(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return memories, nil
}

func (r *Repository) UpdateMemory(ctx context.Context, id int64, input UpdateMemoryInput) (StudyMemory, error) {
	assignments := make([]string, 0)
	args := make([]any, 0)

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return StudyMemory{}, ErrInvalidMemoryInput
		}
		assignments = append(assignments, "title = ?")
		args = append(args, title)
	}
	if input.Content != nil {
		content := strings.TrimSpace(*input.Content)
		if content == "" {
			return StudyMemory{}, ErrInvalidMemoryInput
		}
		assignments = append(assignments, "content = ?")
		args = append(args, content)
	}
	if input.StructuredData != nil {
		if len(*input.StructuredData) > 0 && !json.Valid(*input.StructuredData) {
			return StudyMemory{}, ErrInvalidMemoryInput
		}
		assignments = append(assignments, "structured_data = ?")
		args = append(args, nullableJSON(*input.StructuredData))
	}
	if input.Confidence != nil {
		if !validConfidence(*input.Confidence) {
			return StudyMemory{}, ErrInvalidMemoryInput
		}
		assignments = append(assignments, "confidence = ?")
		args = append(args, *input.Confidence)
	}
	if input.SupportCount != nil {
		assignments = append(assignments, "support_count = ?")
		args = append(args, *input.SupportCount)
	}
	if input.ContradictionCount != nil {
		assignments = append(assignments, "contradiction_count = ?")
		args = append(args, *input.ContradictionCount)
	}
	if input.LastSeenAt != nil {
		assignments = append(assignments, "last_seen_at = ?")
		args = append(args, *input.LastSeenAt)
	}
	if input.Status != nil {
		if !validMemoryStatus(*input.Status) {
			return StudyMemory{}, ErrInvalidMemoryInput
		}
		assignments = append(assignments, "status = ?")
		args = append(args, *input.Status)
	}
	if len(assignments) == 0 {
		return r.GetMemoryByID(ctx, id)
	}

	assignments = append(assignments, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)
	result, err := r.db.ExecContext(ctx, "UPDATE study_memories SET "+strings.Join(assignments, ", ")+" WHERE id = ?", args...)
	if err != nil {
		return StudyMemory{}, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return StudyMemory{}, err
	}
	if rowsAffected == 0 {
		return StudyMemory{}, ErrMemoryNotFound
	}
	return r.GetMemoryByID(ctx, id)
}

func (r *Repository) ArchiveMemory(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `UPDATE study_memories SET status = 'archived', updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrMemoryNotFound
	}
	return nil
}

func (r *Repository) AddEvidence(ctx context.Context, input AddEvidenceInput) (StudyMemoryEvidence, error) {
	if err := validateEvidence(input); err != nil {
		return StudyMemoryEvidence{}, err
	}
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO study_memory_evidence (memory_id, source_type, source_id, evidence_date, excerpt, weight)
		VALUES (?, ?, ?, ?, ?, ?)
	`, input.MemoryID, input.SourceType, nullableInt64(input.SourceID), input.EvidenceDate, nullableString(input.Excerpt), input.Weight)
	if err != nil {
		return StudyMemoryEvidence{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return StudyMemoryEvidence{}, err
	}
	return r.getEvidenceByID(ctx, id)
}

func (r *Repository) EvidenceExists(ctx context.Context, memoryID int64, sourceType string, sourceID int64) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM study_memory_evidence
		WHERE memory_id = ? AND source_type = ? AND source_id = ?
		LIMIT 1
	`, memoryID, sourceType, sourceID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *Repository) ListEvidence(ctx context.Context, memoryID int64) ([]StudyMemoryEvidence, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, memory_id, source_type, source_id, DATE_FORMAT(evidence_date, '%Y-%m-%d'), excerpt, weight, created_at
		FROM study_memory_evidence
		WHERE memory_id = ?
		ORDER BY evidence_date DESC, created_at DESC, id DESC
	`, memoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StudyMemoryEvidence, 0)
	for rows.Next() {
		item, err := scanEvidenceRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) ListMemoryEvidence(ctx context.Context, memoryID int64) ([]StudyMemoryEvidence, error) {
	if memoryID <= 0 {
		return nil, ErrInvalidEvidenceInput
	}
	return r.ListEvidence(ctx, memoryID)
}

func (r *Repository) getEvidenceByID(ctx context.Context, id int64) (StudyMemoryEvidence, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, memory_id, source_type, source_id, DATE_FORMAT(evidence_date, '%Y-%m-%d'), excerpt, weight, created_at
		FROM study_memory_evidence
		WHERE id = ?
	`, id)
	return scanEvidenceRows(row)
}

func (r *Repository) scanMemory(ctx context.Context, query string, args ...any) (StudyMemory, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	memory, err := scanMemoryRows(row)
	if errors.Is(err, sql.ErrNoRows) {
		return StudyMemory{}, ErrMemoryNotFound
	}
	return memory, err
}

type memoryScanner interface {
	Scan(dest ...any) error
}

func scanMemoryRows(row memoryScanner) (StudyMemory, error) {
	var memory StudyMemory
	var projectID sql.NullInt64
	var structuredData sql.NullString
	err := row.Scan(
		&memory.ID,
		&memory.MemoryType,
		&memory.ScopeType,
		&projectID,
		&memory.Title,
		&memory.Content,
		&structuredData,
		&memory.Confidence,
		&memory.SupportCount,
		&memory.ContradictionCount,
		&memory.FirstSeenAt,
		&memory.LastSeenAt,
		&memory.Status,
		&memory.CreatedAt,
		&memory.UpdatedAt,
	)
	if err != nil {
		return StudyMemory{}, err
	}
	if projectID.Valid {
		id := projectID.Int64
		memory.ProjectID = &id
	}
	if structuredData.Valid && json.Valid([]byte(structuredData.String)) {
		memory.StructuredData = json.RawMessage(structuredData.String)
	}
	return memory, nil
}

type evidenceScanner interface {
	Scan(dest ...any) error
}

func scanEvidenceRows(row evidenceScanner) (StudyMemoryEvidence, error) {
	var item StudyMemoryEvidence
	var sourceID sql.NullInt64
	var excerpt sql.NullString
	err := row.Scan(&item.ID, &item.MemoryID, &item.SourceType, &sourceID, &item.EvidenceDate, &excerpt, &item.Weight, &item.CreatedAt)
	if err != nil {
		return StudyMemoryEvidence{}, err
	}
	if sourceID.Valid {
		id := sourceID.Int64
		item.SourceID = &id
	}
	if excerpt.Valid {
		value := excerpt.String
		item.Excerpt = &value
	}
	return item, nil
}

func validateCreateMemory(input CreateMemoryInput) error {
	if !validMemoryType(input.MemoryType) ||
		!validScopeType(input.ScopeType) ||
		strings.TrimSpace(input.Title) == "" ||
		strings.TrimSpace(input.Content) == "" ||
		!validConfidence(input.Confidence) ||
		(input.Status != "" && !validMemoryStatus(input.Status)) ||
		(len(input.StructuredData) > 0 && !json.Valid(input.StructuredData)) {
		return ErrInvalidMemoryInput
	}
	return nil
}

func validateEvidence(input AddEvidenceInput) error {
	if input.MemoryID <= 0 ||
		!validSourceType(input.SourceType) ||
		strings.TrimSpace(input.EvidenceDate) == "" ||
		input.Weight <= 0 ||
		input.Weight > 1 {
		return ErrInvalidEvidenceInput
	}
	if _, err := time.Parse("2006-01-02", input.EvidenceDate); err != nil {
		return ErrInvalidEvidenceInput
	}
	return nil
}

func validMemoryType(value string) bool {
	switch value {
	case "time_pattern", "estimate_bias", "project_pattern", "repeated_blocker", "suggestion_pattern":
		return true
	default:
		return false
	}
}

func validScopeType(value string) bool {
	switch value {
	case "global", "project", "topic":
		return true
	default:
		return false
	}
}

func validMemoryStatus(value string) bool {
	switch value {
	case "active", "archived":
		return true
	default:
		return false
	}
}

func validSourceType(value string) bool {
	switch value {
	case "daily_summary", "weekly_summary", "daily_task", "finish_note", "action_item", "manual":
		return true
	default:
		return false
	}
}

func validConfidence(value float64) bool {
	return value >= 0 && value <= 1
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func normalizeMemoryItemsFilter(filter ListMemoryItemsFilter) (ListMemoryItemsFilter, bool, error) {
	if filter.Status == "" {
		filter.Status = "active"
	}
	allStatuses := filter.Status == "all"
	if !allStatuses && !validMemoryStatus(filter.Status) {
		return filter, false, ErrInvalidMemoryInput
	}
	if filter.MemoryType != "" && !validMemoryType(filter.MemoryType) {
		return filter, false, ErrInvalidMemoryInput
	}
	if filter.Limit <= 0 {
		filter.Limit = defaultListLimit
	}
	if filter.Limit > maxUIListLimit {
		filter.Limit = maxUIListLimit
	}
	return filter, allStatuses, nil
}

func nullableJSON(value json.RawMessage) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

type summaryForExtraction struct {
	ID          int64
	SummaryType string
	StartDate   string
	EndDate     string
	SourceData  json.RawMessage
	ActionItems json.RawMessage
	CreatedAt   time.Time
}

type projectForExtraction struct {
	ID               int64
	Name             string
	IncludeInSummary bool
}

func (r *Repository) GetSummaryForExtraction(ctx context.Context, id int64) (summaryForExtraction, error) {
	var summary summaryForExtraction
	var sourceData sql.NullString
	var actionItems sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, summary_type, DATE_FORMAT(start_date, '%Y-%m-%d'), DATE_FORMAT(end_date, '%Y-%m-%d'), source_data, action_items, created_at
		FROM generated_summaries
		WHERE id = ?
	`, id).Scan(&summary.ID, &summary.SummaryType, &summary.StartDate, &summary.EndDate, &sourceData, &actionItems, &summary.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return summaryForExtraction{}, ErrSummaryNotFound
	}
	if err != nil {
		return summaryForExtraction{}, err
	}
	if sourceData.Valid {
		summary.SourceData = json.RawMessage(sourceData.String)
	}
	if actionItems.Valid && json.Valid([]byte(actionItems.String)) {
		summary.ActionItems = json.RawMessage(actionItems.String)
	}
	return summary, nil
}

func (r *Repository) FindProjectForExtraction(ctx context.Context, projectID *int64, name string) (*projectForExtraction, error) {
	var row *sql.Row
	if projectID != nil {
		row = r.db.QueryRowContext(ctx, `SELECT id, name, include_in_summary FROM projects WHERE id = ?`, *projectID)
	} else {
		row = r.db.QueryRowContext(ctx, `SELECT id, name, include_in_summary FROM projects WHERE name = ?`, name)
	}
	var project projectForExtraction
	err := row.Scan(&project.ID, &project.Name, &project.IncludeInSummary)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &project, nil
}
