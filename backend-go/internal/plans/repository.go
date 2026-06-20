package plans

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetPlannedTotalMinutes(ctx context.Context, date string) (int, error) {
	query := `
		SELECT COALESCE(SUM(dt.estimated_minutes), 0)
		FROM daily_tasks dt
		INNER JOIN projects p ON p.id = dt.project_id
		WHERE dt.task_date = ?
			AND dt.estimated_minutes > 0
			AND dt.status IN ('planned', 'running', 'paused', 'completed')
			AND COALESCE(p.include_in_summary, TRUE)
	`
	var total int
	if err := r.db.QueryRowContext(ctx, query, date).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) ListRecentActiveDayActualMinutes(ctx context.Context, beforeDate string, limit int) ([]ActiveDayActualMinutes, error) {
	query := `
		SELECT task_date, FLOOR(SUM(actual_seconds) / 60) AS actual_minutes
		FROM (
			SELECT
				DATE_FORMAT(dt.task_date, '%Y-%m-%d') AS task_date,
				CASE
					WHEN dt.actual_seconds_override IS NOT NULL AND dt.actual_seconds_override > 0 THEN dt.actual_seconds_override
					ELSE COALESCE(ts.total_seconds, 0)
				END AS actual_seconds
			FROM daily_tasks dt
			INNER JOIN projects p ON p.id = dt.project_id
			LEFT JOIN (
				SELECT daily_task_id, SUM(duration_seconds) AS total_seconds
				FROM time_sessions
				GROUP BY daily_task_id
			) ts ON ts.daily_task_id = dt.id
			WHERE dt.task_date < ?
				AND COALESCE(p.include_in_summary, TRUE)
		) history
		GROUP BY task_date
		HAVING SUM(actual_seconds) > 0
		ORDER BY task_date DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, beforeDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	days := make([]ActiveDayActualMinutes, 0)
	for rows.Next() {
		var day ActiveDayActualMinutes
		if err := rows.Scan(&day.Date, &day.ActualMinutes); err != nil {
			return nil, err
		}
		days = append(days, day)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return days, nil
}
