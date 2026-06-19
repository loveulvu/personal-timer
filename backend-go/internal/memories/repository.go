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
)

var (
	ErrMemoryNotFound       = errors.New("memory not found")
	ErrInvalidMemoryInput   = errors.New("invalid memory input")
	ErrInvalidEvidenceInput = errors.New("invalid memory evidence input")
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

func (r *Repository) ListEvidence(ctx context.Context, memoryID int64) ([]StudyMemoryEvidence, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, memory_id, source_type, source_id, DATE_FORMAT(evidence_date, '%Y-%m-%d'), excerpt, weight, created_at
		FROM study_memory_evidence
		WHERE memory_id = ?
		ORDER BY evidence_date DESC, id DESC
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
