package dailytasks

import (
	"database/sql"
	"errors"
)

var ErrDailyTaskNotFound = errors.New("daily task not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateDailyTask(input CreateDailyTaskInput) (int64, error) {
	query := `
		INSERT INTO daily_tasks (project_id, task_date, title, estimated_minutes)
VALUES (?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, input.ProjectID, input.TaskDate, input.Title, input.EstimatedMinutes)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil

}

func (r *Repository) ListDailyTasksByDate(date string) ([]DailyTask, error) {
	query := `
		SELECT dt.id, dt.project_id, COALESCE(p.name, ''), dt.task_date, dt.title, dt.estimated_minutes, dt.status,
			dt.finish_note, dt.finish_description, dt.completed_at, dt.actual_seconds_override,
			CASE
				WHEN dt.status = 'completed' THEN COALESCE(dt.actual_seconds_override, COALESCE(ts.total_seconds, 0))
				ELSE COALESCE(ts.total_seconds, 0)
			END AS actual_seconds,
			ts.current_session_started_at,
			dt.created_at, dt.updated_at
		FROM daily_tasks dt
		LEFT JOIN projects p ON p.id = dt.project_id
		LEFT JOIN (
			SELECT daily_task_id,
				SUM(duration_seconds) AS total_seconds,
				MAX(CASE WHEN ended_at IS NULL THEN started_at END) AS current_session_started_at
			FROM time_sessions
			GROUP BY daily_task_id
		) ts ON ts.daily_task_id = dt.id
		WHERE dt.task_date = ?
		ORDER BY dt.id DESC
	`

	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]DailyTask, 0)
	for rows.Next() {
		var task DailyTask

		if err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.ProjectName,
			&task.TaskDate,
			&task.Title,
			&task.EstimatedMinutes,
			&task.Status,
			&task.FinishNote,
			&task.FinishDescription,
			&task.CompletedAt,
			&task.ActualSecondsOverride,
			&task.ActualSeconds,
			&task.CurrentSessionStartedAt,
			&task.CreatedAt,
			&task.UpdatedAt,
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

func (r *Repository) GetDailyTaskByID(id int64) (*DailyTask, error) {
	query := `
		SELECT dt.id, dt.project_id, COALESCE(p.name, ''), dt.task_date, dt.title, dt.estimated_minutes, dt.status,
			dt.finish_note, dt.finish_description, dt.completed_at, dt.actual_seconds_override,
			CASE
				WHEN dt.status = 'completed' THEN COALESCE(dt.actual_seconds_override, COALESCE(ts.total_seconds, 0))
				ELSE COALESCE(ts.total_seconds, 0)
			END AS actual_seconds,
			ts.current_session_started_at,
			dt.created_at, dt.updated_at
		FROM daily_tasks dt
		LEFT JOIN projects p ON p.id = dt.project_id
		LEFT JOIN (
			SELECT daily_task_id,
				SUM(duration_seconds) AS total_seconds,
				MAX(CASE WHEN ended_at IS NULL THEN started_at END) AS current_session_started_at
			FROM time_sessions
			GROUP BY daily_task_id
		) ts ON ts.daily_task_id = dt.id
		WHERE dt.id = ?
	`

	var task DailyTask
	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.ProjectID,
		&task.ProjectName,
		&task.TaskDate,
		&task.Title,
		&task.EstimatedMinutes,
		&task.Status,
		&task.FinishNote,
		&task.FinishDescription,
		&task.CompletedAt,
		&task.ActualSecondsOverride,
		&task.ActualSeconds,
		&task.CurrentSessionStartedAt,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDailyTaskNotFound
		}
		return nil, err
	}

	return &task, nil
}

func (r *Repository) UpdateDailyTask(id int64, input UpdateDailyTaskInput) error {
	query := `
		UPDATE daily_tasks
		SET project_id = ?, task_date = ?, title = ?, estimated_minutes = ?, status = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(
		query,
		input.ProjectID,
		input.TaskDate,
		input.Title,
		input.EstimatedMinutes,
		input.Status,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteDailyTask(id int64) error {
	query := `DELETE FROM daily_tasks WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrDailyTaskNotFound
	}

	return nil
}
