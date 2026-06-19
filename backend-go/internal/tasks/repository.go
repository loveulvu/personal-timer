package tasks

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

func (r *Repository) ProjectExists(ctx context.Context, projectID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE id = ?)`, projectID).Scan(&exists)
	return exists, err
}

func (r *Repository) ListEstimateHistorySamples(ctx context.Context, projectID int64, limit int) ([]EstimateHistorySample, error) {
	query := `
		SELECT id, estimated_minutes, actual_seconds
		FROM (
			SELECT
				dt.id,
				dt.estimated_minutes,
				dt.completed_at,
				CASE
					WHEN dt.actual_seconds_override IS NOT NULL AND dt.actual_seconds_override > 0 THEN dt.actual_seconds_override
					ELSE COALESCE(ts.total_seconds, 0)
				END AS actual_seconds
			FROM daily_tasks dt
			LEFT JOIN (
				SELECT daily_task_id, SUM(duration_seconds) AS total_seconds
				FROM time_sessions
				GROUP BY daily_task_id
			) ts ON ts.daily_task_id = dt.id
			WHERE dt.project_id = ?
				AND dt.status = 'completed'
				AND dt.estimated_minutes > 0
		) history
		WHERE actual_seconds > 0
		ORDER BY completed_at DESC, id DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	samples := make([]EstimateHistorySample, 0)
	for rows.Next() {
		var sample EstimateHistorySample
		if err := rows.Scan(&sample.TaskID, &sample.EstimatedMinutes, &sample.ActualSeconds); err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return samples, nil
}
