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
		SELECT id, project_id, task_date, title, estimated_minutes, status, created_at, updated_at
		FROM daily_tasks
		WHERE task_date = ?
		ORDER BY id DESC
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
			&task.TaskDate,
			&task.Title,
			&task.EstimatedMinutes,
			&task.Status,
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
		SELECT id, project_id, task_date, title, estimated_minutes, status, created_at, updated_at
		FROM daily_tasks
		WHERE id = ?
	`

	var task DailyTask
	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.ProjectID,
		&task.TaskDate,
		&task.Title,
		&task.EstimatedMinutes,
		&task.Status,
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
