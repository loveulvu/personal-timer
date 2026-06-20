package plans

type PlanRiskLevel string

const (
	PlanRiskInsufficientData PlanRiskLevel = "insufficient_data"
	PlanRiskLow              PlanRiskLevel = "low"
	PlanRiskMedium           PlanRiskLevel = "medium"
	PlanRiskHigh             PlanRiskLevel = "high"
)

type PlanRiskResponse struct {
	Date                   string        `json:"date"`
	PlannedTotalMinutes    int           `json:"planned_total_minutes"`
	RecentAvgActualMinutes int           `json:"recent_avg_actual_minutes"`
	RecentActiveDays       int           `json:"recent_active_days"`
	PlanRatio              float64       `json:"plan_ratio"`
	RiskLevel              PlanRiskLevel `json:"risk_level"`
	Reason                 string        `json:"reason"`
	Suggestions            []string      `json:"suggestions"`
}

type ActiveDayActualMinutes struct {
	Date          string
	ActualMinutes int
}
