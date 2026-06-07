package dailytasks

import "database/sql"

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
