package summaries

import (
	"encoding/json"
	"time"
)

type GeneratedSummary struct {
	ID          int64           `json:"id"`
	SummaryType string          `json:"summary_type"`
	StartDate   string          `json:"start_date"`
	EndDate     string          `json:"end_date"`
	Content     string          `json:"content"`
	SourceData  json.RawMessage `json:"source_data,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CreateSummaryInput struct {
	SummaryType string
	StartDate   string
	EndDate     string
	Content     string
	SourceData  json.RawMessage
}

type GenerateDailySummaryRequest struct {
	Date string `json:"date"`
}

type GenerateWeeklySummaryRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type GenerateSummaryResult struct {
	SummaryID int64  `json:"summary_id"`
	Content   string `json:"content"`
}

type DailySummarySourceData struct {
	SummaryType   string              `json:"summary_type"`
	TargetDate    string              `json:"target_date"`
	DataQuality   DailyDataQuality    `json:"data_quality"`
	Today         DailySummaryToday   `json:"today"`
	RecentContext DailySummaryContext `json:"recent_context"`
}

type DailyDataQuality struct {
	DaysWithData           int  `json:"days_with_data"`
	CanAnalyzeTrend        bool `json:"can_analyze_trend"`
	ComparisonWindowDays   int  `json:"comparison_window_days"`
	ComparisonDaysWithData int  `json:"comparison_days_with_data"`
}

type DailySummaryToday struct {
	TotalFocusMinutes int                     `json:"total_focus_minutes"`
	CompletedTasks    int                     `json:"completed_tasks"`
	TaskCount         int                     `json:"task_count"`
	ProjectBreakdown  []DailyProjectBreakdown `json:"project_breakdown"`
	TimeDistribution  DailyTimeDistribution   `json:"time_distribution"`
}

type DailyProjectBreakdown struct {
	ProjectName      string `json:"project_name"`
	TotalMinutes     int    `json:"total_minutes"`
	TaskCount        int    `json:"task_count"`
	CompletedCount   int    `json:"completed_count"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	ActualMinutes    int    `json:"actual_minutes"`
	OverrunMinutes   int    `json:"overrun_minutes"`
}

type DailyTimeDistribution struct {
	MorningMinutes   int `json:"morning_minutes"`
	AfternoonMinutes int `json:"afternoon_minutes"`
	EveningMinutes   int `json:"evening_minutes"`
	NightMinutes     int `json:"night_minutes"`
}

type DailySummaryContext struct {
	RecentActiveDays []DailyRecentActiveDay `json:"recent_active_days"`
	ProjectPatterns  []DailyProjectPattern  `json:"project_patterns"`
	RepeatedNotes    []string               `json:"repeated_notes"`
}

type DailyRecentActiveDay struct {
	Date              string `json:"date"`
	TotalFocusMinutes int    `json:"total_focus_minutes"`
	FirstStartTime    string `json:"first_start_time"`
	MainProject       string `json:"main_project"`
}

type DailyProjectPattern struct {
	ProjectName         string  `json:"project_name"`
	ActiveDays          int     `json:"active_days"`
	AvgStartTime        string  `json:"avg_start_time"`
	AvgActualMinutes    int     `json:"avg_actual_minutes"`
	AvgEstimatedMinutes int     `json:"avg_estimated_minutes"`
	OverrunRate         float64 `json:"overrun_rate"`
}
