package stats

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetDailyTaskStats(date string) ([]DailyTaskStat, error) {
	query := `
		SELECT
    dt.id,
    dt.title,
    dt.status,
    dt.estimated_minutes,
    COALESCE(SUM(ts.duration_seconds), 0) AS actual_seconds
FROM daily_tasks dt
LEFT JOIN time_sessions ts ON ts.daily_task_id = dt.id
WHERE dt.task_date = ?
GROUP BY dt.id, dt.title, dt.status, dt.estimated_minutes
ORDER BY dt.id DESC;
	`
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ans := make([]DailyTaskStat, 0)
	for rows.Next() {
		var task DailyTaskStat

		err := rows.Scan(
			&task.TaskID,
			&task.Title,
			&task.Status,
			&task.EstimatedMinutes,
			&task.ActualSeconds,
		)
		if err != nil {
			return nil, err
		}
		task.ActualMinutes = task.ActualSeconds / 60
		ans = append(ans, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ans, nil
}
