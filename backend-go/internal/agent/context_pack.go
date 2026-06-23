package agent

import (
	"encoding/json"
	"time"
)

type ContextPack struct {
	UserGoal          string              `json:"user_goal"`
	TargetDate        string              `json:"target_date"`
	TodayTasks        []ContextTask       `json:"today_tasks"`
	RecentSummaries   []ContextSummary    `json:"recent_summaries"`
	Memories          []ContextMemory     `json:"memories"`
	PlanRisk          json.RawMessage     `json:"plan_risk,omitempty"`
	RecentActionItems []ContextActionItem `json:"recent_action_items"`
	Constraints       []string            `json:"constraints"`
	OmittedSections   []string            `json:"omitted_sections"`
}

type ContextTask struct {
	ID               int64  `json:"id"`
	ProjectID        *int64 `json:"project_id"`
	ProjectName      string `json:"project_name,omitempty"`
	TaskDate         string `json:"task_date"`
	Title            string `json:"title"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	ActualMinutes    int    `json:"actual_minutes"`
	Status           string `json:"status"`
}

type ContextSummary struct {
	ID                 int64     `json:"id"`
	SummaryType        string    `json:"summary_type"`
	StartDate          string    `json:"start_date"`
	EndDate            string    `json:"end_date"`
	ContentExcerpt     string    `json:"content_excerpt"`
	ActionItemsExcerpt string    `json:"action_items_excerpt,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type ContextMemory struct {
	ID                 int64     `json:"id"`
	MemoryType         string    `json:"memory_type"`
	ScopeType          string    `json:"scope_type"`
	ProjectID          *int64    `json:"project_id,omitempty"`
	Title              string    `json:"title"`
	Content            string    `json:"content"`
	Confidence         float64   `json:"confidence"`
	SupportCount       int       `json:"support_count"`
	ContradictionCount int       `json:"contradiction_count"`
	EvidenceCount      int       `json:"evidence_count"`
	EvidenceExcerpt    string    `json:"evidence_excerpt,omitempty"`
	Status             string    `json:"status"`
	LastSeenAt         time.Time `json:"last_seen_at"`
}

type ContextActionItem struct {
	SummaryID    int64  `json:"summary_id"`
	ItemIndex    int    `json:"item_index"`
	Content      string `json:"content"`
	Accepted     bool   `json:"accepted"`
	TargetDate   string `json:"target_date,omitempty"`
	TargetTaskID *int64 `json:"target_task_id,omitempty"`
}
