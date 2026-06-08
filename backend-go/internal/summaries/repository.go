package summaries

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/go-sql-driver/mysql"
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
		if isDuplicateKey(err) {
			return 0, ErrSummaryAlreadyExists
		}
		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) SummaryExists(ctx context.Context, summaryType, startDate, endDate string) (bool, error) {
	query := `
		SELECT 1
		FROM generated_summaries
		WHERE summary_type = ?
			AND start_date = ?
			AND end_date = ?
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRowContext(ctx, query, summaryType, startDate, endDate).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
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

func (r *Repository) DeleteSummary(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM generated_summaries WHERE id = ?`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrSummaryNotFound
	}

	return nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
