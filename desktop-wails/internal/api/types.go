package api

import "time"

type VersionInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Mode    string `json:"mode"`
}

type ConfigStatus struct {
	Database             string `json:"database"`
	LLMConfigured        bool   `json:"llm_configured"`
	LLMBaseURLConfigured bool   `json:"llm_base_url_configured"`
	LLMModelConfigured   bool   `json:"llm_model_configured"`
	Error                string `json:"error,omitempty"`
}

type StartupStatus struct {
	Connected bool          `json:"connected"`
	Version   *VersionInfo  `json:"version,omitempty"`
	Config    *ConfigStatus `json:"config,omitempty"`
	Error     string        `json:"error,omitempty"`
}

type DailyTask struct {
	ID               int64  `json:"id"`
	ProjectID        *int64 `json:"project_id"`
	TaskDate         string `json:"task_date"`
	Title            string `json:"title"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	Status           string `json:"status"`
}

type CreateDailyTaskRequest struct {
	ProjectID        *int64 `json:"project_id"`
	TaskDate         string `json:"task_date"`
	Title            string `json:"title"`
	EstimatedMinutes int    `json:"estimated_minutes"`
}

type CreateResponse struct {
	ID int64 `json:"id"`
}

type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsFixed     bool      `json:"is_fixed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsFixed     bool   `json:"is_fixed"`
}

type dataResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}
