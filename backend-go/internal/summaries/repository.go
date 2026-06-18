package summaries

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var ErrSummaryNotFound = errors.New("summary not found")

type Repository struct {
	db *sql.DB
}

type dailySummaryTaskRow struct {
	TaskID                int64
	Date                  string
	ProjectID             sql.NullInt64
	ProjectName           string
	ProjectCategory       string
	IncludeInSummary      bool
	Title                 string
	EstimatedMinutes      int
	Status                string
	FinishNote            string
	FinishDescription     string
	ActualSecondsOverride sql.NullInt64
	SessionSeconds        int
}

type dailySummarySessionRow struct {
	Date             string
	ProjectID        sql.NullInt64
	ProjectName      string
	ProjectCategory  string
	IncludeInSummary bool
	StartedAt        time.Time
	DurationSeconds  int
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSummary(ctx context.Context, input CreateSummaryInput) (int64, error) {
	query := `
		INSERT INTO generated_summaries (summary_type, start_date, end_date, content, source_data, action_items)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		input.SummaryType,
		input.StartDate,
		input.EndDate,
		input.Content,
		input.SourceData,
		input.ActionItems,
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

func (r *Repository) ListRecentDailyActiveDates(ctx context.Context, beforeDate string, limit int) ([]string, error) {
	query := `
		SELECT DATE_FORMAT(dt.task_date, '%Y-%m-%d') AS task_date
		FROM daily_tasks dt
		LEFT JOIN time_sessions ts ON ts.daily_task_id = dt.id
		LEFT JOIN projects p ON p.id = dt.project_id
		WHERE dt.task_date < ?
			AND p.id IS NOT NULL
			AND COALESCE(p.include_in_summary, TRUE)
			AND (dt.id > 0 OR ts.id IS NOT NULL OR dt.status = 'completed')
		GROUP BY dt.task_date
		ORDER BY dt.task_date DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, beforeDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make([]string, 0, limit)
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dates, nil
}

func (r *Repository) ListDailySummaryTasks(ctx context.Context, dates []string) ([]dailySummaryTaskRow, error) {
	if len(dates) == 0 {
		return []dailySummaryTaskRow{}, nil
	}

	query := `
		SELECT
			dt.id,
			DATE_FORMAT(dt.task_date, '%Y-%m-%d') AS task_date,
			dt.project_id,
			COALESCE(p.name, 'Unassigned') AS project_name,
			COALESCE(p.category, 'study') AS project_category,
			(p.id IS NOT NULL AND COALESCE(p.include_in_summary, TRUE)) AS include_in_summary,
			dt.title,
			dt.estimated_minutes,
			dt.status,
			COALESCE(dt.finish_note, '') AS finish_note,
			COALESCE(dt.finish_description, '') AS finish_description,
			dt.actual_seconds_override,
			COALESCE(ts.total_seconds, 0) AS session_seconds
		FROM daily_tasks dt
		LEFT JOIN projects p ON p.id = dt.project_id
		LEFT JOIN (
			SELECT daily_task_id, SUM(duration_seconds) AS total_seconds
			FROM time_sessions
			GROUP BY daily_task_id
		) ts ON ts.daily_task_id = dt.id
		WHERE dt.task_date IN (` + placeholders(len(dates)) + `)
		ORDER BY dt.task_date DESC, dt.id ASC
	`

	args := stringsToArgs(dates)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]dailySummaryTaskRow, 0)
	for rows.Next() {
		var task dailySummaryTaskRow
		if err := rows.Scan(
			&task.TaskID,
			&task.Date,
			&task.ProjectID,
			&task.ProjectName,
			&task.ProjectCategory,
			&task.IncludeInSummary,
			&task.Title,
			&task.EstimatedMinutes,
			&task.Status,
			&task.FinishNote,
			&task.FinishDescription,
			&task.ActualSecondsOverride,
			&task.SessionSeconds,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Repository) ListDailySummarySessions(ctx context.Context, dates []string) ([]dailySummarySessionRow, error) {
	if len(dates) == 0 {
		return []dailySummarySessionRow{}, nil
	}

	query := `
		SELECT
			DATE_FORMAT(dt.task_date, '%Y-%m-%d') AS task_date,
			dt.project_id,
			COALESCE(p.name, 'Unassigned') AS project_name,
			COALESCE(p.category, 'study') AS project_category,
			(p.id IS NOT NULL AND COALESCE(p.include_in_summary, TRUE)) AS include_in_summary,
			ts.started_at,
			ts.duration_seconds
		FROM time_sessions ts
		INNER JOIN daily_tasks dt ON dt.id = ts.daily_task_id
		LEFT JOIN projects p ON p.id = dt.project_id
		WHERE dt.task_date IN (` + placeholders(len(dates)) + `)
		ORDER BY dt.task_date DESC, ts.started_at ASC, ts.id ASC
	`

	args := stringsToArgs(dates)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]dailySummarySessionRow, 0)
	for rows.Next() {
		var session dailySummarySessionRow
		if err := rows.Scan(
			&session.Date,
			&session.ProjectID,
			&session.ProjectName,
			&session.ProjectCategory,
			&session.IncludeInSummary,
			&session.StartedAt,
			&session.DurationSeconds,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *Repository) ListSummaries(ctx context.Context, summaryType string) ([]GeneratedSummary, error) {
	query := `
		SELECT
			id,
			summary_type,
			DATE_FORMAT(start_date, '%Y-%m-%d') AS start_date,
			DATE_FORMAT(end_date, '%Y-%m-%d') AS end_date,
			content,
			action_items,
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
		var actionItems sql.NullString
		if err := rows.Scan(
			&summary.ID,
			&summary.SummaryType,
			&summary.StartDate,
			&summary.EndDate,
			&summary.Content,
			&actionItems,
			&summary.CreatedAt,
		); err != nil {
			return nil, err
		}
		if actionItems.Valid {
			summary.ActionItems = rawJSONOrNil(actionItems.String)
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
			action_items,
			created_at
		FROM generated_summaries
		WHERE id = ?
	`

	var summary GeneratedSummary
	var sourceData sql.NullString
	var actionItems sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&summary.ID,
		&summary.SummaryType,
		&summary.StartDate,
		&summary.EndDate,
		&summary.Content,
		&sourceData,
		&actionItems,
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
	if actionItems.Valid {
		summary.ActionItems = rawJSONOrNil(actionItems.String)
	}

	return &summary, nil
}

func rawJSONOrNil(value string) json.RawMessage {
	if !json.Valid([]byte(value)) {
		return nil
	}
	return json.RawMessage(value)
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

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func stringsToArgs(values []string) []any {
	args := make([]any, 0, len(values))
	for _, value := range values {
		args = append(args, value)
	}
	return args
}
