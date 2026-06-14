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
    COALESCE(dt.actual_seconds_override, COALESCE(SUM(ts.duration_seconds), 0)) AS actual_seconds
FROM daily_tasks dt
LEFT JOIN time_sessions ts ON ts.daily_task_id = dt.id
WHERE dt.task_date = ?
GROUP BY dt.id, dt.title, dt.status, dt.estimated_minutes, dt.actual_seconds_override
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

func (r *Repository) GetWeeklyDayStats(startDate, endDate string) ([]WeeklyDayStat, error) {
	query := `
		SELECT
			DATE_FORMAT(dt.task_date, '%Y-%m-%d') AS task_date,
			COALESCE(SUM(COALESCE(dt.actual_seconds_override, ts.total_seconds, 0)), 0) AS total_seconds,
			SUM(dt.status = 'completed') AS completed_count,
			SUM(dt.status != 'completed' AND dt.status != 'cancelled') AS unfinished_count
		FROM daily_tasks dt
		LEFT JOIN (
			SELECT daily_task_id, SUM(duration_seconds) AS total_seconds
			FROM time_sessions
			GROUP BY daily_task_id
		) ts ON ts.daily_task_id = dt.id
		WHERE dt.task_date BETWEEN ? AND ?
		GROUP BY dt.task_date
		ORDER BY dt.task_date ASC
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	days := make([]WeeklyDayStat, 0)
	for rows.Next() {
		var day WeeklyDayStat
		if err := rows.Scan(
			&day.Date,
			&day.TotalSeconds,
			&day.CompletedCount,
			&day.UnfinishedCount,
		); err != nil {
			return nil, err
		}
		day.TotalMinutes = day.TotalSeconds / 60
		days = append(days, day)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return days, nil
}

func (r *Repository) GetWeeklyProjectStats(startDate, endDate string) ([]WeeklyProjectStat, error) {
	query := `
		SELECT
			COALESCE(dt.project_id, 0) AS project_id,
			COALESCE(p.name, 'Unassigned') AS project_name,
			COUNT(dt.id) AS task_count,
			SUM(dt.status = 'completed') AS completed_count,
			COALESCE(SUM(COALESCE(dt.actual_seconds_override, ts.total_seconds, 0)), 0) AS total_seconds
		FROM daily_tasks dt
		LEFT JOIN projects p ON p.id = dt.project_id
		LEFT JOIN (
			SELECT daily_task_id, SUM(duration_seconds) AS total_seconds
			FROM time_sessions
			GROUP BY daily_task_id
		) ts ON ts.daily_task_id = dt.id
		WHERE dt.task_date BETWEEN ? AND ?
		GROUP BY dt.project_id, p.name
		ORDER BY total_seconds DESC, project_id ASC
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]WeeklyProjectStat, 0)
	for rows.Next() {
		var project WeeklyProjectStat
		if err := rows.Scan(
			&project.ProjectID,
			&project.ProjectName,
			&project.TaskCount,
			&project.CompletedCount,
			&project.TotalSeconds,
		); err != nil {
			return nil, err
		}
		project.TotalMinutes = project.TotalSeconds / 60
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}
