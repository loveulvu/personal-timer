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
	ID                      int64      `json:"id"`
	ProjectID               *int64     `json:"project_id"`
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
}

type CreateDailyTaskRequest struct {
	ProjectID        *int64 `json:"project_id"`
	TaskDate         string `json:"task_date"`
	Title            string `json:"title"`
	EstimatedMinutes int    `json:"estimated_minutes"`
}

type EstimatePreviewRequest struct {
	ProjectID        int64  `json:"project_id"`
	Title            string `json:"title"`
	EstimatedMinutes int    `json:"estimated_minutes"`
}

type EstimatePreviewResponse struct {
	ProjectID             int64   `json:"project_id"`
	InputEstimatedMinutes int     `json:"input_estimated_minutes"`
	SampleCount           int     `json:"sample_count"`
	AvgEstimatedMinutes   int     `json:"avg_estimated_minutes"`
	AvgActualMinutes      int     `json:"avg_actual_minutes"`
	OverrunRate           float64 `json:"overrun_rate"`
	RiskLevel             string  `json:"risk_level"`
	SuggestedMinutes      int     `json:"suggested_minutes"`
	SplitRecommended      bool    `json:"split_recommended"`
	Reason                string  `json:"reason"`
}

type PlanRiskResponse struct {
	Date                   string   `json:"date"`
	PlannedTotalMinutes    int      `json:"planned_total_minutes"`
	RecentAvgActualMinutes int      `json:"recent_avg_actual_minutes"`
	RecentActiveDays       int      `json:"recent_active_days"`
	PlanRatio              float64  `json:"plan_ratio"`
	RiskLevel              string   `json:"risk_level"`
	Reason                 string   `json:"reason"`
	Suggestions            []string `json:"suggestions"`
}

type FeedbackRequest struct {
	TargetType    string `json:"target_type"`
	TargetID      int64  `json:"target_id"`
	TargetIndex   *int   `json:"target_index"`
	FeedbackValue string `json:"feedback_value"`
	FeedbackNote  string `json:"feedback_note"`
}

type FeedbackResponse struct {
	ID            int64  `json:"id"`
	TargetType    string `json:"target_type"`
	TargetID      int64  `json:"target_id"`
	TargetIndex   *int   `json:"target_index"`
	FeedbackValue string `json:"feedback_value"`
	FeedbackNote  string `json:"feedback_note"`
	CreatedAt     string `json:"created_at"`
}

type MemoryListItem struct {
	ID                 int64   `json:"id"`
	MemoryType         string  `json:"memory_type"`
	ScopeType          string  `json:"scope_type"`
	ProjectID          *int64  `json:"project_id"`
	ProjectName        *string `json:"project_name"`
	Title              string  `json:"title"`
	Content            string  `json:"content"`
	Confidence         float64 `json:"confidence"`
	SupportCount       int     `json:"support_count"`
	ContradictionCount int     `json:"contradiction_count"`
	Status             string  `json:"status"`
	FirstSeenAt        string  `json:"first_seen_at"`
	LastSeenAt         string  `json:"last_seen_at"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type MemoryEvidenceItem struct {
	ID           int64   `json:"id"`
	MemoryID     int64   `json:"memory_id"`
	SourceType   string  `json:"source_type"`
	SourceID     *int64  `json:"source_id"`
	EvidenceDate string  `json:"evidence_date"`
	Excerpt      *string `json:"excerpt"`
	Weight       float64 `json:"weight"`
	CreatedAt    string  `json:"created_at"`
}

type CreateResponse struct {
	ID int64 `json:"id"`
}

type Project struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	IsFixed          bool      `json:"is_fixed"`
	Category         string    `json:"category"`
	IncludeInSummary bool      `json:"include_in_summary"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ProjectInput struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	IsFixed          bool   `json:"is_fixed"`
	Category         string `json:"category"`
	IncludeInSummary bool   `json:"include_in_summary"`
}

type DailyTaskStat struct {
	TaskID           int64  `json:"task_id"`
	Title            string `json:"title"`
	Status           string `json:"status"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	ActualSeconds    int    `json:"actual_seconds"`
	ActualMinutes    int    `json:"actual_minutes"`
}

type DailyStats struct {
	Date            string          `json:"date"`
	TotalSeconds    int             `json:"total_seconds"`
	TotalMinutes    int             `json:"total_minutes"`
	CompletedCount  int             `json:"completed_count"`
	UnfinishedCount int             `json:"unfinished_count"`
	Tasks           []DailyTaskStat `json:"tasks"`
}

type WeeklyDayStat struct {
	Date            string `json:"date"`
	TotalSeconds    int    `json:"total_seconds"`
	TotalMinutes    int    `json:"total_minutes"`
	CompletedCount  int    `json:"completed_count"`
	UnfinishedCount int    `json:"unfinished_count"`
}

type WeeklyProjectStat struct {
	ProjectID      int64  `json:"project_id"`
	ProjectName    string `json:"project_name"`
	TaskCount      int    `json:"task_count"`
	CompletedCount int    `json:"completed_count"`
	TotalSeconds   int    `json:"total_seconds"`
	TotalMinutes   int    `json:"total_minutes"`
}

type WeeklyStats struct {
	StartDate       string              `json:"start_date"`
	EndDate         string              `json:"end_date"`
	TotalSeconds    int                 `json:"total_seconds"`
	TotalMinutes    int                 `json:"total_minutes"`
	CompletedCount  int                 `json:"completed_count"`
	UnfinishedCount int                 `json:"unfinished_count"`
	Days            []WeeklyDayStat     `json:"days"`
	Projects        []WeeklyProjectStat `json:"projects"`
}

type GenerateDailySummaryRequest struct {
	Date string `json:"date"`
}

type GenerateWeeklySummaryRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type AcceptActionItemRequest struct {
	TargetDate string `json:"target_date"`
}

type FinishTaskRequest struct {
	FinishNote        string `json:"finish_note"`
	FinishDescription string `json:"finish_description"`
}

type UpdateCompletedTaskRequest struct {
	FinishNote            string `json:"finish_note"`
	FinishDescription     string `json:"finish_description"`
	ActualMinutesOverride *int   `json:"actual_minutes_override"`
	ClearActualOverride   *bool  `json:"clear_actual_minutes_override,omitempty"`
}

type GenerateSummaryResult struct {
	SummaryID   int64  `json:"summary_id"`
	Content     string `json:"content"`
	ActionItems any    `json:"action_items,omitempty"`
}

type AcceptActionItemResult struct {
	Created       bool       `json:"created"`
	AlreadyExists bool       `json:"already_exists"`
	Task          *DailyTask `json:"task,omitempty"`
	Message       string     `json:"message,omitempty"`
}

type Summary struct {
	ID          int64  `json:"id"`
	SummaryType string `json:"summary_type"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Content     string `json:"content"`
	SourceData  any    `json:"source_data,omitempty"`
	ActionItems any    `json:"action_items,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type LLMTestResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type dataResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}
