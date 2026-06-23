package agent

import (
	"encoding/json"
	"time"
)

type ContextPack struct {
	UserGoal          string           `json:"user_goal"`
	TargetDate        string           `json:"target_date"`
	TodayTasks        []ContextTask    `json:"today_tasks,omitempty"`
	RecentSummaries   []ContextSummary `json:"recent_summaries,omitempty"`
	Memories          []ContextMemory  `json:"memories,omitempty"`
	PlanRisk          json.RawMessage  `json:"plan_risk,omitempty"`
	RecentActionItems json.RawMessage  `json:"recent_action_items,omitempty"`
	Constraints       []string         `json:"constraints,omitempty"`
	OmittedSections   []string         `json:"omitted_sections,omitempty"`
}

type ContextTask struct {
	ID               int64  `json:"id"`
	ProjectID        int64  `json:"project_id"`
	ProjectName      string `json:"project_name,omitempty"`
	TaskDate         string `json:"task_date"`
	Title            string `json:"title"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	ActualMinutes    int    `json:"actual_minutes,omitempty"`
	Status           string `json:"status"`
}

type ContextSummary struct {
	ID          int64           `json:"id"`
	SummaryType string          `json:"summary_type"`
	StartDate   string          `json:"start_date"`
	EndDate     string          `json:"end_date"`
	Content     string          `json:"content"`
	ActionItems json.RawMessage `json:"action_items,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type ContextMemory struct {
	ID           int64           `json:"id"`
	MemoryType   string          `json:"memory_type"`
	ScopeType    string          `json:"scope_type"`
	ProjectID    *int64          `json:"project_id,omitempty"`
	Title        string          `json:"title"`
	Content      string          `json:"content"`
	Confidence   float64         `json:"confidence"`
	SupportCount int             `json:"support_count"`
	LastSeenAt   time.Time       `json:"last_seen_at"`
	Evidence     json.RawMessage `json:"evidence,omitempty"`
}
