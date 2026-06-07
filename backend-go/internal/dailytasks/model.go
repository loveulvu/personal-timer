package dailytasks

import "time"

type DailyTask struct {
	ID               int64     `json:"id"`
	ProjectID        *int64    `json:"project_id"`
	TaskDate         string    `json:"task_date"`
	Title            string    `json:"title"`
	EstimatedMinutes int       `json:"estimated_minutes"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateDailyTaskRequest struct {
	ProjectID        *int64 `json:"project_id"`
	TaskDate         string `json:"task_date"`
	Title            string `json:"title"`
	EstimatedMinutes int    `json:"estimated_minutes"`
}

type CreateDailyTaskInput struct {
	ProjectID        *int64
	TaskDate         string
	Title            string
	EstimatedMinutes int
}
