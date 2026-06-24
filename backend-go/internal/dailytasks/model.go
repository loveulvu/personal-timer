package dailytasks

import "time"

type DailyTask struct {
	ID                      int64      `json:"id"`
	ProjectID               *int64     `json:"project_id"`
	ProjectName             string     `json:"project_name,omitempty"`
	TaskDate                string     `json:"task_date"`
	Title                   string     `json:"title"`
	EstimatedMinutes        int        `json:"estimated_minutes"`
	Status                  string     `json:"status"`
	FinishNote              *string    `json:"finish_note"`
	FinishDescription       *string    `json:"finish_description"`
	CompletedAt             *time.Time `json:"completed_at"`
	ActualSecondsOverride   *int       `json:"actual_seconds_override"`
	ActualSeconds           int        `json:"actual_seconds"`
	CurrentSessionStartedAt *time.Time `json:"current_session_started_at"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
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

type UpdateDailyTaskRequest struct {
	ProjectID        *int64 `json:"project_id"`
	TaskDate         string `json:"task_date"`
	Title            string `json:"title"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	Status           string `json:"status"`
}

type UpdateDailyTaskInput struct {
	ProjectID        *int64
	TaskDate         string
	Title            string
	EstimatedMinutes int
	Status           string
}
