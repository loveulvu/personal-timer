package summaries

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

var ErrSummaryNotFound = errors.New("summary not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSummary(ctx context.Context, input CreateSummaryInput) (int64, error) {
	query := `
		INSERT INTO generated_summaries (summary_type, start_date, end_date, content, source_data)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		input.SummaryType,
		input.StartDate,
		input.EndDate,
		input.Content,
		input.SourceData,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) ListSummaries(ctx context.Context, summaryType string) ([]GeneratedSummary, error) {
	query := `
		SELECT
			id,
			summary_type,
			DATE_FORMAT(start_date, '%Y-%m-%d') AS start_date,
			DATE_FORMAT(end_date, '%Y-%m-%d') AS end_date,
			content,
			created_at
		FROM generated_summaries
	`
	args := make([]any, 0, 1)
	if summaryType != "" {
		query += ` WHERE summary_type = ?`
		args = append(args, summaryType)
	}
	query += ` ORDER BY created_at DESC, id DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]GeneratedSummary, 0)
	for rows.Next() {
		var summary GeneratedSummary
		if err := rows.Scan(
			&summary.ID,
			&summary.SummaryType,
			&summary.StartDate,
			&summary.EndDate,
			&summary.Content,
			&summary.CreatedAt,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *Repository) GetSummaryByID(ctx context.Context, id int64) (*GeneratedSummary, error) {
	query := `
		SELECT
			id,
			summary_type,
			DATE_FORMAT(start_date, '%Y-%m-%d') AS start_date,
			DATE_FORMAT(end_date, '%Y-%m-%d') AS end_date,
			content,
			source_data,
			created_at
		FROM generated_summaries
		WHERE id = ?
	`

	var summary GeneratedSummary
	var sourceData sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&summary.ID,
		&summary.SummaryType,
		&summary.StartDate,
		&summary.EndDate,
		&summary.Content,
		&sourceData,
		&summary.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSummaryNotFound
	}
	if err != nil {
		return nil, err
	}
	if sourceData.Valid {
		summary.SourceData = json.RawMessage(sourceData.String)
	}

	return &summary, nil
}
