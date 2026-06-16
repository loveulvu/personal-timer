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
	Excluded      SummaryExcluded     `json:"excluded"`
	Warnings      []string            `json:"warnings,omitempty"`
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
	FirstStartTime    string                  `json:"first_start_time"`
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

type WeeklySummarySourceData struct {
	SummaryType            string                 `json:"summary_type"`
	WeekStart              string                 `json:"week_start"`
	WeekEnd                string                 `json:"week_end"`
	DataQuality            WeeklyDataQuality      `json:"data_quality"`
	Week                   WeeklySummaryWeek      `json:"week"`
	PreviousWeekComparison PreviousWeekComparison `json:"previous_week_comparison"`
	Excluded               SummaryExcluded        `json:"excluded"`
	Warnings               []string               `json:"warnings,omitempty"`
}

type WeeklyDataQuality struct {
	DaysWithData    int  `json:"days_with_data"`
	CanAnalyzeTrend bool `json:"can_analyze_trend"`
	HasPreviousWeek bool `json:"has_previous_week"`
}

type WeeklySummaryWeek struct {
	TotalFocusMinutes int                      `json:"total_focus_minutes"`
	CompletedTasks    int                      `json:"completed_tasks"`
	TaskCount         int                      `json:"task_count"`
	DailyTotals       []WeeklyDailyTotal       `json:"daily_totals"`
	ProjectBreakdown  []WeeklyProjectBreakdown `json:"project_breakdown"`
	TimeDistribution  DailyTimeDistribution    `json:"time_distribution"`
	StartTimePatterns []WeeklyStartTimePattern `json:"start_time_patterns"`
	RepeatedNotes     []string                 `json:"repeated_notes"`
}

type WeeklyDailyTotal struct {
	Date              string `json:"date"`
	TotalFocusMinutes int    `json:"total_focus_minutes"`
	CompletedTasks    int    `json:"completed_tasks"`
	TaskCount         int    `json:"task_count"`
	FirstStartTime    string `json:"first_start_time"`
	MainProject       string `json:"main_project"`
}

type WeeklyProjectBreakdown struct {
	ProjectName      string  `json:"project_name"`
	TotalMinutes     int     `json:"total_minutes"`
	ActiveDays       int     `json:"active_days"`
	TaskCount        int     `json:"task_count"`
	CompletedCount   int     `json:"completed_count"`
	EstimatedMinutes int     `json:"estimated_minutes"`
	ActualMinutes    int     `json:"actual_minutes"`
	OverrunMinutes   int     `json:"overrun_minutes"`
	OverrunRate      float64 `json:"overrun_rate"`
}

type WeeklyStartTimePattern struct {
	ProjectName     string   `json:"project_name"`
	ActiveDays      int      `json:"active_days"`
	FirstStartTimes []string `json:"first_start_times"`
	AvgStartTime    string   `json:"avg_start_time"`
}

type PreviousWeekComparison struct {
	Available         bool                      `json:"available"`
	TotalFocusMinutes int                       `json:"total_focus_minutes"`
	DaysWithData      int                       `json:"days_with_data"`
	MainProjects      []PreviousWeekMainProject `json:"main_projects"`
}

type PreviousWeekMainProject struct {
	ProjectName  string `json:"project_name"`
	TotalMinutes int    `json:"total_minutes"`
}

type SummaryExcluded struct {
	ExcludedTaskCount      int                      `json:"excluded_task_count"`
	ExcludedTotalMinutes   int                      `json:"excluded_total_minutes"`
	ExcludedProjects       []SummaryExcludedProject `json:"excluded_projects"`
	UnassignedTaskCount    int                      `json:"unassigned_task_count"`
	UnassignedTotalMinutes int                      `json:"unassigned_total_minutes"`
}

type SummaryExcludedProject struct {
	ProjectName  string `json:"project_name"`
	Category     string `json:"category"`
	TotalMinutes int    `json:"total_minutes"`
}
